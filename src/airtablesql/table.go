package airtablesql

import (
	"fmt"
	"io"
	"time"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/util"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/mehanizm/airtable"
)

const (
	cachedPages    = 100
	recordCacheTTL = 10 * time.Second
)

type table struct {
	name        string
	schema      sql.Schema
	table       *airtable.Table
	tableSchema *airtable.TableSchema
	cache       *expirable.LRU[string, airtable.Records]

	parent *Provider
}

var _ sql.Table = &table{}

func NewTable(base *airtable.Base, ts *airtable.TableSchema, provider *Provider) sql.Table {
	schema := tableSchemaFromAirtable(ts)
	cache := expirable.NewLRU[string, airtable.Records](cachedPages, nil, recordCacheTTL)
	return &table{
		name:        util.ToSnakecase(ts.Name),
		table:       provider.client.GetTable(base.ID, ts.ID),
		cache:       cache,
		schema:      schema,
		parent:      provider,
		tableSchema: ts,
	}
}

// Name returns the name.
func (t *table) Name() string {
	return t.name
}

// Implements fmt.Stringer
func (t *table) String() string {
	return t.name
}

// Schema returns the table's schema.
func (t *table) Schema() sql.Schema {
	return t.schema
}

// Collation returns the table's collation.
func (t *table) Collation() sql.CollationID {
	return sql.Collation_Default
}

func (t *table) GetRecords(offset string) (*airtable.Records, error) {
	cacheKey := fmt.Sprintf("records:%s:%s", t.name, offset)
	if v, found := t.cache.Get(cacheKey); found {
		return &v, nil
	}
	call := t.table.GetRecords()
	if offset != "" {
		call = call.WithOffset(offset)
	}
	records, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get records from table %q: %v", t.name, err)
	}
	t.cache.Add(cacheKey, *records)

	return records, nil
}

// Partitions returns the table's partitions in an iterator.
func (t *table) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &pageIter{
		parent:     t,
		lastOffset: "",
		first:      true,
	}, nil
}

// PartitionRows returns the rows in the given partition, which was returned by Partitions.
func (t *table) PartitionRows(ctx *sql.Context, p sql.Partition) (sql.RowIter, error) {
	return &rowIter{
		schema:  t.schema,
		ts:      t.tableSchema,
		records: p.(*page).records.Records,
	}, nil
}

type page struct {
	records *airtable.Records
}

func (p *page) Key() []byte {
	if p.records.Offset != "" {
		return []byte("last")
	}
	return []byte(p.records.Offset)
}

type pageIter struct {
	first      bool
	lastOffset string
	parent     *table
}

func (it *pageIter) Close(ctx *sql.Context) error {
	return nil
}

func (it *pageIter) Next(ctx *sql.Context) (sql.Partition, error) {
	if !it.first && it.lastOffset == "" {
		return nil, io.EOF
	}

	records, err := it.parent.GetRecords(it.lastOffset)
	if err != nil {
		return nil, err
	}
	it.lastOffset = records.Offset
	it.first = false

	return &page{
		records: records,
	}, nil
}
