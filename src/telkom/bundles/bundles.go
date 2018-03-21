// Package bundles defines a Telkom Mobile bundle data structure and functions
// to parse raw bundle data retrieved by the Telkom Mobile portal net client.
package bundles

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type BundleList []Bundle

// CapRemainingBytes returns the amount of bytes remaining in your data cap
// disregarding night surfer and wifi bundles.
func (bl BundleList) CapRemainingBytes() int64 {
	var r int64
	for _, b := range bl {
		lower := strings.ToLower(b.Name)
		if !strings.Contains(lower, "night") &&
			!strings.Contains(lower, "wi-fi") {
			r += b.BytesRemaining
		}
	}
	return r
}

// Bundle represents a Telkom Mobile bundle and its state.
type Bundle struct {
	Name           string
	ExpiryDate     time.Time
	BytesUsed      int64
	BytesRemaining int64
}

// BytesTotal returns the total bytes included in this bundle
func (b Bundle) BytesTotal() int64 {
	return b.BytesRemaining + b.BytesUsed
}

func (b Bundle) String() string {
	return fmt.Sprintf("%s REMAINING: %.2f / %.2f GB",
		b.Name,
		float64(b.BytesRemaining)/1024/1024/1024,
		float64(b.BytesTotal()/1024/1024/1024))
}

// parseKV parses and cleans a Telkom data variable key-value pair.
func parseKV(s string) (key, value string, err error) {
	parts := strings.Split(s, "=")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("element is not a key/value pair")
	}
	kp := strings.Split(parts[0], ".")
	if len(kp) != 2 {
		return "", "", fmt.Errorf("invalid key format (%s)", kp)
	}
	key = kp[1]
	value = strings.Trim(parts[1], "\"")
	return
}

// parseKVLine parses and cleans a Telkom data variable line.
func parseKVLine(s string) (values map[string]string, err error) {
	values = make(map[string]string)
	for _, p := range strings.Split(s, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		k, v, err := parseKV(p)
		if err != nil {
			return nil, err
		}
		values[k] = v
	}
	return
}

// Parse parses the provided raw Telkom Mobile bundle data retrieved by the
// net client and returns a list of bundles.
func Parse(telkomBundleData []byte) (bl BundleList, err error) {
	const dateFormat = "Mon Jan 2 2006"
	for _, line := range strings.Split(string(telkomBundleData), "\n") {
		if !strings.Contains(line, ".endBillCycle") {
			continue
		}
		values, err := parseKVLine(line)
		if err != nil {
			return nil, err
		}
		b := Bundle{}
		b.Name = values["typeName"]
		if b.ExpiryDate, err = time.Parse(dateFormat, values["expiryDate"]); err != nil {
			return nil, err
		}
		if b.BytesUsed, err = strconv.ParseInt(values["usedAmount"], 10, 64); err != nil {
			return nil, err
		}
		if b.BytesRemaining, err = strconv.ParseInt(values["totalAmount"], 10, 64); err != nil {
			return nil, err
		}
		bl = append(bl, b)
	}
	return
}
