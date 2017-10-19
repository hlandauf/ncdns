package backend_test

import (
	"github.com/namecoin/ncdns/backend"
	"github.com/namecoin/ncdns/testutil"
	"gopkg.in/hlandau/madns.v1/merr"
	"net"
	"testing"
)

func TestBackend(t *testing.T) {
	b, err := backend.New(&backend.Config{
		CanonicalNameservers: []string{"ncdns-ns1.example.com.", "ncdns-ns2.example.com."},
		VanityIPs:            []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("::1")},
		Hostmaster:           "hostmaster@example.com",

		FakeNames: map[string]string{
			"d/example":  `{"ip": "192.0.2.1"}`,
			"d/example2": `{"ip": ["192.0.2.1","192.0.2.2"]}`,
		},
		FakesOnly: true,
	})
	if err != nil {
		t.Fatalf("couldn't instantiate backend")
	}

	// Most of the mapping functionality is tested in the ncdomain package, so this
	// focuses on excercising the parts specific to the backend package.
	var items = []struct {
		qname, result string
		err           error
	}{
		// Normal requests.
		{
			`example.bit.`,
			`example.bit. IN A 192.0.2.1`,
			nil,
		},
		{
			`example2.bit.`,
			`example2.bit. IN A 192.0.2.1
example2.bit. IN A 192.0.2.2`,
			nil,
		},

		// Out-of-zone requests.
		{
			`example.com.`,
			``,
			merr.ErrNotInZone,
		},
		{
			`com.`,
			``,
			merr.ErrNotInZone,
		},

		// Root domain.
		{
			`bit.`,
			`bit. 86400 IN A 192.0.2.1
bit. 86400 IN AAAA ::1
bit. 86400 IN NS ncdns-ns1.example.com.
bit. 86400 IN NS ncdns-ns2.example.com.
bit. 86400 IN SOA ncdns-ns1.example.com. hostmaster.example.com. 1 600 600 7200 600`,
			nil,
		},

		// The meta domain should not be present when canonical nameservers are configured.
		{
			`this.x--nmc.bit.`,
			``,
			merr.ErrNoSuchDomain,
		},
	}

	for i, item := range items {
		rrs, err := b.Lookup(item.qname)
		if item.err != nil {
			if err != item.err {
				t.Errorf("item %d was supposed to fail with error %v but instead failed with error %v", i, item.err, err)
			}
			continue
		}

		if err != nil {
			t.Errorf("item %d failed with error: %v", i, err)
			continue
		}

		rrstr := testutil.CanonicalizeRRsToString(rrs)
		if rrstr != item.result {
			t.Errorf("item %d didn't match:\n%+v    !=\n%+v\n\n", i, rrstr, item.result)
		}
	}
}

func TestBackendMetaDomain(t *testing.T) {
	b, err := backend.New(&backend.Config{
		FakesOnly: true,
		SelfIP:    "127.127.127.127",
	})
	if err != nil {
		t.Fatalf("couldn't instantiate backend")
	}

	rrs, err := b.Lookup("this.x--nmc.bit.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `this.x--nmc.bit. 86400 IN A 127.127.127.127`
	rrstr := testutil.CanonicalizeRRsToString(rrs)
	if rrstr != expected {
		t.Errorf("didn't match:\n%+v    !=\n%+v\n\n", rrstr, expected)
	}
}
