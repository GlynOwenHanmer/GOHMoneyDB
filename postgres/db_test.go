package postgres_test

import (
	"io"
	"os/user"
	"strings"
	"testing"

	"github.com/glynternet/go-accounting-storage/postgres"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
	"github.com/glynternet/go-accounting-storage"
)

func TestNewConnectionString(t *testing.T) {
	c, err := postgres.NewConnectionString("", "name", "", "")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "user=name", c)

	c, err = postgres.NewConnectionString("localhost", "user", "dbname", "disable")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	expected := map[string]string{
		"host":    "localhost",
		"user":    "user",
		"dbname":  "dbname",
		"sslmode": "disable",
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
	cs, err := postgres.NewConnectionString("172.17.0.1", "glynhanmer", "", "disable")
	assert.Nil(t, err)
	name := "moneytest"
	err = postgres.CreateStorage(cs, name, "glynhanmer")
	//todo check it exists
	assert.Nil(t, err)
	err = postgres.DeleteStorage(cs, name)
	//todo check it doesn't exist
}

//todo prepareTestDB should be given a base name for a db, which it should append a timestamp onto.
// prepareTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occured whilst preparing the connection.
func prepareTestDB(t *testing.T) storage.Storage {
	usr, err := user.Current()
	common.FatalIfError(t, err, "Error getting current user")
	if len(usr.HomeDir) < 1 {
		t.Fatalf("User's home directory is zero length")
	}
	connectionString, err := postgres.LoadDBConnectionString(usr.HomeDir + `/.gohmoney/.gohmoneydbtestconnectionstring`)
	common.FatalIfError(t, err, "Error loading DB connection string")
	db, err := postgres.New(connectionString)
	common.FatalIfError(t, err, "Error opening ")
	return db
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, err := postgres.New("INVALID CONNECTION STRING")
	assert.NotNil(t, err)
	if unavailableDb.Available() {
		t.Error("Available returned true when it should have been false.")
	}

	availableDb := prepareTestDB(t)
	assert.True(t, availableDb.Available(), "Available returned false when it should have been true.")
	err = availableDb.Close()
	common.FatalIfError(t, err, "Error closing DB")
}

func TestLoadDBConnectionString(t *testing.T) {
	if _, err := postgres.LoadDBConnectionString(""); err == nil {
		t.Errorf("Expected error but got none.")
	}
	if _, err := postgres.LoadDBConnectionString("asjdhgaksd"); err == nil {
		t.Errorf("Expected error but got none.")
	}
}

func close(t *testing.T, c io.Closer) {
	err := c.Close()
	if err != nil {
		t.Fatal(err)
	}
}
