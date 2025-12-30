package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alvarowolfx/gamer-journal-wrapped/src/airtablesql"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/server"
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

	recordCacheTTL := os.Getenv("AIRTABLE_RECORD_CACHE_TTL")
	recordCacheTTLDuration, err := time.ParseDuration(recordCacheTTL)
	if err != nil {
		log.Printf("failed to parse record cache ttl: %v \n", err)
		recordCacheTTLDuration = 1 * time.Minute
	}
	provider, err := airtablesql.NewProvider(client, recordCacheTTLDuration)
	if err != nil {
		log.Fatalf("failed to init airtable sql provider: %v", err)
	}

	engine := sqle.NewDefault(provider)

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
