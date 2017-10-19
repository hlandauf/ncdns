package testutil

import (
	"github.com/miekg/dns"
	"sort"
	"strings"
)

// Used for comparing RRs. Testing use only.
func CanonicalizeRRsToString(rrs []dns.RR) string {
	var rrstrs []string
	for _, rr := range rrs {
		s := rr.String()
		s = strings.Replace(s, "\t600\t", "\t", -1) // XXX
		rrstrs = append(rrstrs, strings.Replace(s, "\t", " ", -1))
	}
	sort.Strings(rrstrs)
	return strings.Join(rrstrs, "\n")
}
