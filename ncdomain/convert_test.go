package ncdomain_test

import (
	"fmt"
	_ "github.com/hlandau/nctestsuite"
	"github.com/namecoin/ncdns/ncdomain"
	"github.com/namecoin/ncdns/testutil"
	"github.com/namecoin/ncdns/util"
	"testing"
)

func TestSuite(t *testing.T) {
	items := testutil.SuiteReader(t)
	for ti := range items {
		resolve := func(name string) (string, error) {
			v, ok := ti.Names[name]
			if !ok {
				return "", fmt.Errorf("not found")
			}

			return v, nil
		}

		for k, jsonValue := range ti.Names {
			dnsName, err := util.NamecoinKeyToBasename(k)
			if err != nil {
				continue
			}

			errCount := 0
			errFunc := func(err error, isWarning bool) {
				if !isWarning {
					//t.Logf("error: %v", err)
					errCount++
				}
			}

			v := ncdomain.ParseValue(k, jsonValue, resolve, errFunc)
			if v == nil {
				if ti.NumErrors != 1000 {
					t.Errorf("Couldn't parse item %q: %v", ti.ID, jsonValue)
				}
				continue
			} else {
				if ti.NumErrors == 1000 {
					t.Errorf("Parsed item which should be unparseable: %q: %v", ti.ID, jsonValue)
				}
			}

			rrs, _ := v.RRsRecursive(nil, dnsName+".bit.", dnsName+".bit.")
			rrstr := testutil.CanonicalizeRRsToString(rrs)

			// CHECK MATCH
			if rrstr != ti.Records {
				t.Errorf("Didn't match: %s\n%+v\n    !=\n%+v\n\n%#v\n\n%#v", ti.ID, rrstr, ti.Records, v, rrs)
			}

			if errCount != ti.NumErrors {
				t.Errorf("Error count didn't match: %d != %d (%s)\n", errCount, ti.NumErrors, ti.ID)
			}
		}
	}
}

func TestHostmaster(t *testing.T) {
	errCount := 0
	errFunc := func(err error, isWarning bool) {
		if !isWarning {
			errCount++
		}
	}

	//
	v := ncdomain.ParseValue("d/example", `{"email":"hostmaster@example.bit"}`, nil, errFunc)
	if v == nil {
		t.Errorf("Couldn't parse")
		return
	}

	if v.Hostmaster != "hostmaster@example.bit" {
		t.Errorf("Incorrect hostmaster: %q", v.Hostmaster)
	}

	//
	v = ncdomain.ParseValue("d/example", `{"email":"@example.bit"}`, nil, errFunc)
	if v == nil {
		t.Errorf("Couldn't parse")
	}

	if v.Hostmaster != "" {
		t.Errorf("Accepted malformed hostmaster: %q", v.Hostmaster)
	}

	//
	v = ncdomain.ParseValue("d/example", `{"email":["foo@example.bit"]}`, nil, errFunc)
	if v == nil {
		t.Errorf("Couldn't parse")
	}

	if v.Hostmaster != "" {
		t.Errorf("Accepted malformed hostmaster")
	}
}
