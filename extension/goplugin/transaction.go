package goplugin

import (
	"github.com/cloudwan/gohan/db/pagination"
	"github.com/cloudwan/gohan/db/transaction"
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/schema"
)

//Transaction is common interface for handling transaction
type Transaction struct {
	tx transaction.Transaction
}

func (t *Transaction) findRawSchema(id string) *schema.Schema {
	manager := schema.GetManager()
	schema, ok := manager.Schema(id)

	if !ok {
		log.Warning("cannot find schema '%s'", id)
		return nil
	}
	return schema
}

func (t *Transaction) Create(s goext.ISchema, resource goext.RawResource) error {
	res, err := schema.NewResource(t.findRawSchema(s.ID()), resource)
	if err != nil {
		return err
	}
	return t.tx.Create(res)
}

func (t *Transaction) Update(s goext.ISchema, resource goext.RawResource) error {
	res, err := schema.NewResource(t.findRawSchema(s.ID()), resource)
	if err != nil {
		return err
	}
	return t.tx.Update(res)
}

func mapGoExtResourceState(resourceState *goext.ResourceState) *transaction.ResourceState {
	if resourceState == nil {
		return nil
	}
	return &transaction.ResourceState{
		ConfigVersion: resourceState.ConfigVersion,
		StateVersion:  resourceState.StateVersion,
		Error:         resourceState.Error,
		State:         resourceState.State,
		Monitoring:    resourceState.Monitoring,
	}
}

func mapTransactionResourceState(resourceState transaction.ResourceState) goext.ResourceState {
	return goext.ResourceState{
		ConfigVersion: resourceState.ConfigVersion,
		StateVersion:  resourceState.StateVersion,
		Error:         resourceState.Error,
		State:         resourceState.State,
		Monitoring:    resourceState.Monitoring,
	}
}

func (t *Transaction) StateUpdate(s goext.ISchema, resource goext.RawResource, resourceState *goext.ResourceState) error {
	res, err := schema.NewResource(t.findRawSchema(s.ID()), resource)
	if err != nil {
		return err
	}
	return t.tx.StateUpdate(res, mapGoExtResourceState(resourceState))
}
func (t *Transaction) Delete(schema goext.ISchema, resourceID interface{}) error {
	return t.tx.Delete(t.findRawSchema(schema.ID()), resourceID)
}
func (t *Transaction) Fetch(schema goext.ISchema, filter goext.Filter) (goext.RawResource, error) {
	res, err := t.tx.Fetch(t.findRawSchema(schema.ID()), transaction.Filter(filter))
	if err != nil {
		return nil, err
	}
	return res.Data(), nil
}
func (t *Transaction) LockFetch(schema goext.ISchema, filter goext.Filter, lockPolicy goext.LockPolicy) (goext.RawResource, error) {
	//TODO: implement proper locking
	return t.Fetch(schema, filter)
}
func (t *Transaction) StateFetch(schema goext.ISchema, filter goext.Filter) (goext.ResourceState, error) {
	transactionResourceState, err := t.tx.StateFetch(t.findRawSchema(schema.ID()), transaction.Filter(filter))
	if err != nil {
		return goext.ResourceState{}, err
	}
	return mapTransactionResourceState(transactionResourceState), err
}
func (t *Transaction) List(schema goext.ISchema, filter goext.Filter, listOptions *goext.ListOptions, paginator *goext.Paginator) ([]goext.RawResource, uint64, error) {
	schemaID := schema.ID()

	data, _, err := t.tx.List(t.findRawSchema(schemaID), transaction.Filter(filter), nil, (*pagination.Paginator)(paginator))
	if err != nil {
		return nil, 0, err
	}

	resourceProperties := make([]goext.RawResource, len(data))
	for i := range data {
		resourceProperties[i] = data[i].Data()
	}

	return resourceProperties, uint64(len(resourceProperties)), nil

}
func (t *Transaction) LockList(schema goext.ISchema, filter goext.Filter, listOptions *goext.ListOptions, paginator *goext.Paginator, lockingPolicy goext.LockPolicy) ([]goext.RawResource, uint64, error) {
	return t.List(schema, filter, listOptions, paginator)
}
func (t *Transaction) RawTransaction() interface{} {
	return t.RawTransaction()
}
func (t *Transaction) Query(schema goext.ISchema, query string, args []interface{}) (list []goext.RawResource, err error) {
	schemaID := schema.ID()

	data, err := t.tx.Query(t.findRawSchema(schemaID), query, args)
	if err != nil {
		return nil, err
	}

	resourceProperties := make([]goext.RawResource, len(data))
	for i := range data {
		resourceProperties[i] = data[i].Data()
	}

	return resourceProperties, nil
}
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}
func (t *Transaction) Exec(query string, args ...interface{}) error {
	return t.tx.Exec(query, args)
}
func (t *Transaction) Close() error {
	return t.tx.Close()
}
func (t *Transaction) Closed() bool {
	return t.tx.Closed()
}
func (t *Transaction) GetIsolationLevel() goext.Type {
	return goext.Type(t.tx.GetIsolationLevel())
}
