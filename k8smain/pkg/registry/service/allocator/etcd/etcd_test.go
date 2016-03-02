/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcd

import (
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/registry/service/allocator"
	"k8s.io/kubernetes/pkg/storage/etcd/etcdtest"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"

	"golang.org/x/net/context"
)

func newStorage(t *testing.T) (*Etcd, *etcdtesting.EtcdTestServer, allocator.Interface) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	mem := allocator.NewAllocationMap(100, "rangeSpecValue")
	etcd := NewEtcd(mem, "/ranges/serviceips", api.Resource("serviceipallocations"), etcdStorage)
	return etcd, server, mem
}

func validNewRangeAllocation() *api.RangeAllocation {
	return &api.RangeAllocation{
		Range: "rangeSpecValue",
	}
}

func key() string {
	s := "/ranges/serviceips"
	return etcdtest.AddPrefix(s)
}

func TestEmpty(t *testing.T) {
	storage, server, _ := newStorage(t)
	defer server.Terminate(t)
	if _, err := storage.Allocate(1); !strings.Contains(err.Error(), "cannot allocate resources of type serviceipallocations at this time") {
		t.Fatal(err)
	}
}

func TestStore(t *testing.T) {
	storage, server, backing := newStorage(t)
	defer server.Terminate(t)
	if err := storage.storage.Set(context.TODO(), key(), validNewRangeAllocation(), nil, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := storage.Allocate(2); err != nil {
		t.Fatal(err)
	}
	ok, err := backing.Allocate(2)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Expected backing allocation to fail")
	}
	if ok, err := storage.Allocate(2); ok || err != nil {
		t.Fatal("Expected allocation to fail")
	}

	other := allocator.NewAllocationMap(100, "rangeSpecValue")

	allocation := &api.RangeAllocation{}
	if err := storage.storage.Get(context.TODO(), key(), allocation, false); err != nil {
		t.Fatal(err)
	}
	if allocation.Range != "rangeSpecValue" {
		t.Errorf("unexpected stored Range: %s", allocation.Range)
	}
	if err := other.Restore("rangeSpecValue", allocation.Data); err != nil {
		t.Fatal(err)
	}
	if !other.Has(2) {
		t.Fatalf("could not restore allocated IP: %#v", other)
	}

	other = allocator.NewAllocationMap(100, "rangeSpecValue")
	otherStorage := NewEtcd(other, "/ranges/serviceips", api.Resource("serviceipallocations"), storage.storage)
	if ok, err := otherStorage.Allocate(2); ok || err != nil {
		t.Fatal(err)
	}
}
