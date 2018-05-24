package netclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const (
	userAgent     = "TelkomBytes 0.2"
	colonSub      = "COLON"
	betaLoginHost = "https://beta-login.telkom.co.za"
)

// Bundle represents a Telkom bundle including the bundle name, total amount and
// used amount in bytes.
type Bundle struct {
	Name           string
	Service        string
	RemainingBytes int64
	UsedBytes      int64
}

func (b Bundle) TotalBytes() int64 {
	return b.UsedBytes + b.RemainingBytes
}

// Service represents a Telkom service associated with a user account. Each
// service is identified by an MSISDN.
type Service struct {
	Msisdn  string
	Bundles []Bundle
}

func (s Service) String() string {
	res := s.Msisdn + ":\n"
	for _, b := range s.Bundles {
		res += fmt.Sprintf("Remaining %.2f GB of %.2f GB on %s\n",
			float64(b.RemainingBytes)/1024/1024/1024,
			float64(b.TotalBytes())/1024/1024/1024, b.Name)
	}
	return res
}

func (s Service) NonFreeBytesRemaining() (remaining int64) {
	for _, b := range s.Bundles {
		if b.Service != "GPRS" {
			continue
		}
		lowerName := strings.ToLower(b.Name)
		if strings.Contains(lowerName, "night surfer") {
			continue
		}
		if strings.Contains(lowerName, "all networks data") ||
			strings.Contains(lowerName, "smartbroadband data") {
			remaining += b.RemainingBytes
		}
	}
	return
}

func (s Service) NightSurferRemaining() (remaining int64) {
	for _, b := range s.Bundles {
		if b.Service != "GPRS" {
			continue
		}
		lowerName := strings.ToLower(b.Name)
		if strings.Contains(lowerName, "night surfer") {
			remaining += b.RemainingBytes
		}
	}
	return
}

func printRequest(req *http.Request) {
	fmt.Printf("> %s %s %s\n", req.Method, req.URL.Path, req.Proto)
	fmt.Printf("> Host: %s\n", req.URL.Host)
	for name, fields := range req.Header {
		fmt.Printf("> %s: %s\n", name, strings.Join(fields, "; "))
	}
	fmt.Println()
}

func handleRedir(req *http.Request, via []*http.Request, jar *cookiejar.Jar) error {
	if len(via) == 0 {
		return nil
	} else if len(via) >= 10 {
		// Avoid infinite redirect loops - Telkom site has a habbit of creating these.
		return http.ErrUseLastResponse
	}

	prevReq := *via[len(via)-1]
	prevResp := req.Response
	if prevResp != nil {
		// Telkom site returns cookies with invalid names (contains ":" rune)
		// Go rejects these cookies and won't store them in the jar, so we
		// have to sanitise them first and add them to the cookie jar manually.
		for i, cookie := range prevResp.Header["Set-Cookie"] {
			if strings.Contains(cookie, ":") {
				prevResp.Header["Set-Cookie"][i] =
					strings.Replace(cookie, ":", colonSub, -1)
			}
			// fmt.Printf("< Setting cookie: %v for domain %s\n",
			// 	prevResp.Header["Set-Cookie"][i], prevReq.URL.Host)
		}
		jar.SetCookies(prevReq.URL, prevResp.Cookies())
		// For good measure:
		u, _ := url.Parse(betaLoginHost)
		jar.SetCookies(u, prevResp.Cookies())
	}

	// Update Cookies on request (desanitise problematic cookies)
	req.Header["Cookie"] = []string{}
	for _, c := range jar.Cookies(req.URL) {
		req.AddCookie(c)
	}
	for i, ch := range req.Header["Cookie"] {
		req.Header["Cookie"][i] = strings.Replace(ch, "COLON", ":", -1)
	}

	//printRequest(req)
	return nil
}

func logIn(username, password string) (jar *cookiejar.Jar, err error) {
	const (
		homeURL  = "https://selfservice.telkom.co.za/selfservice"
		loginURL = "https://beta-login.telkom.co.za/oam/server/auth_cred_submit"
	)
	jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: nil})
	if err != nil {
		return
	}

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return handleRedir(req, via, jar)
		},
	}

	// Get context cookies from home page
	homeReq, err := http.NewRequest("GET", homeURL, nil)
	if err != nil {
		return
	}
	homeReq.Header.Set("User-Agent", userAgent)
	homeReq.Header.Set("Accept", "*/*")
	homeResp, err := client.Do(homeReq)
	if err != nil {
		return
	}
	defer homeResp.Body.Close()
	if sc := homeResp.StatusCode; sc != 200 {
		return nil, fmt.Errorf("Request to %s returned unexpected status %d",
			homeURL, sc)
	}

	// Get login cookies:
	loginBody := url.Values{
		"v":             []string{"1.4"},
		"challenge_url": []string{"https%3A%2F%2Fapps.telkom.co.za%2Falpha%2Fpublic%2Flogin"},
		"locale":        []string{"en_GB"},
		"resource_url":  []string{"https%253A%252F%252Fselfservice.telkom.co.za%252Fdigital%252Fdashboard%252F"},
		"ssousername":   []string{username},
		"password":      []string{password},
	}
	b := loginBody.Encode()

	loginReq, err := http.NewRequest("POST", loginURL, strings.NewReader(b))
	if err != nil {
		log.Println(err)
		return
	}
	loginReq.Header.Set("User-Agent", userAgent)
	loginReq.Header.Set("Accept", "*/*")
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Add("Content-Length", strconv.Itoa(len(b)))
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Println(err)
		return
	}
	defer loginResp.Body.Close()
	//printRequest(loginReq)
	if sc := loginResp.StatusCode; sc != 200 {
		return nil, fmt.Errorf("Request to %s returned unexpected status %d",
			loginURL, sc)
	}
	return
}

// getAssociatedServices must be called after login
func getAssociatedServices(cookies *cookiejar.Jar) (msisdns []string, err error) {
	const url = "https://selfservice.telkom.co.za/eportal/eCustomer/api/getCustomerAssociatedServices"
	response := struct {
		Payload []struct {
			ServiceNumber string
		}
	}{}
	b, err := apiPostRequest(url, "", cookies)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &response)
	if err != nil {
		return
	}
	msisdns = make([]string, len(response.Payload))
	for i, s := range response.Payload {
		msisdns[i] = s.ServiceNumber
	}
	return
}

func getFreeResources(msisdn string, cookies *cookiejar.Jar) ([]Bundle, error) {
	const freeResourcesURL = "https://selfservice.telkom.co.za/onnet/protected/api/getFreeResources"
	response := struct {
		ResultMessage string
		Payload       []struct {
			Service                string
			SubscriberFreeResource struct {
				TotalAmount string // Note: Can have value "Unlimited"
				UsedAmount  string // Note: Can have value "Unlimited"
				TypeName    string
				Service     string
			}
		}
	}{}
	body := url.Values{
		"msisdn":  []string{msisdn},
		"version": []string{"1"},
	}
	b, err := apiPostRequest(freeResourcesURL, body.Encode(), cookies)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}
	if rm := response.ResultMessage; rm != "Free resources sucessfully retrieved." {
		return nil, fmt.Errorf("Unexpected result message: %s", rm)
	}

	bundles := []Bundle{}
	parseNumber := func(s string) int64 {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		return -1
	}
	for _, pl := range response.Payload {
		sfr := pl.SubscriberFreeResource
		b := Bundle{
			Name:           sfr.TypeName,
			Service:        sfr.Service,
			RemainingBytes: parseNumber(sfr.TotalAmount),
			UsedBytes:      parseNumber(sfr.UsedAmount),
		}
		bundles = append(bundles, b)
	}
	return bundles, nil
}

func apiPostRequest(url, body string, cookies *cookiejar.Jar) (
	[]byte, error) {

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	// Manually handle cookies because of Telkom invalid cookie name issue
	reqCookies := cookies.Cookies(req.URL)
	for _, c := range reqCookies {
		req.AddCookie(c)
	}
	// Replace substituted cookie name with invalid but expected name directly in header
	req.Header["Cookie"] = []string{strings.Replace(req.Header["Cookie"][0],
		colonSub, ":", -1)}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func GetServiceBundles(username, password string) ([]Service, error) {
	// Log in
	cookies, err := logIn(username, password)
	if err != nil {
		return nil, err
	}

	// Retrieve associated services
	serviceMsisdns, err := getAssociatedServices(cookies)
	if err != nil {
		return nil, err
	}

	// Retrieve bundle information per service
	services := []Service{}
	for _, msisdn := range serviceMsisdns {
		fr, err := getFreeResources(msisdn, cookies)
		if err != nil {
			return nil, err
		}
		services = append(services, Service{
			Msisdn:  msisdn,
			Bundles: fr,
		})
	}
	return services, nil
}

// // GetOnNetBundles retrieves bundle information from the unauthentiated
// // Telkom Mobile portal.
// // Note that the correct mobile number of the SIM card currently used to connect
// // to Telkom Mobile must be specified.
// func GetOnNetBundles(mobileNum string) ([]byte, error) {
// 	const (
// 		urlHome       = "https://onnetsecure.telkom.co.za/onnet/public/mobileData?sid="
// 		urlBundleData = "https://onnetsecure.telkom.co.za/onnet/public/dwr/call/plaincall/mobileDataServiceWrapper.getFreeResources.dwr"
// 	)
// 	// Get cookie from home page:
// 	res, err := http.Get(urlHome)
// 	if err != nil {
// 		return nil, err
// 	}
// 	res.Body.Close()
// 	// Get bundle info:
// 	body := fmt.Sprintf("callCount=1\n"+
// 		"scriptSessionId=\n"+
// 		"c0-scriptName=mobileDataServiceWrapper\n"+
// 		"c0-methodName=getFreeResources\n"+
// 		"c0-id=0\n"+
// 		"c0-param0=string:%s\n"+
// 		"batchId=0",
// 		mobileNum)
// 	req, err := http.NewRequest("POST", urlBundleData, strings.NewReader(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, c := range res.Cookies() {
// 		req.AddCookie(c)
// 	}
// 	client := &http.Client{}
// 	res, err = client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()
// 	return ioutil.ReadAll(res.Body)
// }
