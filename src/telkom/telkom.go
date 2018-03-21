// Package telkom provides functions to retrieve remaining bundle and cap
// information.
package telkom

import (
	"telkom/bundles"
	"telkom/netclient"
)

// GetBundlesInfo returns information about currently active Telkom Mobile
// account bundles
func GetBundlesInfo(mobileNum string) (bundles.BundleList, error) {
	bundleData, err := netclient.GetBundles(mobileNum)
	if err != nil {
		return nil, err
	}
	return bundles.Parse(bundleData)
}
