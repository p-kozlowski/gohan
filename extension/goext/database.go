// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goext

import "fmt"

// IDatabase is an interface to database in Gohan
type IDatabase interface {
	Begin() (ITransaction, error)
	BeginTx(ctx Context, options *TxOptions) (ITransaction, error)
}

func withinImpl(context Context, txBegin func() (ITransaction, error), fn func(tx ITransaction) error) (err error) {
	rawTx, hasOwner := context["transaction"]
	var tx ITransaction

	if hasOwner {
		tx = rawTx.(ITransaction)
	} else {
		tx, err = txBegin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %s", err)
		}
		context["transaction"] = tx
	}

	defer func() {
		if !hasOwner {
			if err == nil {
				err = tx.Commit()
			} else if !tx.Closed() {
				tx.Close()
			}
			delete(context, "transaction")
		}
	}()

	return fn(tx)
}

// Within calls a function in scoped transaction
func Within(env IEnvironment, context Context, fn func(tx ITransaction) error) error {
	return withinImpl(context, func() (ITransaction, error) {
		return env.Database().Begin()
	}, fn)
}

// WithinTx calls a function in scoped transaction with options
func WithinTx(env IEnvironment, ctx Context, options *TxOptions, fn func(tx ITransaction) error) error {
	return withinImpl(ctx, func() (ITransaction, error) {
		return env.Database().BeginTx(ctx, options)
	}, fn)
}
