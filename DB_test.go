package GOHMoneyDB_test

import (
	"database/sql"
	"errors"
	"os/user"
	"testing"

	"github.com/GlynOwenHanmer/GOHMoneyDB"
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
	connectionString, err := GOHMoneyDB.LoadDBConnectionString(usr.HomeDir + `/.gohmoneydbtestconnectionstring`)
	if err != nil {
		return nil, err
	}
	return GOHMoneyDB.OpenDBConnection(connectionString)
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := GOHMoneyDB.OpenDBConnection("INVALID CONNECTION STRING")
	if GOHMoneyDB.DbIsAvailable(unavailableDb) {
		t.Error("isAvailable returned true when it should have been false.")
	}

	availableDb, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error occured whilst prepping DB for test. Error: %s", err.Error())
	}
	if !GOHMoneyDB.DbIsAvailable(availableDb) {
		t.Error("isAvailable returned false when it should have been true.")
	}
}

func TestLoadDBConnectionString(t *testing.T) {
	if _, err := GOHMoneyDB.LoadDBConnectionString(""); err == nil {
		t.Errorf("Expected error but got none.")
	}
	if _, err := GOHMoneyDB.LoadDBConnectionString("asjdhgaksd"); err == nil {
		t.Errorf("Expected error but got none.")
	}

}
