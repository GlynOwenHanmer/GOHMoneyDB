package GOHMoneyDB

import "database/sql"

// OpenDBConnection returns a connection to a DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func OpenDBConnection(connectionString string) (*sql.DB, error) {
	return sql.Open("postgres", connectionString)
}

// DbIsAvailable returns true if a DB is available
func DbIsAvailable(db *sql.DB) bool {
	return db.Ping() == nil // Ping() returns an error if db  is unavailable
}