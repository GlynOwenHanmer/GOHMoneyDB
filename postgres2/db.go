package postgres2

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

const driver = "postgres"

// New returns a connection to a postgres Storage using the given connection string along with any errors that occur whilst attempting to open the connection.
func New(connectionString string) (s *postgres, err error) {
	var db *sql.DB
	db, err = open(connectionString)
	if err != nil {
		return
	}
	return &postgres{db: db}, nil
}

type postgres struct {
	db *sql.DB
}

// dbname can be an empty string when you are connecting to create the Storage
func NewConnectionString(host, user, dbname, sslmode string) (s string, err error) {
	if len(strings.TrimSpace(host)) == 0 {
		err = errors.New("storage host must be non-whitespace and longer than 0 characters")
		return
	}
	if len(strings.TrimSpace(user)) == 0 {
		err = errors.New("storage user must be non-whitespace and longer than 0 characters")
		return
	}
	switch sslmode {
	case "enable", "disable":
	default:
		err = errors.New("storage sslmode must be value enable or disable")
		return
	}
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

func (w *failSafeWriter) writef(format string, args ...interface{}) {
	if w.error != nil {
		return
	}
	bs := []byte(fmt.Sprintf(format, args...))
	_, w.error = w.Writer.Write(bs)
}

func CreateStorage(host, user, dbname, sslmode string) error {
	adminConnect, err := NewConnectionString(host, user, "", sslmode)
	if err != nil {
		return err
	}
	userConnect, err := NewConnectionString(host, user, dbname, sslmode)
	if err != nil {
		return err
	}
	err = createDatabase(adminConnect, dbname, user)
	if err != nil {
		return err
	}
	return createAccountsTable(userConnect)
}

func createDatabase(connection, name, owner string) error {
	// When using $1 whilst creating a DB with the db driver, errors were being
	// returned to do with the use of $ signs.
	// So I've reverted to plain old forming a query string manually.
	q := new(bytes.Buffer)
	w := failSafeWriter{Writer: q}
	w.writef("CREATE DATABASE %s ", name)
	w.writef("WITH OWNER = %s ", owner)
	w.writef("ENCODING = 'UTF8' TABLESPACE = pg_default LC_COLLATE = 'en_GB.UTF-8' LC_CTYPE = 'en_GB.UTF-8' CONNECTION LIMIT = 10;")
	if w.error != nil {
		return w.error
	}
	db, err := open(connection)
	if err != nil {
		return err
	}
	defer nonReturningCloseDB(db)
	_, err = db.Exec(q.String())
	return err
}

func createAccountsTable(connection string) error {
	db, err := open(connection)
	if err != nil {
		return err
	}
	defer nonReturningCloseDB(db)
	_, err = db.Exec(`CREATE TABLE accounts (
	id SERIAL PRIMARY KEY,
	name varchar(100) NOT NULL,
	currency char(3) NOT NULL,
	opened timestamp with time zone NOT NULL,
	closed timestamp with time zone,
	deleted timestamp with time zone
);`)
	return err
}

func DeleteStorage(host, user, name, sslmode string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return errors.New("storage name must be non-whitespace and longer than 0 characters")
	}
	adminConnect, err := NewConnectionString(host, user, "", sslmode)
	if err != nil {
		return err
	}
	db, err := open(adminConnect)
	if err != nil {
		return err
	}
	defer nonReturningCloseDB(db)
	_, err = db.Exec(`DROP DATABASE ` + name)
	return err
}

func open(connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionString)
	return db, err
}

// Available returns true if the Storage is available
func (s *postgres) Available() bool {
	return s.db.Ping() == nil // Ping() returns an error if db  is unavailable
}

func (s postgres) Close() error {
	return s.db.Close()
}

func nonReturningCloseDB(db *sql.DB) {
	if db == nil {
		log.Printf("Attempted to close db but it was nil.")
	}
	if err := db.Close(); err != nil {
		log.Printf("Error closing Closer: %s", err)
	}
}
