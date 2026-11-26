package backup

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

var mysqlAutoIncrementOption = regexp.MustCompile(`^(ENGINE=[^ ]+) AUTO_INCREMENT=[0-9]+`)

func (n schemaTableName) sql(pg bool) string {
	return qualifiedSchemaName(pg, n.namespace, n.name)
}

func qualifiedSchemaName(pg bool, namespace, name string) string {
	if namespace == "" {
		return quoteBranchIdentifier(pg, name)
	}
	return quoteBranchIdentifier(pg, namespace) + "." + quoteBranchIdentifier(pg, name)
}

func schemaTableNames(before, after databaseSchema) []schemaTableName {
	all := make(map[schemaTableName]*schemaTable)
	maps.Copy(all, before.tables)
	maps.Copy(all, after.tables)
	return slices.SortedFunc(maps.Keys(all), func(a, b schemaTableName) int {
		return cmp.Or(cmp.Compare(a.namespace, b.namespace), cmp.Compare(a.name, b.name))
	})
}

func schemaObjectNames(table *schemaTable) []schemaObject {
	return slices.SortedFunc(maps.Keys(table.objects), func(a, b schemaObject) int {
		return cmp.Or(cmp.Compare(a.kind, b.kind), cmp.Compare(a.name, b.name))
	})
}

func schemaColumnsByName(table *schemaTable) map[string]schemaColumn {
	cols := make(map[string]schemaColumn)
	for _, col := range table.columns {
		cols[col.name] = col
	}
	return cols
}

func needsForeignKeyRebuild(before, after databaseSchema) bool {
	for name, old := range before.tables {
		next := after.tables[name]
		if next == nil {
			return true
		}
		cols := schemaColumnsByName(next)
		for _, col := range old.columns {
			newCol, ok := cols[col.name]
			if !ok || col.typ != newCol.typ || col.collation != newCol.collation || col.definition != newCol.definition {
				return true
			}
		}
		for key, definition := range old.objects {
			if key.kind != "foreign" && next.objects[key] != definition {
				return true
			}
		}
	}
	return false
}

func generateMigration(pg bool, before, after databaseSchema) (string, error) {
	var dropForeign, prepare, dropObjects, tables, addObjects, ownSequences, addForeign []string
	rebuildForeign := needsForeignKeyRebuild(before, after)
	if pg {
		for _, namespace := range slices.Sorted(maps.Keys(after.namespaces)) {
			if !before.namespaces[namespace] {
				prepare = append(prepare, "CREATE SCHEMA "+quoteBranchIdentifier(true, namespace))
			}
		}
	}
	for _, name := range schemaTableNames(before, after) {
		old, next := before.tables[name], after.tables[name]
		table := name.sql(pg)
		alter := "ALTER TABLE " + table + " "
		var changes []string
		if old != nil {
			for _, key := range schemaObjectNames(old) {
				same := next != nil && next.objects[key] == old.objects[key]
				if key.kind == "foreign" && (!same || rebuildForeign) {
					dropForeign = append(dropForeign, alter+dropSchemaObject(pg, key))
				} else if key.kind != "foreign" && next != nil && !same {
					if pg && key.kind == "index" {
						dropObjects = append(dropObjects, "DROP INDEX "+qualifiedSchemaName(true, name.namespace, key.name))
					} else if pg {
						dropObjects = append(dropObjects, alter+dropSchemaObject(true, key))
					} else {
						changes = append(changes, dropSchemaObject(false, key))
					}
				}
			}
		}
		if next == nil {
			tables = append(tables, "DROP TABLE "+table)
			continue
		}

		var oldColumns map[string]schemaColumn
		if old != nil {
			oldColumns = schemaColumnsByName(old)
		}
		if pg {
			for _, col := range next.columns {
				if col.sequence != "" && col.sequence != oldColumns[col.name].sequence {
					if col.typ != "smallint" && col.typ != "integer" && col.typ != "bigint" {
						return "", fmt.Errorf("unsupported PostgreSQL owned sequence column type: %w", errSchemaDiffInput)
					}
					prepare = append(prepare, "CREATE SEQUENCE IF NOT EXISTS "+col.sequence+" AS "+col.typ)
					ownSequences = append(ownSequences, "ALTER SEQUENCE "+col.sequence+" OWNED BY "+table+"."+quoteBranchIdentifier(true, col.name))
				}
			}
		}
		var createDefs []string
		if old == nil {
			for _, col := range next.columns {
				createDefs = append(createDefs, schemaColumnDefinition(pg, col))
			}
		} else {
			columnChanges, err := changeSchemaColumns(pg, old, next)
			if err != nil {
				return "", err
			}
			changes = append(changes, columnChanges...)
		}
		for _, key := range schemaObjectNames(next) {
			definition := next.objects[key]
			same := old != nil && old.objects[key] == definition
			switch {
			case key.kind == "foreign" && (!same || rebuildForeign):
				addForeign = append(addForeign, alter+addSchemaObject(pg, key, definition))
			case key.kind == "foreign" || same:
			case pg && key.kind == "index":
				addObjects = append(addObjects, definition)
			case pg:
				addObjects = append(addObjects, alter+addSchemaObject(pg, key, definition))
			case old == nil:
				createDefs = append(createDefs, definition)
			default:
				changes = append(changes, addSchemaObject(false, key, definition))
			}
		}
		if old == nil {
			create := "CREATE TABLE " + table + " ("
			if len(createDefs) != 0 {
				create += "\n    " + strings.Join(createDefs, ",\n    ") + "\n"
			}
			create += ")"
			if !pg && next.options != "" {
				create += " " + mysqlAutoIncrementOption.ReplaceAllString(next.options, "${1}")
			}
			tables = append(tables, create)
		} else if pg {
			for _, change := range changes {
				tables = append(tables, alter+change)
			}
		} else if len(changes) != 0 {
			tables = append(tables, alter+strings.Join(changes, ",\n    "))
		}
	}
	statements := slices.Concat(dropForeign, prepare, dropObjects, tables, addObjects, ownSequences, addForeign)
	if len(statements) == 0 {
		return "", nil
	}
	if pg {
		statements = slices.Concat([]string{"BEGIN", "SET LOCAL search_path = pg_catalog, pg_temp"}, statements, []string{"COMMIT"})
	}
	return strings.Join(statements, ";\n\n") + ";\n", nil
}

func dropSchemaObject(pg bool, key schemaObject) string {
	name := quoteBranchIdentifier(pg, key.name)
	if !pg {
		switch key.kind {
		case "foreign":
			return "DROP FOREIGN KEY " + name
		case "primary":
			return "DROP PRIMARY KEY"
		case "index":
			return "DROP INDEX " + name
		case "check":
			return "DROP CHECK " + name
		}
	}
	return "DROP CONSTRAINT " + name
}

func addSchemaObject(pg bool, key schemaObject, definition string) string {
	if pg {
		return "ADD CONSTRAINT " + quoteBranchIdentifier(true, key.name) + " " + definition
	}
	return "ADD " + definition
}

func schemaColumnDefinition(pg bool, col schemaColumn) string {
	if !pg {
		return col.definition
	}
	definition := quoteBranchIdentifier(true, col.name) + " " + postgresColumnType(col)
	switch {
	case col.identity != "":
		definition += " " + postgresIdentity(col.identity) + " AS IDENTITY"
	case col.generated != "":
		storage := " STORED"
		if col.generated == "v" {
			storage = " VIRTUAL"
		}
		definition += " GENERATED ALWAYS AS (" + col.defaultExpr + ")" + storage
	case col.defaultExpr != "":
		definition += " DEFAULT " + col.defaultExpr
	}
	if col.notNull {
		definition += " NOT NULL"
	}
	return definition
}

func postgresIdentity(identity string) string {
	if identity == "a" {
		return "GENERATED ALWAYS"
	}
	return "GENERATED BY DEFAULT"
}

func postgresColumnType(col schemaColumn) string {
	if col.collation != "" {
		return col.typ + " COLLATE " + col.collation
	}
	return col.typ
}

func changeSchemaColumns(pg bool, old, next *schemaTable) ([]string, error) {
	var changes []string
	oldCols, newCols := schemaColumnsByName(old), schemaColumnsByName(next)
	previous := make(map[string]string)
	var prior string
	for _, col := range old.columns {
		if _, ok := newCols[col.name]; !ok {
			changes = append(changes, "DROP COLUMN "+quoteBranchIdentifier(pg, col.name))
		} else {
			previous[col.name] = prior
			prior = col.name
		}
	}
	prior = ""
	for _, col := range next.columns {
		before, exists := oldCols[col.name]
		if !pg {
			position := " FIRST"
			if prior != "" {
				position = " AFTER " + quoteBranchIdentifier(false, prior)
			}
			if !exists {
				changes = append(changes, "ADD COLUMN "+col.definition+position)
			} else if before != col || previous[col.name] != prior {
				changes = append(changes, "MODIFY COLUMN "+col.definition+position)
			}
		} else if !exists {
			changes = append(changes, "ADD COLUMN "+schemaColumnDefinition(true, col))
		} else {
			alter := "ALTER COLUMN " + quoteBranchIdentifier(true, col.name) + " "
			typeChanged := before.typ != col.typ || before.collation != col.collation
			defaultChanged := before.defaultExpr != col.defaultExpr || typeChanged
			if before.generated != col.generated || (col.generated != "" && defaultChanged) {
				return nil, fmt.Errorf("altering a generated PostgreSQL column is not supported: %w", errSchemaDiffInput)
			}
			if before.identity != "" && col.identity == "" {
				changes = append(changes, alter+"DROP IDENTITY")
			}
			if before.defaultExpr != "" && defaultChanged {
				changes = append(changes, alter+"DROP DEFAULT")
			}
			if typeChanged {
				changes = append(changes, alter+"TYPE "+postgresColumnType(col))
			}
			if before.notNull != col.notNull {
				nullability := "DROP NOT NULL"
				if col.notNull {
					nullability = "SET NOT NULL"
				}
				changes = append(changes, alter+nullability)
			}
			if col.defaultExpr != "" && defaultChanged {
				changes = append(changes, alter+"SET DEFAULT "+col.defaultExpr)
			}
			if before.identity != col.identity && col.identity != "" {
				if before.identity == "" {
					changes = append(changes, alter+"ADD "+postgresIdentity(col.identity)+" AS IDENTITY")
				} else {
					changes = append(changes, alter+"SET "+postgresIdentity(col.identity))
				}
			}
		}
		prior = col.name
	}
	return changes, nil
}
