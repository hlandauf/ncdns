package server

import (
	"github.com/miekg/dns"
	"github.com/namecoin/ncdns/testutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeNewKeys() (path string, err error) {
	path, err = ioutil.TempDir("", "ncdnstest")
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			eraseKeys(path)
		}
	}()

	pPublic := filepath.Join(path, "public")
	pPrivate := filepath.Join(path, "private")

	k := dns.DNSKEY{
		Hdr: dns.RR_Header{
			Name:   "bit.",
			Rrtype: dns.TypeDNSKEY,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		Flags:     0,
		Protocol:  3,
		Algorithm: dns.ECDSAP256SHA256,
	}
	private, err := k.Generate(256)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(pPublic, []byte(k.String()), 0644)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(pPrivate, []byte(k.PrivateKeyString(private)), 0644)
	return
}

func eraseKeys(path string) {
	// Best effort.
	os.Remove(filepath.Join(path, "private"))
	os.Remove(filepath.Join(path, "public"))
	os.Remove(path)
}

// This test spins up the full ncdns server and sends DNS requests for it. This
// tests the ncdns stack throughout almost its entire depth; only the Namecoin
// node querying logic is untested, as mock Namecoin data is used.
func TestServer(t *testing.T) {
	// Create new keys and write them to a temporary directory
	tmpDir, err := makeNewKeys()
	if err != nil {
		t.Fatalf("couldn't write random keys for server testing: %v", err)
	}

	defer eraseKeys(tmpDir)

	// Startup.
	cfg := Config{
		Bind:            ":1253",
		CanonicalSuffix: "bit",
		CacheMaxEntries: 100,

		ZonePublicKey:  filepath.Join(tmpDir, "public"),
		ZonePrivateKey: filepath.Join(tmpDir, "private"),

		fakeNames: map[string]string{
			"d/example": `{"ip":"192.0.2.1"}`,
		},
		fakesOnly: true,
	}
	server, err := New(&cfg)
	if err != nil {
		t.Fatalf("couldn't instantiate server: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("couldn't start server: %v", err)
	}

	// Run tests.
	items := []struct {
		QName    string
		QType    uint16
		Expected string
	}{
		{"example.bit.", dns.TypeA, "example.bit. IN A 192.0.2.1"},
	}

	for _, item := range items {
		testQuery(t, item.QName, item.QType, item.Expected)
	}

	// Shutdown.
	err = server.Stop()
	if err != nil {
		t.Fatalf("couldn't stop server: %v", err)
	}
}

func testQuery(t *testing.T, qname string, qtype uint16, expectString string) {
	msg := &dns.Msg{
		Compress: true,
	}

	msg.SetQuestion(qname, qtype)
	msg.SetEdns0(4096, true)

	res, err := dns.Exchange(msg, ":1253")
	if err != nil {
		t.Fatalf("error making DNS query: %v", err)
		return
	}

	var allRRs []dns.RR
	allRRs = append(allRRs, res.Answer...)
	allRRs = append(allRRs, res.Ns...)
	for _, rr := range res.Extra {
		if rr.Header().Rrtype == dns.TypeOPT {
			continue
		}

		allRRs = append(allRRs, rr)
	}

	rrStr := testutil.CanonicalizeRRsToString(allRRs)

	if rrStr != expectString {
		t.Fatalf("mismatch for query %q %v: got\n----------\n%s\n----------\nexpected\n----------\n%v\n----------\n", qname, qtype, rrStr, expectString)
	}
}

func normalizeRR(rr dns.RR) string {
	return strings.Replace(rr.String(), "\t600\t", "\t", -1)
}
