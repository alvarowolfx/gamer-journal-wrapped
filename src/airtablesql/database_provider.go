package airtablesql

import (
	"fmt"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/mehanizm/airtable"
)

type AirtableSQLProvider struct {
	client *airtable.Client
	bases  []*airtable.Base
	dbs    []sql.Database
}

func NewAirtableSQLProvider(client *airtable.Client) (*AirtableSQLProvider, error) {
	bases, err := client.GetBases().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list airtable bases: %v", err)
	}

	p := &AirtableSQLProvider{
		client: client,
		bases:  bases.Bases,
	}

	return p, nil
}

var _ sql.DatabaseProvider = &AirtableSQLProvider{}

// Database gets a Database from the provider.
func (p *AirtableSQLProvider) Database(ctx *sql.Context, name string) (sql.Database, error) {
	if p.dbs != nil {
		for _, db := range p.dbs {
			if name == db.Name() {
				return db, nil
			}
		}
	}
	for _, b := range p.bases {
		if name == util.ToSnakecase(b.Name) {
			db, err := p.databaseFromAirtableBase(ctx, p.client, b)
			if err != nil {
				return nil, fmt.Errorf("failed to convert airtable base to mysql db: %v", err)
			}
			return db, nil
		}
	}
	return nil, fmt.Errorf("%q database not found", name)
}

// HasDatabase checks if the Database exists in the provider.
func (p *AirtableSQLProvider) HasDatabase(ctx *sql.Context, name string) bool {
	for _, b := range p.bases {
		if name == util.ToSnakecase(b.Name) {
			return true
		}
	}
	return false
}

// AllDatabases returns a slice of all Databases in the provider.
func (p *AirtableSQLProvider) AllDatabases(ctx *sql.Context) []sql.Database {
	if p.dbs != nil {
		return p.dbs
	}
	p.dbs = []sql.Database{}
	for _, b := range p.bases {
		db, err := p.Database(ctx, util.ToSnakecase(b.Name))
		if err != nil {
			continue
		}
		p.dbs = append(p.dbs, db)
	}
	return p.dbs
}

func (p *AirtableSQLProvider) databaseFromAirtableBase(sctx *sql.Context, client *airtable.Client, base *airtable.Base) (sql.Database, error) {
	db := memory.NewDatabase(util.ToSnakecase(base.Name))
	db.EnablePrimaryKeyIndexes()

	airtables, err := client.GetBaseSchema(base.ID).Do()
	if err != nil {
		return nil, err
	}

	for _, ts := range airtables.Tables {
		table := NewTable(base, ts, p)
		db.AddTable(table.Name(), table)
	}

	return db, nil
}
