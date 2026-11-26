package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/lib/pq"
)

type branchColumn struct {
	schema, table, name, dataType string
	kind                          piiKind
}

func (m branchMasker) sql(ctx context.Context, conn *connection.StoredConnection, user, password string) error {
	db, err := openSQLTarget(conn)
	if err != nil {
		return err
	}
	defer db.Close()
	pg := conn.Type == "postgresql"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var schemas []string
	if pg {
		if _, err := tx.ExecContext(ctx, "SET LOCAL search_path = pg_catalog, pg_temp"); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT schema_name FROM information_schema.schemata
			WHERE schema_name NOT IN ('pg_catalog', 'information_schema') AND schema_name !~ '^pg_'`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var schema string
			if err := rows.Scan(&schema); err != nil {
				rows.Close()
				return err
			}
			schemas = append(schemas, schema)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		var materialized bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_matviews)").Scan(&materialized); err != nil {
			return err
		}
		if materialized {
			return fmt.Errorf("materialized views may contain unmasked values")
		}
	}
	columns, err := branchSQLColumns(ctx, tx, pg, conn.DatabaseName)
	if err != nil {
		return err
	}
	if len(columns) > 0 {
		create := "CREATE TEMPORARY TABLE velld_mask_values (digest CHAR(64) PRIMARY KEY, masked TEXT)"
		if pg {
			create += " ON COMMIT DROP"
		}
		if _, err := tx.ExecContext(ctx, create); err != nil {
			return err
		}
	}

	var lastTable string
	for _, col := range columns {
		table := quoteBranchIdentifier(pg, col.schema) + "." + quoteBranchIdentifier(pg, col.table)
		if table != lastTable {
			if pg {
				if lastTable != "" {
					if _, err := tx.ExecContext(ctx, "ALTER TABLE "+lastTable+" ENABLE TRIGGER USER"); err != nil {
						return err
					}
				}
				if _, err := tx.ExecContext(ctx, "ALTER TABLE "+table+" DISABLE TRIGGER USER"); err != nil {
					return err
				}
			} else {
				var triggers bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.triggers
					WHERE event_object_schema = ? AND event_object_table = ? AND event_manipulation = 'UPDATE')`,
					col.schema, col.table).Scan(&triggers); err != nil {
					return err
				}
				if triggers {
					return fmt.Errorf("update triggers may copy unmasked values")
				}
			}
			lastTable = table
		}
		if err := m.maskSQLColumn(ctx, tx, pg, table, col); err != nil {
			return err
		}
	}
	if pg && lastTable != "" {
		if _, err := tx.ExecContext(ctx, "ALTER TABLE "+lastTable+" ENABLE TRIGGER USER"); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return grantBranchSQLAccess(ctx, db, pg, schemas, conn.DatabaseName, user, password)
}

func branchSQLColumns(ctx context.Context, tx *sql.Tx, pg bool, database string) ([]branchColumn, error) {
	typeColumn := "c.data_type"
	if pg {
		typeColumn = "c.udt_name"
	}
	query := `SELECT c.table_schema, c.table_name, c.column_name, ` + typeColumn + `
		FROM information_schema.columns c
		JOIN information_schema.tables t ON (c.table_schema = t.table_schema AND c.table_name = t.table_name)
		WHERE t.table_type = 'BASE TABLE' AND `
	var rows *sql.Rows
	var err error
	if pg {
		rows, err = tx.QueryContext(ctx, query+`c.table_schema <> 'information_schema' AND c.table_schema !~ '^pg_'
			ORDER BY c.table_schema, c.table_name, c.ordinal_position`)
	} else {
		rows, err = tx.QueryContext(ctx, query+`c.table_schema = ? ORDER BY c.table_name, c.ordinal_position`, database)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []branchColumn
	for rows.Next() {
		var col branchColumn
		if err := rows.Scan(&col.schema, &col.table, &col.name, &col.dataType); err != nil {
			return nil, err
		}
		if col.kind = sensitiveField(col.name); col.kind != "" {
			columns = append(columns, col)
		}
	}
	return columns, rows.Err()
}

func (m branchMasker) maskSQLColumn(ctx context.Context, tx *sql.Tx, pg bool, table string, col branchColumn) error {
	field := quoteBranchIdentifier(pg, col.name)
	qualified := "branch_data." + field
	value := "CAST(" + qualified + " AS BINARY)"
	digest := "SHA2(" + value + ", 256)"
	// ONLY avoids processing child tables again via a PostgreSQL inheritance parent.
	if pg {
		table = "ONLY " + table
		value = "(" + qualified + `::text COLLATE "C")`
		digest = "encode(sha256(convert_to(" + value + ", 'UTF8')), 'hex')"
	}
	table += " AS branch_data"
	rows, err := tx.QueryContext(ctx, "SELECT "+value+", count(*) FROM "+table+" WHERE "+qualified+" IS NOT NULL GROUP BY "+value)
	if err != nil {
		return err
	}
	values := make(map[string]string)
	replacements := make(map[string]bool)
	var expected int64
	for rows.Next() {
		var original string
		var count int64
		if err := rows.Scan(&original, &count); err != nil {
			rows.Close()
			return err
		}
		if original != "" {
			masked := m.value(col.kind, original)
			if masked != original && replacements[masked] {
				rows.Close()
				return fmt.Errorf("different values produced the same mask")
			}
			if masked != original {
				values[original], replacements[masked] = masked, true
				expected += count
			}
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil || len(values) == 0 {
		return err
	}

	replacement := "branch_mask.masked"
	switch strings.ToLower(col.dataType) {
	case "text", "tinytext", "mediumtext", "longtext", "char", "varchar", "bpchar", "citext":
	case "smallint", "integer", "bigint", "tinyint", "mediumint", "int", "int2", "int4", "int8", "numeric", "decimal":
		if col.kind != piiPhone {
			return fmt.Errorf("non-text name or email column")
		}
		if pg {
			replacement += "::numeric"
		}
	default:
		return fmt.Errorf("unsupported sensitive column type")
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM velld_mask_values"); err != nil {
		return err
	}
	args := make([]any, 0, 1000)
	bindings := make([]string, 0, 500)
	flush := func() error {
		_, err := tx.ExecContext(ctx, "INSERT INTO velld_mask_values VALUES "+strings.Join(bindings, ","), args...)
		args, bindings = args[:0], bindings[:0]
		return err
	}
	for original, masked := range values {
		sum := sha256.Sum256([]byte(original))
		args = append(args, hex.EncodeToString(sum[:]), masked)
		if pg {
			bindings = append(bindings, fmt.Sprintf("($%d,$%d)", len(args)-1, len(args)))
		} else {
			bindings = append(bindings, "(?,?)")
		}
		if len(bindings) == cap(bindings) {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if len(bindings) != 0 {
		if err := flush(); err != nil {
			return err
		}
	}
	query := "UPDATE " + table + " JOIN velld_mask_values AS branch_mask ON branch_mask.digest = " + digest + " SET " + qualified + " = " + replacement
	if pg {
		query = "UPDATE " + table + " SET " + field + " = " + replacement + " FROM pg_temp.velld_mask_values AS branch_mask WHERE branch_mask.digest = " + digest
	}
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != expected {
		return fmt.Errorf("sensitive column did not update every matching row")
	}
	return nil
}

func quoteBranchIdentifier(pg bool, value string) string {
	if pg {
		return pq.QuoteIdentifier(value)
	}
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

func grantBranchSQLAccess(ctx context.Context, db *sql.DB, pg bool, schemas []string, database, user, password string) error {
	if pg {
		role := pq.QuoteIdentifier(user)
		if _, err := db.ExecContext(ctx, "CREATE ROLE "+role+" LOGIN PASSWORD "+pq.QuoteLiteral(password)); err != nil {
			return err
		}
		for _, schema := range schemas {
			name := pq.QuoteIdentifier(schema)
			for _, grant := range []string{
				"GRANT USAGE ON SCHEMA " + name,
				"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA " + name,
				"GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA " + name,
			} {
				if _, err := db.ExecContext(ctx, grant+" TO "+role); err != nil {
					return err
				}
			}
		}
		return nil
	}
	role := "'" + user + "'@'%'"
	if _, err := db.ExecContext(ctx, "CREATE USER "+role+" IDENTIFIED BY '"+password+"'"); err != nil {
		return err
	}
	database = strings.NewReplacer("%", "\\%", "_", "\\_").Replace(database)
	_, err := db.ExecContext(ctx, "GRANT SELECT, INSERT, UPDATE, DELETE ON "+quoteBranchIdentifier(false, database)+".* TO "+role)
	return err
}
