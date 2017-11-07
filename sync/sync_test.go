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

package sync_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/cloudwan/gohan/sync"
	"github.com/cloudwan/gohan/sync/etcdv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sync", func() {
	const (
		etcdServer = "localhost:2379"
	)
	var (
		syn *etcdv3.Sync
	)
	BeforeEach(func() {
		var err error
		syn, err = etcdv3.NewSync([]string{etcdServer}, time.Second)
		Expect(err).To(Succeed())
		Expect(syn.Delete("/", false)).To(Succeed())
		Expect(syn.Update("/some-key", "{\"some-json-key\":\"some-json-value\"}")).To(Succeed())
	})
	AfterEach(func() {
		syn.Close()
	})
	Describe("Goroutine leak tests", func() {
		const (
			callCount       = 1000
			maxExtraThreads = 20 // note: even 8 should be sufficient but 20 is safer
		)
		var (
			gstart, gend int
		)
		sleepUntilAllGoroutinesAreSwept := func() {
			// this function sleeps at most 30 secs; it checks in one-second intervals
			// if the number of goroutines is still decreasing which means that Go runtime
			// is still sweeping finished goroutines
			times := 0
			glast := runtime.NumGoroutine()
			for {
				time.Sleep(time.Second)
				times++
				if times == 30 {
					break
				}
				gnow := runtime.NumGoroutine()
				if gnow >= glast {
					break
				}
				glast = gnow
			}
		}
		BeforeEach(func() {
			var err error
			syn, err = etcdv3.NewSync([]string{etcdServer}, time.Second)
			Expect(err).To(Succeed())
			Expect(syn.Delete("/", false)).To(Succeed())
			gstart = runtime.NumGoroutine()
		})
		AfterEach(func() {
			gend = runtime.NumGoroutine()
			Expect(gend - gstart).To(BeNumerically("<=", maxExtraThreads))
		})
		It("Fetch should not leak goroutines", func() {
			for i := 0; i < callCount; i++ {
				_, err := syn.Fetch("/some-key")
				Expect(err).To(Succeed())
			}
		})
		It("Update should not leak goroutines", func() {
			for i := 0; i < callCount; i++ {
				err := syn.Update("/some-key", fmt.Sprintf("{\"some-json-key\":\"some-json-value-%d\"}", i))
				Expect(err).To(Succeed())
			}
		})
		It("Delete should not leak goroutines", func() {
			for i := 0; i < callCount; i++ {
				err := syn.Delete("/some-key", false)
				Expect(err).To(Succeed())
			}
		})
		It("Watch should not leak goroutines", func() {
			for i := 0; i < callCount; i++ {
				func() {
					eventCh := make(chan *sync.Event, 32)
					defer close(eventCh)
					stopCh := make(chan bool)
					defer close(stopCh)
					errCh := make(chan error, 1)
					defer close(errCh)
					go func() {
						errCh <- syn.Watch("/some-key", eventCh, stopCh, sync.RevisionCurrent)
					}()
					stopCh <- true
					Expect(<-errCh).To(Succeed())
				}()
			}
			// wait for all goroutines to end
			//
			// this sleep is required to ensure that goroutines leave their function after
			// watch is stopped; it is impossible to add synchronization here since
			// goroutines are swept by runtime package; under all normal circumstances this
			// should happen very fast since there is no real code being executed during this
			// wait period
			sleepUntilAllGoroutinesAreSwept()
		})
		It("WatchContext should not leak goroutines", func() {
			for i := 0; i < callCount; i++ {
				func() {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					_, err := syn.WatchContext(ctx, "/some-key", int64(sync.RevisionCurrent))
					Expect(err).To(Succeed())
				}()
			}
			// wait for all goroutines to end
			//
			// this sleep is required to ensure that goroutines leave their function after
			// context is cancelled; it is impossible to add synchronization here since
			// goroutines are swept by runtime package; under all normal circumstances this
			// should happen very fast since there is no real code being executed during this
			// wait period
			sleepUntilAllGoroutinesAreSwept()
		})
	})
})

func TestSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Golang sync suite")
}
