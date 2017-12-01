package postgres2_test

import (
	"testing"

	"github.com/glynternet/go-accounting-storage"
	postgres "github.com/glynternet/go-accounting-storage/postgres2"
	"github.com/glynternet/go-accounting/common"
)

func prepareTestDB(t *testing.T) storage.Storage {
	cs, err := postgres.NewConnectionString(host, user, realDBName, ssl)
	common.FatalIfError(t, err, "creating connection string")
	store, err := postgres.New(cs)
	common.FatalIfError(t, err, "connecting to postgres store")
	return store
}

//todo
