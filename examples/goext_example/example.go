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

package main

import (
	"github.com/cloudwan/gohan/examples/goext_example/todo"
	"github.com/cloudwan/gohan/extension/goext"
)

// Schemas returns a list of schemas that will be loaded with this extension
func Schemas() []string {
	return []string{
		"todo/entry.yaml",
		"todo/link.yaml",
	}
}

// Init is an entry point of this extension
func Init(env goext.IEnvironment) error {
	// Find the pre-loaded schema
	schema := env.Schemas().Find("entry")

	// Register runtime type for the schema
	schema.RegisterResourceType(todo.Entry{})

	// Register schema handler
	schema.RegisterEventHandler(goext.PreUpdate, func(ctx goext.Context, res goext.Resource, env goext.IEnvironment) error {
		env.Logger().Infof("Called resource pre-update handler")

		// Cast resource to its runtime type
		entry := res.(*todo.Entry)

		// Modify resource
		entry.Name = "name changed in pre_update event"
		env.Logger().Infof("Modified Todo resource in pre-update handler: %v", entry)

		return nil
	}, goext.PriorityDefault)

	return nil
}
