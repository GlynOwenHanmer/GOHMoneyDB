package GOHMoneyDB

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// OpenDBConnection returns a connection to a DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func OpenDBConnection(connectionString string) (*sql.DB, error) {
	log.Print("Opening DB connection.")
	return sql.Open("postgres", connectionString)
}

// DbIsAvailable returns true if a DB is available
func DbIsAvailable(db *sql.DB) bool {
	return db.Ping() == nil // Ping() returns an error if db  is unavailable
}

// LoadDBConnectionString loads the connection string to be used when connecting to the database.
// LoadDBConnectionString will return the connection string and an error description if there is one.
func LoadDBConnectionString(location string) (string, error) {
	if len(location) < 1 {
		return ``, errors.New("no connection string file location given")
	}
	file, err := os.Open(location)
	if err != nil {
		return ``, err
	}
	stat, err := file.Stat()
	if err != nil {
		return ``, err
	}
	maxConnectionString := int64(200)
	fileSize := stat.Size()
	if fileSize > maxConnectionString {
		message := fmt.Sprintf("Connection string file (%s) is too large. Max: %d, Length: %d", location, maxConnectionString, fileSize)
		return ``, errors.New(message)
	}
	connectionString := make([]byte, maxConnectionString)
	bytesCount, err := file.Read(connectionString)
	if err != nil && err != io.EOF {
		return ``, err
	}
	return string(connectionString[0:bytesCount]), err
}

func close(c io.Closer) {
	if c == nil {
		log.Printf("Attempted to close db but it was nil.")
	}
	if err := c.Close(); err != nil {
		log.Printf("Error closing db: %s", err)
	}
}