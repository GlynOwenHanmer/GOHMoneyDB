package postgres_test

import (
	"io"
	"strings"
	"testing"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

const (
	host       = "localhost"
	testDBName = "moneytest"
	user       = "glynhanmer"
	ssl        = "disable"
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

func TestCreateAndDeleteStorage(t *testing.T) {
	cs := adminConnectionString(t)
	err := postgres.CreateStorage(cs, testDBName, user)
	//todo check it exists
	assert.Nil(t, err)
	err = postgres.DeleteStorage(cs, testDBName)
	//todo check it doesn't exist
	common.FatalIfError(t, err, "deleting storage")
}

func Test_createTestDB(t *testing.T) {
	db := createTestDB(t)
	if !assert.NotNil(t, db, `Unable to prepare DB for testing`) {
		t.Fail()
	}
	close(t, db)
	deleteTestDB(t)
}

//todo createTestDB should be given a base name for a db, which it should append a timestamp onto.
// createTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occurred whilst preparing the connection.
func createTestDB(t *testing.T) storage.Storage {
	cs := adminConnectionString(t)
	err := postgres.CreateStorage(cs, testDBName, user)
	common.FatalIfError(t, err, "Error creating storage ")
	cs, err = postgres.NewConnectionString(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "Error creating connection string for storage access")
	db, err := postgres.New(cs)
	common.FatalIfError(t, err, "Error creating DB connection")
	return db
}

func deleteTestDB(t *testing.T) {
	cs := adminConnectionString(t)
	err := postgres.DeleteStorage(cs, testDBName)
	common.FatalIfError(t, err, "Error creating storage ")
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := postgres.New("INVALID CONNECTION STRING")
	assert.False(t, unavailableDb.Available(), "Storage should not be available")
	availableDb := createTestDB(t)
	assert.True(t, availableDb.Available(), "Available returned false when it should have been true.")
	close(t, availableDb)
	deleteTestDB(t)
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
	common.FatalIfErrorf(t, c.Close(), "Error closing io.Closer %v", c)
}

func adminConnectionString(t *testing.T) string {
	cs, err := postgres.NewConnectionString(host, user, "", ssl)
	common.FatalIfError(t, err, "generating new admin connection string")
	return cs
}
