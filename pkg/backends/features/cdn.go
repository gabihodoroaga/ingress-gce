/*
Copyright 2018 The Kubernetes Authors.

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

package features

import (
	"reflect"

	"k8s.io/ingress-gce/pkg/composite"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/klog"
)

// EnsureCDN reads the CDN configuration specified in the ServicePort.BackendConfig
// and applies it to the BackendService. It returns true if there were existing
// settings on the BackendService that were overwritten.
func EnsureCDN(sp utils.ServicePort, be *composite.BackendService) bool {
	if sp.BackendConfig.Spec.Cdn == nil {
		return false
	}
	beTemp := &composite.BackendService{}
	applyCDNSettings(sp, beTemp)
	// Only compare CdnPolicy if it was specified.
	if (beTemp.CdnPolicy != nil && !reflect.DeepEqual(beTemp.CdnPolicy, be.CdnPolicy)) || beTemp.EnableCDN != be.EnableCDN {
		applyCDNSettings(sp, be)
		klog.V(2).Infof("Updated CDN settings for service %v/%v.", sp.ID.Service.Namespace, sp.ID.Service.Name)
		return true
	}
	return false
}

// applyCDNSettings applies the CDN settings specified in the BackendConfig
// to the passed in compute.BackendService. A GCE API call still needs to be
// made to actually persist the changes.
func applyCDNSettings(sp utils.ServicePort, be *composite.BackendService) {
	cdnConfig := sp.BackendConfig.Spec.Cdn
	// Apply the boolean switch
	be.EnableCDN = cdnConfig.Enabled
	cacheKeyPolicy := cdnConfig.CachePolicy
	// Apply the cache key policies if the BackendConfig contains them.
	if cacheKeyPolicy != nil {
		be.CdnPolicy = &composite.BackendServiceCdnPolicy{CacheKeyPolicy: &composite.CacheKeyPolicy{}}
		be.CdnPolicy.CacheKeyPolicy.IncludeHost = cacheKeyPolicy.IncludeHost
		be.CdnPolicy.CacheKeyPolicy.IncludeProtocol = cacheKeyPolicy.IncludeProtocol
		be.CdnPolicy.CacheKeyPolicy.IncludeQueryString = cacheKeyPolicy.IncludeQueryString
		be.CdnPolicy.CacheKeyPolicy.QueryStringBlacklist = cacheKeyPolicy.QueryStringBlacklist
		be.CdnPolicy.CacheKeyPolicy.QueryStringWhitelist = cacheKeyPolicy.QueryStringWhitelist
	}
	// Note that upon creation of a BackendServices, the fields 'IncludeHost',
	// 'IncludeProtocol' and 'IncludeQueryString' all default to true if not
	// explicitly specified.
	if cdnConfig.CacheMode != nil {
		be.CdnPolicy.CacheMode = *cdnConfig.CacheMode
	}
	if cdnConfig.RequestCoalescing != nil {
		be.CdnPolicy.RequestCoalescing = *cdnConfig.RequestCoalescing
	}
	if cdnConfig.ServeWhileStaleSec != nil {
		be.CdnPolicy.ServeWhileStale = *cdnConfig.ServeWhileStaleSec
	}
}
