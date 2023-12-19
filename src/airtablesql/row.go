package airtablesql

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/mehanizm/airtable"
)

type rowIter struct {
	schema  sql.Schema
	ts      *airtable.TableSchema
	records []*airtable.Record
}

// Next retrieves the next row. It will return io.EOF if it's the last row.
// After retrieving the last row, Close will be automatically closed.
func (it *rowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if len(it.records) == 0 {
		return nil, io.EOF
	}
	record := it.records[0]
	it.records = it.records[1:]
	return rowFromRecord(it.schema, it.ts, record)
}

func (it *rowIter) Close(ctx *sql.Context) error {
	return nil
}

func rowFromRecord(schema sql.Schema, ts *airtable.TableSchema, rec *airtable.Record) (sql.Row, error) {
	values := []any{}
	for _, column := range schema {
		if column.Name == recordIDFieldName {
			values = append(values, rec.ID)
			continue
		}
		originalFieldName := column.Name
		for _, f := range ts.Fields {
			if column.Name == util.ToSnakecase(f.Name) {
				originalFieldName = f.Name
				break
			}
		}
		if value, ok := rec.Fields[originalFieldName]; ok {
			sval := fmt.Sprintf("%v", value)
			if column.Type == types.Float64 {
				values = append(values, value)
				continue
			}
			if column.Type == types.Date {
				t, err := time.Parse(time.DateOnly, sval)
				if err != nil {
					return nil, err
				}
				values = append(values, t)
				continue
			}
			if strings.Contains(sval, "[") && column.Type == types.JSON {
				sval = strings.TrimPrefix(sval, "[")
				sval = strings.TrimSuffix(sval, "]")
				parts := strings.Split(sval, " ")
				sval = "[\"" + strings.Join(parts, "\",\"") + "\"]"
				doc := types.MustJSON(sval)
				values = append(values, doc)
			} else {
				values = append(values, sval)
			}
		} else {
			values = append(values, nil)
		}
	}
	return sql.NewRow(values...), nil
}
