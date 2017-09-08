package GOHMoneyDB

import (
	"database/sql"
	"errors"
	"os"
	"fmt"
	"io"
	"os/user"
)

// OpenDBConnection returns a connection to a DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func OpenDBConnection(connectionString string) (*sql.DB, error) {
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
		return ``, errors.New("No connection string file location given.")
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

// prepareTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occured whilst preparing the connection.
func prepareTestDB() (*sql.DB, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	if len(usr.HomeDir) < 1 {
		return nil, errors.New("No home directory for current user.")
	}
	connectionString, err := LoadDBConnectionString(usr.HomeDir + `/.gohmoneydbtestconnectionstring`)
	if err != nil {
		return nil, err
	}
	return OpenDBConnection(connectionString)
}