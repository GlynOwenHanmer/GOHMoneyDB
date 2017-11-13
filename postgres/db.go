package postgres

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// New returns a connection to a postgres DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func New(connectionString string) (s *storage, err error) {
	var db *sql.DB
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return
	}
	return &storage{db:db}, nil
}

type storage struct {
	db *sql.DB
}

func NewConnectionString(host, user, dbname, sslmode string) (s string, err error) {
	kvs := map[string]string{
		"host":    host,
		"user":    user,
		"dbname":  dbname,
		"sslmode": sslmode,
	}
	cs := new(bytes.Buffer)
	for k, v := range kvs {
		if len(v) > 0 {
			_, err = fmt.Fprintf(cs, "%s=%s ", k, v)
			if err != nil {
				return
			}
		}
	}
	s = strings.TrimSpace(cs.String())
	return
}

type failSafeWriter struct {
	io.Writer
	error
}

func(w *failSafeWriter) writef(format string, args ...interface{}) {
	if w.error != nil {
		return
	}
	bs := []byte(fmt.Sprintf(format, args...))
	_, w.error = w.Writer.Write(bs)
}

func CreateStorage(connectionString, name, owner string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return errors.New("storage name must be non-whitespace and longer than 0 characters")
	}
	if len(strings.TrimSpace(owner)) == 0 {
		return errors.New("owner must be non-whitespace and longer than 0 characters")
	}
	// When using $1 whilst creating a DB with the db driver, errors were being
	// returned to do with the use of $ signs.
	// So I've reverted to plain old forming a query string manually.
	q := new(bytes.Buffer)
	w := failSafeWriter{Writer:q}
	w.writef("CREATE DATABASE %s ", name)
	w.writef("WITH OWNER = %s ", owner)
	w.writef("ENCODING = 'UTF8' TABLESPACE = pg_default LC_COLLATE = 'en_GB.UTF-8' LC_CTYPE = 'en_GB.UTF-8' CONNECTION LIMIT = 10;")
	if w.error != nil {
		return w.error
	}
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	defer deferredClose(db)
	_, err = db.Exec(q.String())
	return err
}

func DeleteStorage(connectionString, name string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return errors.New("storage name must be non-whitespace and longer than 0 characters")
	}
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DROP DATABASE ` + name)
	return err
}

// Available returns true if the Storage is available
func (s *storage)Available() bool {
	return s.db.Ping() == nil // Ping() returns an error if db  is unavailable
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

func (s storage) Close() error {
	return s.db.Close()
}

func deferredCloseDB(s storage) {
	if s.db == nil {
		log.Printf("Attempted to close db but it was nil.")
	}
	deferredClose(s)
}

func deferredClose(c io.Closer) {
	if c == nil {
		log.Printf("Attempted to close Closer but it was nil.")
	}
	if err := c.Close(); err != nil {
		log.Printf("Error closing Closer: %s", err)
	}
}
