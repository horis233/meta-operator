//
// Copyright 2021 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package util

import (
	"os"
	"sort"

	"k8s.io/client-go/discovery"
)

// GetOperatorNamespace returns the Namespace of the operator
func GetOperatorNamespace() string {
	ns, found := os.LookupEnv("OPERATOR_NAMESPACE")
	if !found {
		return ""
	}
	return ns
}

// GetWatchNamespace returns the Namespace of the operator
func GetWatchNamespace() string {
	ns, found := os.LookupEnv("WATCH_NAMESPACE")
	if !found {
		return GetOperatorNamespace()
	}
	return ns
}

// GetInstallScope returns the scope of the installation
func GetInstallScope() string {
	ns, found := os.LookupEnv("INSTALL_SCOPE")
	if !found {
		return "cluster"
	}
	return ns
}

// ResourceExists returns true if the given resource kind exists
// in the given api groupversion
func ResourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {
	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

//StringSliceContentEqual checks if the contant from two string slice are the same
func StringSliceContentEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
