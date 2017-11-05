package moneypostgres_test

import (
	"testing"
	"github.com/glynternet/go-money/common"
	"database/sql"
	"os/user"
	"github.com/glynternet/go-moneypostgres"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
)

func TestNewConnectionString(t *testing.T) {
	c, err := moneypostgres.NewConnectionString("", "name", "", "")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "user=name", c)

	c, err = moneypostgres.NewConnectionString("localhost", "user", "dbname", "disable")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	expected := map[string]string{
		"host":"localhost",
		"user":"user",
		"dbname":"dbname",
		"sslmode":"disable",
	}
	ss := strings.Split(c, ` `)
	assert.Len(t, ss, len(expected))
	for _, s := range ss {
		kv := strings.Split(s, `=`)
		assert.Len(t, kv, 2)
		key := kv[0]
		v, ok := expected[key]
		assert.True(t, ok)
		assert.Equal(t, v, kv[1])
		delete(expected, key)
	}
	// Should be none left
	assert.Len(t, expected, 0)
}

func Test_prepareTestDB(t *testing.T) {
	db := prepareTestDB(t)
	if db == nil {
		t.Fatalf(`Unable to prepare DB for testing`)
	}
	err := db.Close()
	common.FatalIfError(t, err, "Error closing test DB")
}

func TestCreateAndDeleteDB(t *testing.T) {
	cs, err := moneypostgres.NewConnectionString("172.17.0.1", "glynhanmer", "", "disable")
	assert.Nil(t, err)
	db, err := moneypostgres.OpenDBConnection(cs)
	assert.Nil(t, err)
	err = moneypostgres.CreateDB(db, "moneytest", "glynhanmer")
	assert.Nil(t, err)
	err = moneypostgres.DeleteDB(db, "moneytest")
}

//todo prepareTestDB should be given a base name for a db, which it should append a timestamp onto.
// prepareTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occured whilst preparing the connection.
func prepareTestDB(t *testing.T) *sql.DB {
	usr, err := user.Current()
	common.FatalIfError(t, err, "Error getting current user")
	if len(usr.HomeDir) < 1 {
		t.Fatalf("User's home directory is zero length")
	}
	connectionString, err := moneypostgres.LoadDBConnectionString(usr.HomeDir + `/.gohmoney/.gohmoneydbtestconnectionstring`)
	common.FatalIfError(t, err, "Error loading DB connection string")
	db, err := moneypostgres.OpenDBConnection(connectionString)
	common.FatalIfError(t, err, "Error opening ")
	return db
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := moneypostgres.OpenDBConnection("INVALID CONNECTION STRING")
	if moneypostgres.DbIsAvailable(unavailableDb) {
		t.Error("isAvailable returned true when it should have been false.")
	}

	availableDb := prepareTestDB(t)
	if !moneypostgres.DbIsAvailable(availableDb) {
		t.Error("isAvailable returned false when it should have been true.")
	}
	err := availableDb.Close()
	common.FatalIfError(t, err, "Error closing DB")
}

func TestLoadDBConnectionString(t *testing.T) {
	if _, err := moneypostgres.LoadDBConnectionString(""); err == nil {
		t.Errorf("Expected error but got none.")
	}
	if _, err := moneypostgres.LoadDBConnectionString("asjdhgaksd"); err == nil {
		t.Errorf("Expected error but got none.")
	}
}

func close(t *testing.T, c io.Closer) {
	err := c.Close()
	if err != nil {
		t.Fatal(err)
	}
}
