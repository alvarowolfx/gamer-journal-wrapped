package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/joho/godotenv"
	"github.com/mehanizm/airtable"
)

const (
	address           = "localhost"
	port              = 3306
	recordIDFieldName = "record_id"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to read .env: %v \n", err)
	}

	airtableAPIKey := os.Getenv("AIRTABLE_API_KEY")
	client := airtable.NewClient(airtableAPIKey)

	bases, err := client.GetBases().Do()
	if err != nil {
		log.Fatalf("failed to list airtable bases: %v", err)
	}

	sctx := sql.NewEmptyContext()
	dbs := []sql.Database{}
	for _, b := range bases.Bases {
		db, err := createDatabaseFromAirtableBase(sctx, client, b)
		if err != nil {
			log.Fatalf("failed to convert airtable base to mysql db: %v", err)
		}
		dbs = append(dbs, db)
	}

	engine := sqle.NewDefault(memory.NewDBProvider(dbs...))

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", address, port),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		log.Fatalf("failed to create mysql server: %v", err)
	}
	if err = s.Start(); err != nil {
		log.Fatalf("failed to start mysql server: %v", err)
	}
}

func createDatabaseFromAirtableBase(sctx *sql.Context, client *airtable.Client, base *airtable.Base) (sql.Database, error) {
	dbName := base.Name
	db := memory.NewDatabase(dbName)
	db.EnablePrimaryKeyIndexes()

	airtables, err := client.GetBaseSchema(base.ID).Do()
	if err != nil {
		return nil, err
	}

	for _, ts := range airtables.Tables {
		tableName := toSnakecase(ts.Name)
		schema := tableSchemaFromAirtable(ts)
		fk := db.GetForeignKeyCollection()
		table := memory.NewTable(tableName, sql.NewPrimaryKeySchema(schema), fk)
		db.AddTable(tableName, table)

		atable := client.GetTable(base.ID, ts.ID)
		records, err := fetchAllRecords(atable)
		if err != nil {
			return nil, fmt.Errorf("failed to get records from table %q: %v", tableName, err)
		}

		inserter := table.Inserter(sctx)
		for _, rec := range records {
			row, err := rowFromRecord(schema, ts, rec)
			if err != nil {
				return nil, fmt.Errorf("failed to convert record to row for table %q: %v", tableName, err)
			}
			err = inserter.Insert(sctx, row)
			if err != nil {
				return nil, fmt.Errorf("failed to insert data into table %q: %v", tableName, err)
			}
		}
		err = inserter.Close(sctx)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func fetchAllRecords(atable *airtable.Table) ([]*airtable.Record, error) {
	allRecords := []*airtable.Record{}
	offset := ""
	for {
		call := atable.GetRecords()
		if offset != "" {
			call = call.WithOffset(offset)
		}
		records, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to get records from table %v: %v", atable, err)
		}
		offset = records.Offset
		allRecords = append(allRecords, records.Records...)
		if offset == "" {
			break
		}
	}
	return allRecords, nil
}

func tableSchemaFromAirtable(tableSchema *airtable.TableSchema) sql.Schema {
	tableName := toSnakecase(tableSchema.Name)
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
			Name:       toSnakecase(field.Name),
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

func toSnakecase(name string) string {
	lower := strings.ToLower(name)
	clean := strings.ReplaceAll(strings.ReplaceAll(lower, "(", ""), ")", "")
	elems := strings.Split(clean, " ")
	return strings.Join(elems, "_")
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
			if column.Name == toSnakecase(f.Name) {
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
