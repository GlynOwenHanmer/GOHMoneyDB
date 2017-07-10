package GOHMoneyDB

import (
	"database/sql"
	"os"
	"io"
	"os/user"
	"testing"
	"errors"
)

func Test_prepareTestDB(t *testing.T) {
	_, err := prepareTestDB()
	if err != nil {
		t.Fatalf(`Unable to prepare DB for testing.`)
	}
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
	f, err := os.Open(usr.HomeDir + `/.gohmoneydbtestconnectionstring`)
	if err != nil {
		return nil, err
	}
	connectionString := make([]byte, 200)
	bytesCount, err := f.Read(connectionString)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return OpenDBConnection(string(connectionString[0:bytesCount]))
}

