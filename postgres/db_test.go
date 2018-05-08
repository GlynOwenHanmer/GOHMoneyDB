// +build integration

package postgres_test

import (
	"io"
	"strings"
	"testing"

	"fmt"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

const (
	host       = "localhost"
	realDBName = "money"
	testDBName = realDBName
	user       = "glynhanmer"
	ssl        = "disable"
)

func TestNewConnectionString(t *testing.T) {
	c := postgres.NewConnectionString("", "name", "", "")
	assert.NotNil(t, c)
	assert.Equal(t, "user=name", c)

	c = postgres.NewConnectionString("localhost", "user", "dbname", "disable")
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

//todo prepareTestDB should be given a base name for a db, which it should append a timestamp onto.
// prepareTestDB prepares a DB connection to the test DB and return it, if possible, with any errors that occurred whilst preparing the connection.
func prepareTestDB(t *testing.T) storage.Storage {
	cs := postgres.NewConnectionString(host, user, testDBName, ssl)
	db, err := postgres.New(cs)
	common.FatalIfError(t, err, "Error creating DB connection")
	return db
}

func Test_isAvailable(t *testing.T) {
	unavailableDb, _ := postgres.New("INVALID CONNECTION STRING")
	assert.False(t, unavailableDb.Available(), "Storage should not be available")
	availableDb := prepareTestDB(t)
	assert.True(t, availableDb.Available(), "Available returned false when it should have been true.")
	nonReturningClose(t, availableDb, "availableDb")
}

func nonReturningClose(t *testing.T, c io.Closer, name string) {
	var nameInsert string
	if name != "" {
		nameInsert = fmt.Sprintf("(%s) ", name)
	}
	if c == nil {
		t.Errorf("Attempted to close io.Closer %sbut it was nil.", nameInsert)
		return
	}
	common.FatalIfErrorf(t, c.Close(), "Error closing io.Closer %s%v", nameInsert, c)
}

func nonReturningCloseStorage(t *testing.T, s storage.Storage) {
	if s == nil {
		t.Errorf("Attempted to close Storage but it was nil.")
		return
	}
	nonReturningClose(t, s, "Storage")
}
