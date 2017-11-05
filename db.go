package moneypostgres

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"bytes"
	"strings"
)

func NewConnectionString(host, user, dbname, sslmode string) (s string, err error) {
	kvs := map[string]string{
		"host":host,
		"user":user,
		"dbname":dbname,
		"sslmode":sslmode,
	}
	cs := new(bytes.Buffer)
	for k, v := range kvs {
		if len(v) > 0 {
			_, err = fmt.Fprintf(cs, "%s=%s ", k,v)
			if err != nil {
				return
			}
		}
	}
	s = strings.TrimSpace(cs.String())
	return
}

// OpenDBConnection returns a connection to a DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func OpenDBConnection(connectionString string) (db *sql.DB, err error) {
	log.Print("Opening DB connection.")
	return sql.Open("postgres", connectionString)
}

func CreateDB(db *sql.DB, name, owner string) error {
	log.Printf("Creating database with name %s and owner %s", name, owner)
	// When using $1 whilst creating a DB with the db driver, errors were being
	// returned to do with the use of $ signs.
	// So I've reverted to plain old forming a query string manually.
	q := new(bytes.Buffer)
	_, err := fmt.Fprintf(q, "CREATE DATABASE %s ", name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(q, "WITH OWNER = %s ", owner)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(q, "ENCODING = 'UTF8' TABLESPACE = pg_default LC_COLLATE = 'en_GB.UTF-8' LC_CTYPE = 'en_GB.UTF-8' CONNECTION LIMIT = 10;")
	if err != nil {
		return err
	}
	_, err = db.Exec(q.String())
	//_, err := db.Exec(`CREATE DATABASE moneytest WITH OWNER = glynhanmer ENCODING = 'UTF8' TABLESPACE = pg_default LC_COLLATE = 'en_GB.UTF-8' LC_CTYPE = 'en_GB.UTF-8' CONNECTION LIMIT = 10;`, name, owner)
	return err
}

func DeleteDB(db *sql.DB, name string) error {
	log.Printf("Deleting database with name %s", name)
	_, err := db.Exec(`DROP DATABASE ` + name)
	return err
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

func deferredCloseDB(db *sql.DB){
	if db == nil {
		log.Printf("Attempted to close db but it was nil.")
	}
	log.Print("Closing DB connection.")
	deferredClose(db)
}

func deferredClose(c io.Closer) {
	if c == nil {
		log.Printf("Attempted to close Closer but it was nil.")
	}
	if err := c.Close(); err != nil {
		log.Printf("Error closing Closer: %s", err)
	}
}
