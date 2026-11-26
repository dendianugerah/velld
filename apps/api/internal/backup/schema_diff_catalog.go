package backup

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
)

type schemaTableName struct{ namespace, name string }

type schemaColumn struct {
	name, definition, typ, collation, defaultExpr, generated, identity, sequence string
	notNull                                                                      bool
}

type schemaObject struct{ kind, name string }

type schemaTable struct {
	columns []schemaColumn
	objects map[schemaObject]string
	options string
}

type databaseSchema struct {
	namespaces map[string]bool
	tables     map[schemaTableName]*schemaTable
}

func readDatabaseSchema(ctx context.Context, conn *connection.StoredConnection) (databaseSchema, error) {
	db, err := openSQLTarget(conn)
	if err != nil {
		return databaseSchema{}, err
	}
	defer db.Close()
	session, err := db.Conn(ctx)
	if err != nil {
		return databaseSchema{}, err
	}
	defer session.Close()
	schema := databaseSchema{namespaces: map[string]bool{}, tables: map[schemaTableName]*schemaTable{}}
	pg := conn.Type == "postgresql"
	if !pg {
		// SHOW CREATE uses the session's identifier-quoting and string-escaping rules.
		if _, err := session.ExecContext(ctx, "SET SESSION sql_quote_show_create = 1, SESSION sql_mode = ''"); err != nil {
			return schema, err
		}
	}
	tx, err := session.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return schema, err
	}
	defer tx.Rollback()
	if pg {
		err = readPostgresSchema(ctx, tx, &schema)
	} else {
		err = readMySQLSchema(ctx, tx, conn.DatabaseName, &schema)
	}
	if err != nil {
		return schema, err
	}
	return schema, tx.Commit()
}

func scanSchemaRows(ctx context.Context, tx *sql.Tx, query string, scan func(*sql.Rows) error, args ...any) error {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func readPostgresSchema(ctx context.Context, tx *sql.Tx, schema *databaseSchema) error {
	if _, err := tx.ExecContext(ctx, "SET LOCAL search_path = pg_catalog, pg_temp"); err != nil {
		return err
	}
	const userNamespace = "n.nspname <> 'information_schema' AND n.nspname !~ '^pg_'"
	if err := scanSchemaRows(ctx, tx, "SELECT n.nspname FROM pg_namespace n WHERE "+userNamespace, func(rows *sql.Rows) error {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		schema.namespaces[name] = true
		return nil
	}); err != nil {
		return err
	}
	if err := scanSchemaRows(ctx, tx, `SELECT n.nspname, c.relname,
		c.relkind = 'p' OR EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = c.oid)
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind IN ('r', 'p') AND `+userNamespace, func(rows *sql.Rows) error {
		var name schemaTableName
		var inherited bool
		if err := rows.Scan(&name.namespace, &name.name, &inherited); err != nil {
			return err
		}
		if inherited {
			return fmt.Errorf("partitioned and inherited PostgreSQL tables are not supported: %w", errSchemaDiffInput)
		}
		schema.tables[name] = &schemaTable{objects: map[schemaObject]string{}}
		return nil
	}); err != nil {
		return err
	}
	if err := scanSchemaRows(ctx, tx, `SELECT n.nspname, c.relname, a.attname,
		format_type(a.atttypid, a.atttypmod),
		CASE WHEN a.attcollation <> t.typcollation THEN format('%I.%I', cn.nspname, co.collname) ELSE '' END,
		COALESCE(pg_get_expr(d.adbin, d.adrelid), ''), a.attnotnull, a.attidentity, a.attgenerated,
		CASE WHEN a.attidentity = '' THEN COALESCE(pg_get_serial_sequence(format('%I.%I', n.nspname, c.relname), a.attname), '') ELSE '' END
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_type t ON t.oid = a.atttypid
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		LEFT JOIN pg_collation co ON co.oid = a.attcollation
		LEFT JOIN pg_namespace cn ON cn.oid = co.collnamespace
		WHERE c.relkind = 'r' AND a.attnum > 0 AND NOT a.attisdropped AND `+userNamespace+`
		ORDER BY c.oid, a.attnum`, func(rows *sql.Rows) error {
		var name schemaTableName
		var col schemaColumn
		if err := rows.Scan(&name.namespace, &name.name, &col.name, &col.typ, &col.collation,
			&col.defaultExpr, &col.notNull, &col.identity, &col.generated, &col.sequence); err != nil {
			return err
		}
		schema.tables[name].columns = append(schema.tables[name].columns, col)
		return nil
	}); err != nil {
		return err
	}
	return scanSchemaRows(ctx, tx, `SELECT n.nspname, c.relname,
		CASE WHEN k.contype = 'f' THEN 'foreign' ELSE 'constraint' END, k.conname, pg_get_constraintdef(k.oid)
		FROM pg_constraint k JOIN pg_class c ON c.oid = k.conrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r' AND `+userNamespace+`
		UNION ALL
		SELECT n.nspname, c.relname, 'index', ix.relname, pg_get_indexdef(i.indexrelid)
		FROM pg_index i JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace JOIN pg_class ix ON ix.oid = i.indexrelid
		WHERE c.relkind = 'r' AND i.indisvalid AND `+userNamespace+`
		AND NOT EXISTS (SELECT 1 FROM pg_constraint k WHERE k.conrelid = c.oid AND k.conindid = i.indexrelid AND k.contype IN ('p','u','x'))`, func(rows *sql.Rows) error {
		var name schemaTableName
		var object schemaObject
		var definition string
		if err := rows.Scan(&name.namespace, &name.name, &object.kind, &object.name, &definition); err != nil {
			return err
		}
		schema.tables[name].objects[object] = definition
		return nil
	})
}

func readMySQLSchema(ctx context.Context, tx *sql.Tx, database string, schema *databaseSchema) error {
	var names []string
	err := scanSchemaRows(ctx, tx, `SELECT TABLE_NAME FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'`, func(rows *sql.Rows) error {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		names = append(names, name)
		return nil
	}, database)
	if err != nil {
		return err
	}
	for _, name := range names {
		var ddl string
		if err := tx.QueryRowContext(ctx, "SHOW CREATE TABLE "+quoteBranchIdentifier(false, name)).Scan(&name, &ddl); err != nil {
			return err
		}
		table, err := parseMySQLTable(ddl)
		if err != nil {
			return err
		}
		schema.tables[schemaTableName{name: name}] = table
	}
	return nil
}

func mysqlQuotedIdentifier(text string) (string, string, bool) {
	if !strings.HasPrefix(text, "`") {
		return "", "", false
	}
	var value strings.Builder
	for i := 1; i < len(text); i++ {
		if text[i] == '`' {
			if i+1 == len(text) || text[i+1] != '`' {
				return value.String(), strings.TrimSpace(text[i+1:]), true
			}
			i++
		}
		value.WriteByte(text[i])
	}
	return "", "", false
}

func parseMySQLTable(ddl string) (*schemaTable, error) {
	_, body, ok := mysqlQuotedIdentifier(strings.TrimPrefix(ddl, "CREATE TABLE "))
	if !ok || !strings.HasPrefix(body, "(") {
		return nil, fmt.Errorf("unrecognized MySQL table definition")
	}
	table := &schemaTable{objects: map[schemaObject]string{}}
	start, depth := 1, 1
	var quote byte
	for i := start; i < len(body); i++ {
		c := body[i]
		if quote != 0 {
			if c == '\\' && quote != '`' {
				i++
			} else if c == quote {
				if i+1 < len(body) && body[i+1] == quote {
					i++
				} else {
					quote = 0
				}
			}
			continue
		}
		switch c {
		case '`', '\'', '"':
			quote = c
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 || (depth == 1 && c == ',') {
			if err := addMySQLTableDefinition(table, strings.TrimSpace(body[start:i])); err != nil {
				return nil, err
			}
			start = i + 1
			if depth == 0 {
				table.options = strings.TrimSpace(body[start:])
				return table, nil
			}
		}
	}
	return nil, fmt.Errorf("incomplete MySQL table definition")
}

func addMySQLTableDefinition(table *schemaTable, definition string) error {
	if name, _, ok := mysqlQuotedIdentifier(definition); ok {
		table.columns = append(table.columns, schemaColumn{name: name, definition: definition})
		return nil
	}
	if strings.HasPrefix(definition, "PRIMARY KEY ") {
		table.objects[schemaObject{"primary", ""}] = definition
		return nil
	}
	for _, prefix := range []string{"KEY ", "UNIQUE KEY ", "FULLTEXT KEY ", "SPATIAL KEY ", "CONSTRAINT "} {
		name, rest, ok := mysqlQuotedIdentifier(strings.TrimPrefix(definition, prefix))
		if !strings.HasPrefix(definition, prefix) || !ok {
			continue
		}
		kind := "index"
		if prefix == "CONSTRAINT " {
			switch {
			case strings.HasPrefix(rest, "FOREIGN KEY "):
				kind = "foreign"
			case strings.HasPrefix(rest, "CHECK "):
				kind = "check"
			default:
				return fmt.Errorf("unsupported MySQL constraint definition: %w", errSchemaDiffInput)
			}
		}
		table.objects[schemaObject{kind, name}] = definition
		return nil
	}
	return fmt.Errorf("unsupported MySQL table definition: %w", errSchemaDiffInput)
}
