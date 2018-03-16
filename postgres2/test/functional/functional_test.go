package functional

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres2"
	"github.com/glynternet/go-accounting-storage/test"
	"github.com/glynternet/go-money/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	// viper keys
	keyDBHost    = "db-host"
	keyDBUser    = "db-user"
	keyDBName    = "db-name"
	keyDBSSLMode = "db-sslmode"
)

func init() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	const retries = 5
	errs := make([]error, retries)
	var i int
	for i = 0; i < retries; i++ {
		err := postgres2.CreateStorage(
			viper.GetString(keyDBHost),
			viper.GetString(keyDBUser),
			viper.GetString(keyDBName),
			viper.GetString(keyDBSSLMode),
		)
		if err == nil {
			break
		}
		errs[i] = err
		time.Sleep(time.Second)
	}
	if errs[retries-1] != nil {
		for i, err := range errs {
			fmt.Printf("[retry: %02d] %v\n", i, err)
		}
	}
}

func TestSuite(t *testing.T) {
	store := createStorage(t)
	test.Test(t, store)
}

func createStorage(t *testing.T) storage.Storage {
	cs, err := postgres2.NewConnectionString(
		viper.GetString(keyDBHost),
		viper.GetString(keyDBUser),
		viper.GetString(keyDBName),
		viper.GetString(keyDBSSLMode),
	)
	common.FatalIfError(t, err, "creating connection string")
	store, err := postgres2.New(cs)
	common.FatalIfError(t, err, "creating storage")
	if !assert.True(t, store.Available(), "store should be available") {
		t.FailNow()
	}
	return store
}
