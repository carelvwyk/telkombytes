package bundles

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input           []byte
		expectError     bool
		expectedBundles BundleList
	}{
		{[]byte{}, false, nil},
		{[]byte("s2.endBillCycle="), true, nil},
		{[]byte(testData), false,
			BundleList{
				Bundle{
					Name:           "Once-off LTE/LTE-A Night Surfer Data",
					ExpiryDate:     time.Date(2018, 4, 20, 0, 0, 0, 0, time.UTC),
					BytesUsed:      0,
					BytesRemaining: 21474836480,
				},
				Bundle{
					Name:           "Once-off LTE/LTE-A Anytime Data",
					ExpiryDate:     time.Date(2018, 4, 20, 0, 0, 0, 0, time.UTC),
					BytesUsed:      701075310,
					BytesRemaining: 20773761170,
				},
				Bundle{
					Name:           "Wi-Fi Data Unlimited Speed",
					ExpiryDate:     time.Date(2018, 3, 31, 0, 0, 0, 0, time.UTC),
					BytesUsed:      0,
					BytesRemaining: 10737418240,
				},
			}},
	}
	for i, test := range tests {
		bl, err := Parse(test.input)
		if (err != nil) != test.expectError {
			t.Errorf("Test %d did not pass expected error check", i)
			continue
		}
		if test.expectError {
			continue
		}
		actual, _ := json.Marshal(bl)
		expected, _ := json.Marshal(test.expectedBundles)
		if string(actual) != string(expected) {
			t.Errorf("Test %d\nExpected:\n%s\nActual:\n%s", i, string(expected),
				string(actual))
		}
	}
}

const testData = `throw 'allowScriptTagRemoting is false.';
//#DWR-INSERT
//#DWR-REPLY
var s0={};var s1=[];var s2={};var s7={};var s3={};var s8={};var s4={};var s9={};var s5={};var s10={};var s6={};var s11={};s0.errorCode=null;s0.errorMessage=null;s0.resultCode="0";
s1[0]=s2;s1[1]=s3;s1[2]=s4;s1[3]=s5;s1[4]=s6;
s2.info="GPRS: 21474836480 Bytes remaining 0 Bytes used  Expires on Sat Apr 21 2018";s2.service="GPRS";s2.subscriberFreeResource=s7;
s7.endBillCycle="Sat Apr 21 2018";s7.expiryDate="Fri Apr 20 2018";s7.measure="Bytes";s7.service="GPRS";s7.startBillCycle="Sat Apr 21 2018";s7.timeBased=false;s7.totalAmount="21474836480";s7.totalAmountAndMeasure="20480 MB";s7.type="5125";s7.typeName="Once-off LTE/LTE-A Night Surfer Data";s7.usedAmount="0";s7.usedAmountAndMeasure="0 MB";
s3.info="GPRS: 20773761170 Bytes remaining 701075310 Bytes used  Expires on Sat Apr 21 2018";s3.service="GPRS";s3.subscriberFreeResource=s8;
s8.endBillCycle="Sat Apr 21 2018";s8.expiryDate="Fri Apr 20 2018";s8.measure="Bytes";s8.service="GPRS";s8.startBillCycle="Sat Apr 21 2018";s8.timeBased=false;s8.totalAmount="20773761170";s8.totalAmountAndMeasure="19811 MB";s8.type="5127";s8.typeName="Once-off LTE/LTE-A Anytime Data";s8.usedAmount="701075310";s8.usedAmountAndMeasure="669 MB";
s6.info="WLAN: 10737418240 Bytes remaining 0 Bytes used  Expires on Sun Apr 01 2018";s6.service="WLAN";s6.subscriberFreeResource=s11;
s11.endBillCycle="Sun Apr 01 2018";s11.expiryDate="Sat Mar 31 2018";s11.measure="Bytes";s11.service="WLAN";s11.startBillCycle="Sun Apr 01 2018";s11.timeBased=false;s11.totalAmount="10737418240";s11.totalAmountAndMeasure="10240 MB";s11.type="5049";s11.typeName="Wi-Fi Data Unlimited Speed";s11.usedAmount="0";s11.usedAmountAndMeasure="0 MB";
dwr.engine._remoteHandleCallback('0','0',{result:s0,freeResources:s1});`
