// Package telkom provides functions to retrieve remaining bundle and cap
// information.
package telkom

import (
	"errors"

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
	if len(bundleData) == 0 {
		return nil, errors.New("No data returned, please check mobilenumber")
	}
	return bundles.Parse(bundleData)
}
