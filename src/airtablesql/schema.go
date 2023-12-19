package airtablesql

import (
	"fmt"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/mehanizm/airtable"
)

const (
	recordIDFieldName = "record_id"
)

func tableSchemaFromAirtable(tableSchema *airtable.TableSchema) sql.Schema {
	tableName := util.ToSnakecase(tableSchema.Name)
	schema := sql.Schema{
		&sql.Column{
			Name:       recordIDFieldName,
			Type:       types.Text,
			Nullable:   false,
			Source:     tableName,
			PrimaryKey: true,
		},
	}
	for _, field := range tableSchema.Fields {
		col := &sql.Column{
			Name:       util.ToSnakecase(field.Name),
			Type:       fromAirtableType(field.Type),
			Comment:    fmt.Sprintf("airtable type: %s; airtable field: %s", field.Type, field.Name),
			Nullable:   true,
			Source:     tableName,
			PrimaryKey: field.ID == tableSchema.PrimaryFieldID,
		}
		schema = append(schema, col)
	}
	return schema
}

func fromAirtableType(ftype string) sql.Type {
	switch ftype {
	case "date":
		return types.Date
	case "autoNumber":
		return types.Float64
	case "singleSelect", "multilineText", "singleLineText":
		return types.Text
	case "multipleRecordLinks", "multipleLookupValues":
		return types.JSON
	default:
		return types.Text
	}
}
