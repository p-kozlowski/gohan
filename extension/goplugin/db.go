package goplugin

import (
	"context"

	"github.com/cloudwan/gohan/db"
	"github.com/cloudwan/gohan/db/transaction"
	"github.com/cloudwan/gohan/extension/goext"
)

type Db struct {
	db db.DB
}

func NewDB(environment *Environment) goext.IDatabase {
	return &Db{db: environment.db}
}

func (db *Db) Begin() (goext.ITransaction, error) {
	t, _ := db.db.Begin()
	return &Transaction{t}, nil
}
func (db *Db) BeginTx(ctx goext.Context, options *goext.TxOptions) (goext.ITransaction, error) {
	opts := transaction.TxOptions{IsolationLevel: transaction.Type(options.IsolationLevel)}
	t, _ := db.db.BeginTx(context.Background(), &opts)
	return &Transaction{t}, nil
}
