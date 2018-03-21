// Package netclient retrieves data from the unauthenticated Telkom Mobile portal
// available when connected to the internet via Telkom Mobile.
package netclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	urlHome       = "https://onnetsecure.telkom.co.za/onnet/public/mobileData?sid="
	urlBundleData = "https://onnetsecure.telkom.co.za/onnet/public/dwr/call/plaincall/mobileDataServiceWrapper.getFreeResources.dwr"
)

// GetBundles retrieves bundle information from the unauthentiated
// Telkom Mobile portal.
// Note that the correct mobile number of the SIM card currently used to connect
// to Telkom Mobile must be specified.
func GetBundles(mobileNum string) ([]byte, error) {
	// Get cookie from home page:
	res, err := http.Get(urlHome)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	// Get bundle info:
	body := fmt.Sprintf("callCount=1\n"+
		"scriptSessionId=\n"+
		"c0-scriptName=mobileDataServiceWrapper\n"+
		"c0-methodName=getFreeResources\n"+
		"c0-id=0\n"+
		"c0-param0=string:%s\n"+
		"batchId=0",
		mobileNum)
	req, err := http.NewRequest("POST", urlBundleData, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for _, c := range res.Cookies() {
		req.AddCookie(c)
	}
	client := &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
