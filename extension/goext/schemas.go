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

// LockPolicy indicates lock policy
type LockPolicy int

const (
	// LockRelatedResources indicates that related resources are also locked
	LockRelatedResources LockPolicy = iota

	// SkipRelatedResources indicates that related resources are not locked
	SkipRelatedResources

	// NoLock indicates that no lock is acquired at all
	NoLock
)

// Resource represents a resource
type Resource interface{}

// Resources is a list of resources
type Resources []Resource

// Context represents a context of a handler
type Context map[string]interface{}

// Filter represents filtering options for fetching functions
type Filter map[string]interface{}

// Paginator represents a paginator
type Paginator struct {
	Key    string
	Order  string
	Limit  uint64
	Offset uint64
}

// MakeContext creates an empty context
func MakeContext() Context {
	return make(map[string]interface{})
}

// WithSchemaID appends schema ID to given context
func (ctx Context) WithSchemaID(schemaID string) Context {
	ctx["schema_id"] = schemaID
	return ctx
}

// WithISchema appends ISchema to given context
func (ctx Context) WithISchema(schema ISchema) Context {
	ctx["schema"] = schema
	return ctx
}

// WithResource appends resource to given context
func (ctx Context) WithResource(resource Resource) Context {
	ctx["resource"] = resource
	return ctx
}

// WithResourceID appends resource ID to given context
func (ctx Context) WithResourceID(resourceID string) Context {
	ctx["id"] = resourceID
	return ctx
}

// WithTransaction appends transaction to given context
func (ctx Context) WithTransaction(tx ITransaction) Context {
	ctx["transaction"] = tx
	return ctx
}

// Clone returns copy of context
func (ctx Context) Clone() Context {
	contextCopy := MakeContext()
	for k, v := range ctx {
		contextCopy[k] = v
	}
	return contextCopy
}

// Priority represents handler priority; can be negative
type Priority = int

// PriorityDefault is a default handler priority
const PriorityDefault Priority = 0

// ISchema is an interface representing a single schema in Gohan
type ISchema interface {
	IEnvironmentRef

	// properties
	ID() string

	// database
	List(filter Filter, paginator *Paginator, context Context) ([]interface{}, error)
	ListRaw(rawResources interface{}, filter Filter, paginator *Paginator, context Context) error

	LockList(filter Filter, paginator *Paginator, context Context, lockPolicy LockPolicy) ([]interface{}, error)
	LockListRaw(rawResources interface{}, filter Filter, paginator *Paginator, context Context, lockPolicy LockPolicy) error

	Fetch(id string, context Context) (interface{}, error)
	FetchRaw(id string, rawResource interface{}, context Context) error

	LockFetchRaw(id string, rawResource interface{}, context Context, lockPolicy LockPolicy) error
	FetchRelatedRaw(rawResource interface{}, relatedResource interface{}, context Context) error

	CreateRaw(rawResource interface{}, context Context) error

	UpdateRaw(rawResource interface{}, context Context) error
	DbUpdateRaw(rawResource interface{}, context Context) error

	DeleteRaw(filter Filter, context Context) error

	// events
	RegisterEventHandler(event string, handler func(context Context, resource Resource, environment IEnvironment) error, priority Priority)

	// runtime types
	RegisterType(resourceType interface{})
	RegisterRawType(rawResourceType interface{})
}

// ISchemas is an interface to schemas manager in Gohan
type ISchemas interface {
	IEnvironmentRef

	List() []ISchema
	Find(id string) ISchema
}
