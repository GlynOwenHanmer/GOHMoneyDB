package GOHMoneyDB_test

import (
	"database/sql"
	"os/user"
	"testing"

	"github.com/glynternet/GOHMoneyDB"
	"io"
	"github.com/glynternet/GOHMoney/common"
)

func Test_prepareTestDB(t *testing.T) {
	db := prepareTestDB(t)
	if db == nil {
		t.Fatalf(`Unable to prepare DB for testing`)
	}
	err := db.Close()
	common.FatalIfError(t, err, "Error closing test DB")
}

// prepareTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occured whilst preparing the connection.
func prepareTestDB(t *testing.T) *sql.DB {
	usr, err := user.Current()
	common.FatalIfError(t, err, "Error getting current user")
	if len(usr.HomeDir) < 1 {
		t.Fatalf("User's home directory is zero length")
	}
	connectionString, err := GOHMoneyDB.LoadDBConnectionString(usr.HomeDir + `/.gohmoney/.gohmoneydbtestconnectionstring`)
	common.FatalIfError(t, err, "Error loading DB connection string")
	db, err := GOHMoneyDB.OpenDBConnection(connectionString)
	common.FatalIfError(t, err, "Error opening ")
	return db
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := GOHMoneyDB.OpenDBConnection("INVALID CONNECTION STRING")
	if GOHMoneyDB.DbIsAvailable(unavailableDb) {
		t.Error("isAvailable returned true when it should have been false.")
	}

	availableDb := prepareTestDB(t)
	if !GOHMoneyDB.DbIsAvailable(availableDb) {
		t.Error("isAvailable returned false when it should have been true.")
	}
	err := availableDb.Close()
	common.FatalIfError(t, err, "Error closing DB")
}

func TestLoadDBConnectionString(t *testing.T) {
	if _, err := GOHMoneyDB.LoadDBConnectionString(""); err == nil {
		t.Errorf("Expected error but got none.")
	}
	if _, err := GOHMoneyDB.LoadDBConnectionString("asjdhgaksd"); err == nil {
		t.Errorf("Expected error but got none.")
	}
}

func close(t *testing.T, c io.Closer) {
	err := c.Close()
	if err != nil {
		t.Fatal(err)
	}
}
