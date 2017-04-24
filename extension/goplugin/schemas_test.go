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

package goplugin_test

import (
	"os"

	"github.com/cloudwan/gohan/db"
	"github.com/cloudwan/gohan/db/options"
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/extension/goplugin"
	"github.com/cloudwan/gohan/extension/goplugin/test_data/ext_good/test"
	"github.com/cloudwan/gohan/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {
	var (
		env *goplugin.Environment
	)

	const (
		DB_FILE     = "test.db"
		DB_TYPE     = "sqlite3"
		SCHEMA_PATH = "test_data/test_schema.yaml"
	)

	BeforeEach(func() {
		manager := schema.GetManager()
		Expect(manager.LoadSchemaFromFile(SCHEMA_PATH)).To(Succeed())
		db, err := db.ConnectDB(DB_TYPE, DB_FILE, db.DefaultMaxOpenConn, options.Default())
		Expect(err).To(BeNil())
		env = goplugin.NewEnvironment("test", db, nil, nil)
	})

	AfterEach(func() {
		os.Remove(DB_FILE)

	})

	Describe("CRUD", func() {
		var (
			testSchema      goext.ISchema
			createdResource test.Test
			context         goext.Context
		)

		BeforeEach(func() {
			loaded, err := env.Load("test_data/ext_good/ext_good.so", nil)
			Expect(loaded).To(BeTrue())
			Expect(err).To(BeNil())
			Expect(env.Start()).To(Succeed())
			testSchema = env.Schemas().Find("test")
			Expect(testSchema).To(Not(BeNil()))

			err = db.InitDBWithSchemas(DB_TYPE, DB_FILE, true, true, false)
			Expect(err).To(BeNil())

			tx, err := env.Database().Begin()
			Expect(err).To(BeNil())

			context = goext.MakeContext().WithTransaction(tx)

			createdResource = test.Test{
				ID:          "some-id",
				Description: "description",
			}
		})

		AfterEach(func() {
			env.Stop()
		})

		It("Should create resource", func() {
			Expect(testSchema.Create(&createdResource, context)).To(Succeed())
		})

		It("Lists previously created resource", func() {
			Expect(testSchema.Create(&createdResource, context)).To(Succeed())
			returnedResources := []test.Test{}
			err := testSchema.List(&returnedResources, goext.Filter{}, nil, context)
			Expect(err).To(BeNil())
			Expect(returnedResources).To(HaveLen(1))
			returnedResource := returnedResources[0]
			Expect(createdResource).To(Equal(returnedResource))
		})

		It("Fetch previously created resource", func() {
			Expect(testSchema.Create(&createdResource, context)).To(Succeed())
			returnedResource := test.Test{}
			Expect(testSchema.Fetch(createdResource.ID, &returnedResource, context)).To(Succeed())
			Expect(createdResource).To(Equal(returnedResource))
		})

		It("Delete previously created resource", func() {
			Expect(testSchema.Create(&createdResource, context)).To(Succeed())
			Expect(testSchema.Delete(goext.Filter{"id": createdResource.ID}, context)).To(Succeed())
			returnedResource := test.Test{}
			err := testSchema.Fetch(createdResource.ID, &returnedResource, context)
			Expect(err).To(HaveOccurred())
		})

		It("Update previously created resource", func() {
			Expect(testSchema.Create(&createdResource, context)).To(Succeed())
			createdResource.Description = "other-description"
			Expect(testSchema.Update(&createdResource, context)).To(Succeed())
			returnedResource := test.Test{}
			Expect(testSchema.Fetch(createdResource.ID, &returnedResource, context)).To(Succeed())
			Expect(returnedResource.Description).To(Equal("other-description"))
		})
	})
})
