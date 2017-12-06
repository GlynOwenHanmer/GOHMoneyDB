package main

import (
	"log"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres"
	"github.com/glynternet/go-accounting-storage/postgres2"
)

const (
	host      = "localhost"
	oldDBName = "money"
	newDBName = "moneyv2"
	user      = "glynhanmer"
	ssl       = "disable"
)

type AccountBalances struct {
	storage.Account
	storage.Balances
}

func main() {
	oldCS, err := postgres.NewConnectionString(host, user, oldDBName, ssl)
	if err != nil {
		log.Fatalf("could not create connection string: %v", err)
	}
	old, err := postgres.New(oldCS)
	if err != nil {
		log.Fatalf("could not create old store: %v", err)
	}
	oldAs, err := old.SelectAccounts()
	if err != nil {
		log.Fatalf("could not select old accounts: %v", err)
	}
	_ = postgres2.DeleteStorage(host, user, newDBName, ssl)
	//adminCS, err := postgres2.NewConnectionString(host, user, "", ssl)
	err = postgres2.CreateStorage(host, user, newDBName, ssl)
	if err != nil {
		log.Fatalf("error creating new DB: %v", err)
	}
	defer postgres2.DeleteStorage(host, user, newDBName, ssl)
	newCS, err := postgres.NewConnectionString(host, user, newDBName, ssl)
	if err != nil {
		log.Fatalf("could not create connection string: %v", err)
	}
	var oldAbs []AccountBalances
	for _, oa := range *oldAs {
		bs, err := old.SelectAccountBalances(oa)
		if err != nil {
			log.Fatalf("error selecting balances for account %v: %v", oa, err)
		}
		oldAbs = append(oldAbs, AccountBalances{Account: oa, Balances: *bs})
	}

	new, err := postgres2.New(newCS)
	if err != nil {
		log.Fatalf("error creating new store: %v", err)
	}
	var newAbs []AccountBalances
	for _, ab := range oldAbs {
		newA, err := new.InsertAccount(ab.Account.Account)
		if err != nil {
			log.Fatalf("could not insert account: %v", err)
		}
		var newBs storage.Balances
		for _, b := range ab.Balances {
			newB, err := new.InsertBalance(*newA, b.Balance)
			if err != nil {
				log.Fatalf("error inserting new balance: %v", err)
			}
			newBs = append(newBs, *newB)
		}
		newAbs = append(newAbs, AccountBalances{Account: *newA, Balances: newBs})
	}
	newAs, err := new.SelectAccounts()
	if err != nil {
		log.Fatalf("error selecting new accounts: %v", err)
	}
	newlen := len(*newAs)
	oldLen := len(*oldAs)
	if newlen != oldLen {
		log.Fatalf("accounts length not equal. old - %d, new - %d", oldLen, newlen)
	}
	for i, a := range *oldAs {
		if !a.Account.Equal([]storage.Account(*newAs)[i].Account) {
			log.Fatal("accounts not equals")
		}
	}
	//_ = postgres2.DeleteStorage(host, user, newDBName, ssl)
}
