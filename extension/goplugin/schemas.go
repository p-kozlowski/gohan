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

package goplugin

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cloudwan/gohan/db/transaction"
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/schema"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/twinj/uuid"
	"strings"
)

var (
	NotPointerErr = fmt.Errorf("passed resource should be a pointer to concrete struct")
)

func isPointer(resource interface{}) bool {
	rv := reflect.ValueOf(resource)
	return rv.Kind() == reflect.Ptr
}

// Schemas in an implementation of ISchemas
type Schemas struct {
	environment *Environment
}

// List returns a list of loaded schemas
func (thisSchemas *Schemas) List() []goext.ISchema {
	manager := schema.GetManager()
	result := []goext.ISchema{}
	for _, rawSchema := range manager.OrderedSchemas() {
		result = append(result, NewSchema(thisSchemas.environment, rawSchema))
	}
	return result
}

// Find returns a schema by id or nil if not found
func (thisSchemas *Schemas) Find(id string) goext.ISchema {
	manager := schema.GetManager()
	schema, ok := manager.Schema(id)

	if !ok {
		log.Warning("cannot find schema '%s'", id)
		return nil
	}

	return NewSchema(thisSchemas.environment, schema)
}

// Environment returns the parent environment
func (thisSchemas *Schemas) Environment() goext.IEnvironment {
	return thisSchemas.environment
}

// NewSchemas allocates a new Schemas
func NewSchemas(environment *Environment) goext.ISchemas {
	return &Schemas{environment: environment}
}

// Schema is an implementation of ISchema
type Schema struct {
	environment *Environment
	rawSchema   *schema.Schema
}

// ID returns ID of schema
func (thisSchema *Schema) ID() string {
	return thisSchema.rawSchema.ID
}

func (thisSchema *Schema) structToMap(resource interface{}) map[string]interface{} {
	fieldsMap := map[string]interface{}{}

	mapper := reflectx.NewMapper("db")

	for _, property := range thisSchema.rawSchema.Properties {
		field := property.ID
		v := mapper.FieldByName(reflect.ValueOf(resource), property.ID)
		val := v.Interface()
		if field == "id" && v.String() == "" {
			id := uuid.NewV4().String()
			fieldsMap[field] = id
			v.SetString(id)
		} else if strings.Contains(v.Type().String(), "goext.Null") {
			valid := v.Field(1).Bool()
			if valid {
				fieldsMap[field] = v.Field(0).Interface()
			} else {
				fieldsMap[field] = nil
			}
		} else {
			fieldsMap[field] = val
		}
	}

	return fieldsMap
}

func (thisSchema *Schema) structToResource(resource interface{}) (*schema.Resource, error) {
	fieldsMap := thisSchema.structToMap(resource)
	return schema.NewResource(thisSchema.rawSchema, fieldsMap)
}

func (thisSchema *Schema) assignField(name string, field reflect.Value, value interface{}) error {
	if field.Kind() == reflect.Struct || field.Kind() == reflect.Slice {
		mapJSON, err := json.Marshal(value)
		if err != nil {
			return err
		}
		newField := reflect.New(field.Type())
		fieldJSON := string(mapJSON)
		fieldInterface := newField.Interface()
		err = json.Unmarshal([]byte(fieldJSON), &fieldInterface)
		if err != nil {
			return err
		}
		field.Set(newField.Elem())
	} else {
		setValue(field, reflect.ValueOf(value))
	}
	return nil
}

// List lists schema resources
func (thisSchema *Schema) List(resources interface{}, filter goext.Filter, paginator *goext.Paginator, context goext.Context) error {
	if !isPointer(resources) {
		return NotPointerErr
	}
	slicePtrValue := reflect.ValueOf(resources)
	slicePtrType := reflect.TypeOf(resources)
	sliceValue := slicePtrValue.Elem()
	sliceType := slicePtrType.Elem()
	elemType := sliceType.Elem()

	sliceValue.SetLen(0)

	if context == nil {
		context = goext.MakeContext()
	}

	tx, hasOpenTransaction := contextGetTransaction(context)
	if !hasOpenTransaction {
		var err error
		tx, err = thisSchema.environment.Database().Begin()

		if err != nil {
			return err
		}

		defer tx.Close()
	}

	data, _, err := tx.List(thisSchema, filter, nil, paginator)

	if err != nil {
		return err
	}

	mapper := reflectx.NewMapper("db")

	for i := 0; i < len(data); i++ {
		resource := reflect.New(elemType)
		mapped := mapper.FieldMap(resource)

		for name, field := range mapped {
			value := data[i][name]
			if err := thisSchema.assignField(name, field, value); err != nil {
				return err
			}
		}

		sliceValue.Set(reflect.Append(sliceValue, resource.Elem()))
	}

	return nil
}

// LockList locks and returns resources
func (thisSchema *Schema) LockList(resources interface{}, filter goext.Filter, paginator *goext.Paginator, context goext.Context, policy goext.LockPolicy) error {
	if !isPointer(resources) {
		return NotPointerErr
	}
	//TODO: implement proper locking
	return thisSchema.List(resources, filter, paginator, context)
}

// FetchRelated fetches related resources
func (thisSchema *Schema) FetchRelated(resource interface{}, relatedResource interface{}, context goext.Context) error {
	if !isPointer(resource) {
		return NotPointerErr
	}
	for _, property := range thisSchema.rawSchema.Properties {
		if property.Relation != "" {
			relatedSchema, ok := schema.GetManager().Schema(property.Relation)

			if !ok {
				return fmt.Errorf("Could not get related schema: %s for: %s", property.Relation, thisSchema.rawSchema.ID)
			}

			mapper := reflectx.NewMapper("db")
			id := mapper.FieldByName(reflect.ValueOf(resource), property.ID).String()

			NewSchema(thisSchema.environment, relatedSchema).Fetch(id, relatedResource, context)

			return nil
		}
	}

	return nil
}

// Fetch fetches a resource by ID
func (thisSchema *Schema) Fetch(id string, res interface{}, context goext.Context) error {
	if !isPointer(res) {
		return NotPointerErr
	}
	if context == nil {
		context = goext.MakeContext()
	}
	tx, hasOpenTransaction := contextGetTransaction(context)
	if !hasOpenTransaction {
		var err error
		tx, err = thisSchema.environment.Database().Begin()

		if err != nil {
			return err
		}

		defer tx.Close()

		contextSetTransaction(context, tx)
	}

	filter := goext.Filter{"id": id}

	data, err := tx.Fetch(thisSchema, filter)

	if err != nil {
		return err
	}
	resourceType, ok := GlobResourceTypes[thisSchema.rawSchema.ID]
	if !ok {
		return fmt.Errorf("No type registered for schema title: %s", thisSchema.rawSchema.ID)
	}
	resource := reflect.ValueOf(res)

	for i := 0; i < resourceType.NumField(); i++ {
		field := resource.Elem().Field(i)

		fieldType := resourceType.Field(i)
		propertyName := fieldType.Tag.Get("db")
		property, err := thisSchema.rawSchema.GetPropertyByID(propertyName)
		if err != nil {
			return err
		}
		value := data[property.ID]
		thisSchema.assignField(propertyName, field, value)
	}

	return nil
}

func setValue(field, value reflect.Value) {
	if value.IsValid() {
		if value.Type() != field.Type() && field.Kind() == reflect.Slice { // empty slice has type []interface{}
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		} else {
			field.Set(value)
		}
	}
}

// LockFetch locks and fetches resource by ID
func (thisSchema *Schema) LockFetch(id string, res interface{}, context goext.Context, policy goext.LockPolicy) error {
	return thisSchema.Fetch(id, res, context)
}

type TransactionData struct {
	Tx    transaction.Transaction
	Depth int
}

// Create creates a resource
func (thisSchema *Schema) Create(resource interface{}, context goext.Context) error {
	if !isPointer(resource) {
		return NotPointerErr
	}

	if context == nil {
		context = goext.MakeContext()
	}
	tx, hasOpenTransaction := contextGetTransaction(context)
	if hasOpenTransaction {
		contextCopy := goext.MakeContext().
			WithSchemaID(thisSchema.ID()).
			WithResource(thisSchema.structToMap(resource))
		contextSetTransaction(contextCopy, tx)
		return thisSchema.createInTransaction(resource, contextCopy, tx)
	}

	context.WithSchemaID(thisSchema.ID()).
		WithResource(thisSchema.structToMap(resource))

	if err := thisSchema.environment.HandleEvent(goext.PreCreate, context); err != nil {
		return err
	}

	tx, err := thisSchema.environment.Database().Begin()
	if err != nil {
		return err
	}
	defer tx.Close()
	contextSetTransaction(context, tx)

	if err = thisSchema.environment.HandleEvent(goext.PreCreateTx, context); err != nil {
		return err
	}

	if err = tx.Create(thisSchema, context["resource"].(map[string]interface{})); err != nil {
		return err
	}

	if err = thisSchema.environment.updateResourceFromContext(resource, context); err != nil {
		return err
	}

	if err = thisSchema.environment.HandleEvent(goext.PostCreateTx, context); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	if err = tx.Close(); err != nil {
		return err
	}

	return thisSchema.environment.HandleEvent(goext.PostCreate, context)
}

func (thisSchema *Schema) createInTransaction(resource interface{}, context goext.Context, tx goext.ITransaction) error {
	var err error

	if err := thisSchema.environment.HandleEvent(goext.PreCreate, context); err != nil {
		return err
	}

	if err = thisSchema.environment.HandleEvent(goext.PreCreateTx, context); err != nil {
		return err
	}

	if err = tx.Create(thisSchema, context["resource"].(map[string]interface{})); err != nil {
		return err
	}

	if err = thisSchema.environment.updateResourceFromContext(resource, context); err != nil {
		return err
	}

	if err = thisSchema.environment.HandleEvent(goext.PostCreateTx, context); err != nil {
		return err
	}

	return thisSchema.environment.HandleEvent(goext.PostCreate, context)
}

// Update updates a resource
func (thisSchema *Schema) Update(resource interface{}, context goext.Context) error {
	if !isPointer(resource) {
		return NotPointerErr
	}

	var tx goext.ITransaction
	var resourceData *schema.Resource
	var err error

	if resourceData, err = thisSchema.structToResource(resource); err != nil {
		return err
	}

	if context == nil {
		context = goext.MakeContext()
	}

	contextCopy := goext.MakeContext()
	for k, v := range context {
		contextCopy[k] = v
	}

	contextCopy.WithResource(thisSchema.structToMap(resource)).
		WithResourceID(resourceData.ID()).
		WithSchemaID(thisSchema.ID())

	if err = thisSchema.environment.HandleEvent(goext.PreUpdate, contextCopy); err != nil {
		return err
	}

	tx, hasOpenTransaction := contextGetTransaction(contextCopy)
	if !hasOpenTransaction {
		if tx, err = thisSchema.environment.Database().Begin(); err != nil {
			return err
		}

		defer tx.Close()
		contextSetTransaction(contextCopy, tx)
		contextSetTransaction(context, tx)
	}

	if err = thisSchema.environment.HandleEvent(goext.PreUpdateTx, contextCopy); err != nil {
		return err
	}

	if err = tx.Update(thisSchema, contextCopy["resource"].(map[string]interface{})); err != nil {
		return err
	}

	if err = thisSchema.environment.updateResourceFromContext(resource, contextCopy); err != nil {
		return err
	}

	if err = thisSchema.environment.HandleEvent(goext.PostUpdateTx, contextCopy); err != nil {
		return err
	}

	if !hasOpenTransaction {
		if err = tx.Commit(); err != nil {
			return err
		}
	}

	return thisSchema.environment.HandleEvent(goext.PostUpdate, contextCopy)
}

// Delete deletes resource by ID
func (thisSchema *Schema) Delete(filter goext.Filter, context goext.Context) error {
	var tx goext.ITransaction
	var err error
	if context == nil {
		context = goext.MakeContext()
	}
	tx, hasOpenTransaction := contextGetTransaction(context)
	if !hasOpenTransaction {
		if tx, err = thisSchema.environment.Database().Begin(); err != nil {
			return err
		}

		defer tx.Close()

		contextSetTransaction(context, tx)
	}
	contextTx := goext.MakeContext()
	contextSetTransaction(contextTx, tx)

	resources := reflect.Zero(reflect.SliceOf(GlobResourceTypes[thisSchema.ID()]))
	x := reflect.New(resources.Type())
	x.Elem().Set(resources)
	if err := thisSchema.LockList(x.Interface(), filter, nil, contextTx, goext.LockRelatedResources); err != nil {
		return err
	}

	x = x.Elem()
	mapper := reflectx.NewMapper("db")
	for i := 0; i < x.Len(); i++ {
		resource := x.Index(i)
		resourceID := mapper.FieldByName(resource, "id").Interface()

		contextTx = contextTx.WithResource(thisSchema.structToMap(resource.Addr().Interface())).
			WithSchemaID(thisSchema.ID())

		if err = thisSchema.environment.HandleEvent(goext.PreDelete, contextTx); err != nil {
			return err
		}

		if err = thisSchema.environment.HandleEvent(goext.PreDeleteTx, contextTx); err != nil {
			return err
		}

		if err = tx.Delete(thisSchema, resourceID); err != nil {
			return err
		}

		if err = thisSchema.environment.HandleEvent(goext.PostDeleteTx, contextTx); err != nil {
			return err
		}

		if err = thisSchema.environment.HandleEvent(goext.PostDelete, contextTx); err != nil {
			return err
		}
	}

	if !hasOpenTransaction {
		tx.Commit()
	}

	return nil
}

// RegisterEventHandler registers a schema handler
func (thisSchema *Schema) RegisterEventHandler(event string, handler func(context goext.Context, resource goext.Resource, environment goext.IEnvironment) error, priority goext.Priority) {
	thisSchema.environment.RegisterSchemaEventHandler(thisSchema.rawSchema.ID, event, handler, priority)
}

// RegisterResourceType registers a runtime type for a resource
func (thisSchema *Schema) RegisterResourceType(typeValue interface{}) {
	thisSchema.environment.RegisterResourceType(thisSchema.rawSchema.ID, typeValue)
}

// Environment returns the parent environment
func (thisSchema *Schema) Environment() goext.IEnvironment {
	return thisSchema.environment
}

// NewSchema allocates a new Schema
func NewSchema(environment *Environment, rawSchema *schema.Schema) goext.ISchema {
	return &Schema{environment: environment, rawSchema: rawSchema}
}

func contextSetTransaction(ctx goext.Context, tx goext.ITransaction) goext.Context {
	ctx["transaction"] = tx
	return ctx
}

func contextGetTransaction(ctx goext.Context) (goext.ITransaction, bool) {
	ctxTx := ctx["transaction"]
	if ctxTx == nil {
		return nil, false
	}
	return ctxTx.(goext.ITransaction), true
}
