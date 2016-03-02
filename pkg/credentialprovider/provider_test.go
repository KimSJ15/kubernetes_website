/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package credentialprovider

import (
	"testing"
	"time"
)

func TestCachingProvider(t *testing.T) {
	provider := &testProvider{
		Count: 0,
	}

	cache := &CachingDockerConfigProvider{
		Provider: provider,
		Lifetime: 1 * time.Second,
	}

	if provider.Count != 0 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 1 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}

	time.Sleep(cache.Lifetime)
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 2 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}

	time.Sleep(cache.Lifetime)
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 3 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
}
