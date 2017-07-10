package GOHMoneyDB

import "database/sql"

// OpenDBConnection returns a connection to a DB using the given connection string along with any errors that occur whilst attempting to open the connection.
func OpenDBConnection(connectionString string) (*sql.DB, error) {
	return sql.Open("postgres", connectionString)
}