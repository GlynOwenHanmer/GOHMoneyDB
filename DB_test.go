package GOHMoneyDB

import (
	"os"
	"io"
	"os/user"
	"testing"
	"errors"
	"database/sql"
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

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := OpenDBConnection("INVALID CONNECTION STRING")
	if DbIsAvailable(unavailableDb) {
		t.Error("isAvailable returned true when it should have been false.")
	}

	availableDb, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error occured whilst prepping DB for test. Error: %s", err.Error())
	}
	if !DbIsAvailable(availableDb) {
		t.Error("isAvailable returned false when it should have been true.")
	}
}
