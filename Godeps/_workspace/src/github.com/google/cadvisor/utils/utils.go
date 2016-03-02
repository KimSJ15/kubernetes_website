// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import "fmt"

// Returns a mask of all cores on the machine if the passed-in mask is empty.
func FixCpuMask(mask string, cores int) string {
	if mask == "" {
		if cores > 1 {
			mask = fmt.Sprintf("0-%d", cores-1)
		} else {
			mask = "0"
		}
	}
	return mask
}
