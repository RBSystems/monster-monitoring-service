package accessors

import (
	_ "github.com/go-sql-driver/mysql" // Blank import due to its use as a driver

	"database/sql"
	"log"
)

// AccessorGroup holds all configuration for the accessors.
type AccessorGroup struct {
	Database *sql.DB
}

// Open creates a database connection and sets it in the struct
func (accessorGroup *AccessorGroup) Open(dataSourceName string) {
	database, err := sql.Open("mysql", dataSourceName)
	if err != nil { // Die if there was an error
		log.Panicf("Could not connect to the database: %s\n", err)
	}

	accessorGroup.Database = database
}
