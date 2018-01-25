package postgres2

import (
	"strings"
	"testing"

	"database/sql"

	"github.com/glynternet/go-accounting-storage"
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
	c, err := NewConnectionString("localhost", "user", "dbname", "disable")
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

func TestCreateStorage_InvalidParamaters(t *testing.T) {
	assert.NotNil(t, CreateStorage("", "user", "dbname", "sslmode"), "expected error for empty host")
	assert.NotNil(t, CreateStorage("host", "", "dbname", "sslmode"), "expected error for empty storage owner")
}

func TestDeleteStorage_InvalidParamaters(t *testing.T) {
	assert.NotNil(t, DeleteStorage(host, user, "", ssl), "expected error for empty storage name")
}

func TestCreateAndDeleteStorage(t *testing.T) {
	cs := adminConnectionString(t)

	err := CreateStorage(host, user, testDBName, ssl)
	assert.Nil(t, err)

	// Test DB has been created
	db, err := sql.Open("postgres", cs)
	common.FatalIfError(t, err, "opening db connection")
	defer nonReturningCloseDB(db)
	var data string
	err = db.QueryRow("select datname from pg_database where datname = $1", testDBName).Scan(&data)
	common.FatalIfError(t, err, "scanning db name query for data")
	assert.Equal(t, testDBName, data)

	// TODO: Check that tables have been created
	//SELECT table_schema,table_name FROM information_schema.tables WHERE table_schema = 'public';

	err = DeleteStorage(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "deleting storage")

	// Test DB no longer exists
	err = db.QueryRow("select datname from pg_database where datname = $1", testDBName).Scan(&data)
	assert.Equal(t, sql.ErrNoRows, err)
}

func Test_createTestDB(t *testing.T) {
	db := createTestDB(t)
	if !assert.NotNil(t, db, `Unable to prepare DB for testing`) {
		t.Fail()
	}
	nonReturningCloseStorage(db)
	deleteTestDB(t)
}

//todo createTestDB should be given a base name for a db, which it should append a timestamp onto.
// createTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occurred whilst preparing the connection.
func createTestDB(t *testing.T) storage.Storage {
	err := CreateStorage(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "Error creating storage ")
	var cs string
	cs, err = NewConnectionString(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "Error creating connection string for storage access")
	db, err := New(cs)
	common.FatalIfError(t, err, "Error creating DB connection")
	return db
}

func deleteTestDBIgnorantly(t *testing.T) {
	_ = DeleteStorage(host, user, testDBName, ssl)
}

func deleteTestDB(t *testing.T) {
	err := DeleteStorage(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "Error deleting storage")
}

func Test_isAvailable(t *testing.T) {
	unavailableStore, _ := New("INVALID CONNECTION STRING")
	assert.False(t, unavailableStore.Available(), "Storage should not be available")
	availableStore := createTestDB(t)
	assert.True(t, availableStore.Available(), "Available returned false when it should have been true.")
	nonReturningCloseStorage(availableStore)
	deleteTestDB(t)
}

func adminConnectionString(t *testing.T) string {
	cs, err := NewConnectionString(host, user, "", ssl)
	common.FatalIfError(t, err, "generating new admin connection string")
	return cs
}
