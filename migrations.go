package gomysql

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrMigrationDestructive = errors.New("migration requires destructive changes")
	ErrMigrationNotNull     = errors.New("migration introduces NOT NULL column without data")
)

type MigrationOptions struct {
	AllowDestructive bool
	Renames          map[string]string // old column name -> new column name
}

type MigrationReport struct {
	Table          string
	AddedColumns   []string
	DroppedColumns []string
	ChangedColumns []string
	RenamedColumns map[string]string // old column name -> new column name
	Rebuilt        bool
}

type columnInfo struct {
	Name string
	Type string
}

type foreignKeyInfo struct {
	From  string
	Table string
	To    string
}

type copyColumnMapping struct {
	destName  string
	srcName   string
	transform func(any) (any, error)
}

func normalizeIdentifier(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeSQLType(typeName string) string {
	return strings.ToUpper(strings.Join(strings.Fields(typeName), " "))
}

func foreignKeyRefsEqual(info *foreignKeyInfo, ref *ForeignKeyRef) bool {
	switch {
	case info == nil && ref == nil:
		return true
	case info == nil || ref == nil:
		return false
	default:
		return normalizeIdentifier(info.Table) == normalizeIdentifier(ref.TableName) &&
			normalizeIdentifier(info.To) == normalizeIdentifier(ref.ColumnName)
	}
}

func lookupForeignKeyRef(foreignKeys map[string]foreignKeyInfo, key string) *foreignKeyInfo {
	info, ok := foreignKeys[key]
	if !ok {
		return nil
	}

	return &info
}

func needsLegacyTimeMigration(srcType string, field RegisteredStructField) bool {
	return field.InternalType == TypeRepTime && normalizeSQLType(srcType) == normalizeSQLType(typeNameString(TypeRepStructBlob))
}

func migrateLegacyTimeValue(raw any) (any, error) {
	if raw == nil {
		return nil, nil
	}

	switch value := raw.(type) {
	case time.Time:
		return formatSQLTimeValue(value), nil
	case string:
		parsed, err := parseSQLTimeValue(value)
		if err != nil {
			return nil, err
		}
		return formatSQLTimeValue(parsed), nil
	case []byte:
		if decoded, handled, err := decodeTimeValue(value); err != nil {
			return nil, err
		} else if handled {
			return formatSQLTimeValue(decoded), nil
		}

		parsed, err := parseSQLTimeValue(string(value))
		if err != nil {
			return nil, err
		}
		return formatSQLTimeValue(parsed), nil
	default:
		return nil, fmt.Errorf("unsupported legacy time value %T", raw)
	}
}

func (d *Driver) tableColumns(table string) ([]columnInfo, error) {
	rows, err := d.db.Query(fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return nil, fmt.Errorf("describe table %s: %w", table, err)
	}
	defer rows.Close()

	var columns []columnInfo
	for rows.Next() {
		var (
			cid     int
			name    string
			sqlType string
			notnull int
			dflt    sql.NullString
			pk      int
		)

		if err := rows.Scan(&cid, &name, &sqlType, &notnull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan table info %s: %w", table, err)
		}

		columns = append(columns, columnInfo{
			Name: name,
			Type: sqlType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate table info %s: %w", table, err)
	}

	return columns, nil
}

func (d *Driver) tableForeignKeys(table string) (map[string]foreignKeyInfo, error) {
	rows, err := d.db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s);", table))
	if err != nil {
		return nil, fmt.Errorf("describe foreign keys %s: %w", table, err)
	}
	defer rows.Close()

	foreignKeys := make(map[string]foreignKeyInfo)
	for rows.Next() {
		var (
			id       int
			seq      int
			refTable string
			from     string
			to       string
			onUpdate string
			onDelete string
			match    string
		)

		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, fmt.Errorf("scan foreign key info %s: %w", table, err)
		}

		// This package only generates single-column foreign keys.
		if seq != 0 {
			continue
		}

		foreignKeys[normalizeIdentifier(from)] = foreignKeyInfo{
			From:  from,
			Table: refTable,
			To:    to,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate foreign key info %s: %w", table, err)
	}

	return foreignKeys, nil
}

func (r *RegisteredStruct[T]) Migrate(opts MigrationOptions) (*MigrationReport, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	report := &MigrationReport{
		Table:          r.Name,
		RenamedColumns: make(map[string]string),
	}

	existingColumns, err := r.db.tableColumns(r.Name)
	if err != nil {
		return report, err
	}

	existingForeignKeys, err := r.db.tableForeignKeys(r.Name)
	if err != nil {
		return report, err
	}

	if len(existingColumns) == 0 {
		if _, err := r.db.db.Exec(r.createTableSQL); err != nil {
			return report, fmt.Errorf("create table %s: %w", r.Name, err)
		}
		for _, field := range r.Fields {
			report.AddedColumns = append(report.AddedColumns, field.Opts.KeyName)
		}
		sort.Strings(report.AddedColumns)
		return report, nil
	}

	existingByKey := make(map[string]columnInfo, len(existingColumns))
	for _, col := range existingColumns {
		existingByKey[normalizeIdentifier(col.Name)] = col
	}

	desiredByKey := make(map[string]RegisteredStructField, len(r.Fields))
	for _, field := range r.Fields {
		desiredByKey[normalizeIdentifier(field.Opts.KeyName)] = field
	}

	renameNewToOld := make(map[string]string, len(opts.Renames))
	for oldName, newName := range opts.Renames {
		oldKey := normalizeIdentifier(oldName)
		newKey := normalizeIdentifier(newName)
		if _, ok := existingByKey[oldKey]; !ok {
			return report, fmt.Errorf("rename source column %s not found", oldName)
		}
		if _, ok := desiredByKey[newKey]; !ok {
			return report, fmt.Errorf("rename target column %s not found in struct", newName)
		}
		renameNewToOld[newKey] = oldKey
	}

	usedExisting := make(map[string]bool, len(existingByKey))
	for _, field := range r.Fields {
		keyName := field.Opts.KeyName
		key := normalizeIdentifier(keyName)

		if oldKey, ok := renameNewToOld[key]; ok {
			oldCol := existingByKey[oldKey]
			report.RenamedColumns[oldCol.Name] = keyName
			usedExisting[oldKey] = true

			if normalizeSQLType(oldCol.Type) != normalizeSQLType(typeNameString(field.InternalType)) ||
				!foreignKeyRefsEqual(lookupForeignKeyRef(existingForeignKeys, oldKey), field.Opts.ForeignKey) {
				report.ChangedColumns = append(report.ChangedColumns, keyName)
			}
			continue
		}

		if col, ok := existingByKey[key]; ok {
			usedExisting[key] = true
			if normalizeSQLType(col.Type) != normalizeSQLType(typeNameString(field.InternalType)) ||
				!foreignKeyRefsEqual(lookupForeignKeyRef(existingForeignKeys, key), field.Opts.ForeignKey) {
				report.ChangedColumns = append(report.ChangedColumns, keyName)
			}
			continue
		}

		report.AddedColumns = append(report.AddedColumns, keyName)
	}

	for key, col := range existingByKey {
		if usedExisting[key] {
			continue
		}
		report.DroppedColumns = append(report.DroppedColumns, col.Name)
	}

	sort.Strings(report.AddedColumns)
	sort.Strings(report.DroppedColumns)
	sort.Strings(report.ChangedColumns)

	needsRebuild := len(report.DroppedColumns) > 0 || len(report.ChangedColumns) > 0 || len(report.RenamedColumns) > 0
	if needsRebuild && !opts.AllowDestructive {
		return report, fmt.Errorf("%w: columns=%v drops=%v renames=%v", ErrMigrationDestructive, report.ChangedColumns, report.DroppedColumns, report.RenamedColumns)
	}

	for _, name := range report.AddedColumns {
		field := desiredByKey[normalizeIdentifier(name)]
		if field.Opts.NotNull {
			return report, fmt.Errorf("%w: column %s", ErrMigrationNotNull, name)
		}
	}

	if needsRebuild {
		if err := r.rebuildTable(existingByKey, renameNewToOld); err != nil {
			return report, err
		}
		report.Rebuilt = true
		return report, nil
	}

	for _, name := range report.AddedColumns {
		field := desiredByKey[normalizeIdentifier(name)]
		columnSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", r.Name, columnDefinition(field, false))
		if _, err := r.db.db.Exec(columnSQL); err != nil {
			return report, fmt.Errorf("add column %s: %w", field.Opts.KeyName, err)
		}

		if field.Opts.Unique {
			indexName := fmt.Sprintf("%s_%s_unique", r.Name, field.Opts.KeyName)
			indexSQL := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(%s);", indexName, r.Name, field.Opts.KeyName)
			if _, err := r.db.db.Exec(indexSQL); err != nil {
				return report, fmt.Errorf("add unique index %s: %w", indexName, err)
			}
		}
	}

	return report, nil
}

func (r *RegisteredStruct[T]) rebuildTable(existingByKey map[string]columnInfo, renameNewToOld map[string]string) error {
	tempName := fmt.Sprintf("%s__gomysql_tmp_%d", r.Name, time.Now().UnixNano())
	createSQL := strings.Replace(r.createTableSQL, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", r.Name), fmt.Sprintf("CREATE TABLE %s", tempName), 1)

	var (
		destCols          []string
		srcCols           []string
		mappings          []copyColumnMapping
		requiresTransform bool
	)

	for _, field := range r.Fields {
		destName := field.Opts.KeyName
		destKey := normalizeIdentifier(destName)
		if oldKey, ok := renameNewToOld[destKey]; ok {
			if oldCol, ok := existingByKey[oldKey]; ok {
				destCols = append(destCols, destName)
				srcCols = append(srcCols, oldCol.Name)
				mapping := copyColumnMapping{
					destName: destName,
					srcName:  oldCol.Name,
				}
				if needsLegacyTimeMigration(oldCol.Type, field) {
					mapping.transform = migrateLegacyTimeValue
					requiresTransform = true
				}
				mappings = append(mappings, mapping)
			}
			continue
		}

		if col, ok := existingByKey[destKey]; ok {
			destCols = append(destCols, destName)
			srcCols = append(srcCols, col.Name)
			mapping := copyColumnMapping{
				destName: destName,
				srcName:  col.Name,
			}
			if needsLegacyTimeMigration(col.Type, field) {
				mapping.transform = migrateLegacyTimeValue
				requiresTransform = true
			}
			mappings = append(mappings, mapping)
		}
	}

	tx, err := r.db.db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", r.Name, err)
	}

	if _, err := tx.Exec("PRAGMA defer_foreign_keys = ON;"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("defer foreign keys for %s: %w", r.Name, err)
	}

	if _, err := tx.Exec(createSQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create temp table %s: %w", tempName, err)
	}

	if len(destCols) > 0 {
		if requiresTransform {
			if err := r.copyRowsWithTransform(tx, tempName, mappings); err != nil {
				_ = tx.Rollback()
				return err
			}
		} else {
			insertSQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s;", tempName, strings.Join(destCols, ", "), strings.Join(srcCols, ", "), r.Name)
			if _, err := tx.Exec(insertSQL); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("copy data into %s: %w", tempName, err)
			}
		}
	}

	if _, err := tx.Exec(fmt.Sprintf("DROP TABLE %s;", r.Name)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("drop old table %s: %w", r.Name, err)
	}

	if _, err := tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tempName, r.Name)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("rename temp table %s: %w", tempName, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", r.Name, err)
	}

	return nil
}

func (r *RegisteredStruct[T]) copyRowsWithTransform(tx *sql.Tx, tempName string, mappings []copyColumnMapping) error {
	destCols := make([]string, 0, len(mappings))
	srcCols := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		destCols = append(destCols, mapping.destName)
		srcCols = append(srcCols, mapping.srcName)
	}

	querySQL := fmt.Sprintf("SELECT %s FROM %s;", strings.Join(srcCols, ", "), r.Name)
	rows, err := tx.Query(querySQL)
	if err != nil {
		return fmt.Errorf("query rows for migration %s: %w", r.Name, err)
	}
	defer rows.Close()

	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		tempName,
		strings.Join(destCols, ", "),
		strings.Repeat("?, ", len(destCols)-1)+"?",
	)

	values := make([]any, len(mappings))
	scanArgs := make([]any, len(mappings))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("scan rows for migration %s: %w", r.Name, err)
		}

		args := make([]any, len(mappings))
		for i, mapping := range mappings {
			args[i] = values[i]
			if mapping.transform != nil {
				transformed, err := mapping.transform(values[i])
				if err != nil {
					return fmt.Errorf("transform column %s during migration %s: %w", mapping.srcName, r.Name, err)
				}
				args[i] = transformed
			}
		}

		if _, err := tx.Exec(insertSQL, args...); err != nil {
			return fmt.Errorf("insert transformed row into %s: %w", tempName, err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate rows for migration %s: %w", r.Name, err)
	}

	return nil
}
