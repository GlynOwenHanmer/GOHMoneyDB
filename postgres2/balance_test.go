package postgres2_test

import (
	"testing"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres2"
	"github.com/glynternet/go-accounting/common"
)

func prepareTestDB(t *testing.T) storage.Storage {
	cs, err := postgres2.NewConnectionString(host, user, testDBName, ssl)
	common.FatalIfError(t, err, "creating connection string")
	store, err := postgres2.New(cs)
	common.FatalIfError(t, err, "connecting to postgres store")
	return store
}

//todo
