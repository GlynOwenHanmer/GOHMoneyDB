package postgres

import (
	"database/sql"
	"fmt"
	"strings"
)

// New returns a connection to a postgres DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func New(connectionString string) (s *postgres, err error) {
	var db *sql.DB
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return
	}
	return &postgres{db: db}, nil
}

type postgres struct {
	db *sql.DB
}

func NewConnectionString(host, user, dbname, sslmode string) (s string) {
	kvs := map[string]string{
		"host":    host,
		"user":    user,
		"dbname":  dbname,
		"sslmode": sslmode,
	}
	var options []string
	for k, v := range kvs {
		if len(v) > 0 {
			options = append(options, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return strings.Join(options, " ")
}

// Available returns true if the Storage is available
func (pg postgres) Available() bool {
	return pg.db.Ping() == nil // Ping() returns an error if db  is unavailable
}

func (pg postgres) Close() error {
	return pg.db.Close()
}
