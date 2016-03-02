/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package stats

import "time"

// ResourceAnalyzer provides statistics on node resource consumption
type ResourceAnalyzer interface {
	Start()

	fsResourceAnalyzerInterface
}

// resourceAnalyzer implements ResourceAnalyzer
type resourceAnalyzer struct {
	*fsResourceAnalyzer
}

var _ ResourceAnalyzer = &resourceAnalyzer{}

// NewResourceAnalyzer returns a new ResourceAnalyzer
func NewResourceAnalyzer(statsProvider StatsProvider, calVolumeFrequency time.Duration) ResourceAnalyzer {
	return &resourceAnalyzer{newFsResourceAnalyzer(statsProvider, calVolumeFrequency)}
}

// Start starts background functions necessary for the ResourceAnalyzer to function
func (ra *resourceAnalyzer) Start() {
	ra.fsResourceAnalyzer.Start()
}
