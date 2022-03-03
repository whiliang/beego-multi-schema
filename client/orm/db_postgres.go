// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import (
	"context"
	"fmt"
	"strconv"
)

// postgresql operators.
var postgresOperators = map[string]string{
	"exact":       "= ?",
	"iexact":      "= UPPER(?)",
	"contains":    "LIKE ?",
	"icontains":   "LIKE UPPER(?)",
	"gt":          "> ?",
	"gte":         ">= ?",
	"lt":          "< ?",
	"lte":         "<= ?",
	"eq":          "= ?",
	"ne":          "!= ?",
	"startswith":  "LIKE ?",
	"endswith":    "LIKE ?",
	"istartswith": "LIKE UPPER(?)",
	"iendswith":   "LIKE UPPER(?)",
}

// postgresql column field types.
var postgresTypes = map[string]string{
	"auto":                "serial NOT NULL PRIMARY KEY",
	"pk":                  "NOT NULL PRIMARY KEY",
	"bool":                "bool",
	"string":              "varchar(%d)",
	"string-char":         "char(%d)",
	"string-text":         "text",
	"time.Time-date":      "date",
	"time.Time":           "timestamp with time zone",
	"int8":                `smallint CHECK("%COL%" >= -127 AND "%COL%" <= 128)`,
	"int16":               "smallint",
	"int32":               "integer",
	"int64":               "bigint",
	"uint8":               `smallint CHECK("%COL%" >= 0 AND "%COL%" <= 255)`,
	"uint16":              `integer CHECK("%COL%" >= 0)`,
	"uint32":              `bigint CHECK("%COL%" >= 0)`,
	"uint64":              `bigint CHECK("%COL%" >= 0)`,
	"float64":             "double precision",
	"float64-decimal":     "numeric(%d, %d)",
	"json":                "json",
	"jsonb":               "jsonb",
	"time.Time-precision": "timestamp(%d) with time zone",
}

// postgresql dbBaser.
type dbBasePostgres struct {
	dbBase
}

var _ dbBaser = new(dbBasePostgres)

// get postgresql operator.
func (d *dbBasePostgres) OperatorSQL(operator string) string {
	return postgresOperators[operator]
}

// generate functioned sql string, such as contains(text).
func (d *dbBasePostgres) GenerateOperatorLeftCol(fi *fieldInfo, operator string, leftCol *string) {
	switch operator {
	case "contains", "startswith", "endswith":
		*leftCol = fmt.Sprintf("%s::text", *leftCol)
	case "iexact", "icontains", "istartswith", "iendswith":
		*leftCol = fmt.Sprintf("UPPER(%s::text)", *leftCol)
	}
}

// postgresql unsupports updating joined record.
func (d *dbBasePostgres) SupportUpdateJoin() bool {
	return false
}

func (d *dbBasePostgres) MaxLimit() uint64 {
	return 0
}

// postgresql quote is ".
func (d *dbBasePostgres) TableQuote() string {
	return `"`
}

// postgresql value placeholder is $n.
// replace default ? to $n.
func (d *dbBasePostgres) ReplaceMarks(query *string) {
	q := *query
	num := 0
	for _, c := range q {
		if c == '?' {
			num++
		}
	}
	if num == 0 {
		return
	}
	data := make([]byte, 0, len(q)+num)
	num = 1
	for i := 0; i < len(q); i++ {
		c := q[i]
		if c == '?' {
			data = append(data, '$')
			data = append(data, []byte(strconv.Itoa(num))...)
			num++
		} else {
			data = append(data, c)
		}
	}
	*query = string(data)
}

// make returning sql support for postgresql.
func (d *dbBasePostgres) HasReturningID(mi *modelInfo, query *string) bool {
	fi := mi.fields.pk
	if fi.fieldType&IsPositiveIntegerField == 0 && fi.fieldType&IsIntegerField == 0 {
		return false
	}

	if query != nil {
		*query = fmt.Sprintf(`%s RETURNING "%s"`, *query, fi.column)
	}
	return true
}

// sync auto key
func (d *dbBasePostgres) setval(ctx context.Context, db dbQuerier, mi *modelInfo, autoFields []string) error {
	if len(autoFields) == 0 {
		return nil
	}

	Q := d.ins.TableQuote()
	for _, name := range autoFields {
		query := fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s', '%s'), (SELECT MAX(%s%s%s) FROM %s%s%s));",
			mi.table, name,
			Q, name, Q,
			Q, mi.table, Q)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

// show table sql for postgresql.
func (d *dbBasePostgres) ShowTablesQuery(ctx context.Context) string {
	query := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')"
	_schema := ctx.Value(ContextKeySchema)
	if nil != _schema {
		schema := _schema.(string)
		if len(schema) > 0 {
			query = fmt.Sprintf("%s AND table_schema='%s'", query, schema)
		}
	}
	return query
}

// show table columns sql for postgresql.
func (d *dbBasePostgres) ShowColumnsQuery(ctx context.Context, table string) string {
	query := fmt.Sprintf("SELECT column_name, data_type, is_nullable FROM information_schema.columns where table_schema NOT IN ('pg_catalog', 'information_schema') and table_name = '%s'", table)
	_schema := ctx.Value(ContextKeySchema)
	if nil != _schema {
		schema := ctx.Value(ContextKeySchema).(string)
		if len(schema) > 0 {
			query = fmt.Sprintf("%s AND table_schema='%s'", query, schema)
		}
	}
	return query
}

// get column types of postgresql.
func (d *dbBasePostgres) DbTypes() map[string]string {
	return postgresTypes
}

// check index exist in postgresql.
func (d *dbBasePostgres) IndexExists(ctx context.Context, db dbQuerier, table string, name string) bool {
	var schema string
	_schema := ctx.Value(ContextKeySchema)
	if nil != _schema {
		schema = _schema.(string)
		if schema == "" {
			schema = "public"
		}
	} else {
		schema = "public"
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM pg_indexes WHERE tablename = '%s' AND indexname = '%s' AND schemaname = '%s'", table, name, schema)
	row := db.QueryRowContext(ctx, query)
	var cnt int
	row.Scan(&cnt)
	return cnt > 0
}

// GenerateSpecifyIndex return a specifying index clause
func (d *dbBasePostgres) GenerateSpecifyIndex(tableName string, useIndex int, indexes []string) string {
	DebugLog.Println("[WARN] Not support any specifying index action, so that action is ignored")
	return ``
}

// create new postgresql dbBaser.
func newdbBasePostgres() dbBaser {
	b := new(dbBasePostgres)
	b.ins = b
	return b
}
