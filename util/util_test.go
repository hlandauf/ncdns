package util_test

import "testing"
import "github.com/namecoin/ncdns/util"
import "gopkg.in/hlandau/madns.v1/merr"

type item struct {
	input            string
	expectedHead     string
	expectedRest     string
	expectedTail     string
	expectedTailRest string
}

var items = []item{
	item{"", "", "", "", ""},
	item{"a", "a", "", "a", ""},
	item{"alpha", "alpha", "", "alpha", ""},
	item{"alpha.beta", "beta", "alpha", "alpha", "beta"},
	item{"alpha.beta.gamma", "gamma", "alpha.beta", "alpha", "beta.gamma"},
	item{"alpha.beta.gamma.delta", "delta", "alpha.beta.gamma", "alpha", "beta.gamma.delta"},
	item{"alpha.beta.gamma.delta.", "delta", "alpha.beta.gamma", "alpha", "beta.gamma.delta."},
}

func TestSplitDomainHeadTail(t *testing.T) {
	for i := range items {
		head, rest := util.SplitDomainHead(items[i].input)
		tail, trest := util.SplitDomainTail(items[i].input)
		if head != items[i].expectedHead {
			t.Errorf("Input \"%s\": head \"%s\" does not equal expected value \"%s\"", items[i].input, head, items[i].expectedHead)
		}
		if rest != items[i].expectedRest {
			t.Errorf("Input \"%s\": rest \"%s\" does not equal expected value \"%s\"", items[i].input, rest, items[i].expectedRest)
		}
		if tail != items[i].expectedTail {
			t.Errorf("Input \"%s\": tail \"%s\" does not equal expected value \"%s\"", items[i].input, tail, items[i].expectedTail)
		}
		if trest != items[i].expectedTailRest {
			t.Errorf("Input \"%s\": tail rest \"%s\" does not equal expected value \"%s\"", items[i].input, trest, items[i].expectedTailRest)
		}
	}
}

type aitem struct {
	input            string
	anchor           string
	expectedSubname  string
	expectedBasename string
	expectedRootname string
	expectedError    error
}

var aitems = []aitem{
	aitem{"", "bit", "", "", "", merr.ErrNotInZone},
	aitem{".", "bit", "", "", "", merr.ErrNotInZone},
	aitem{"d.", "bit", "", "", "", merr.ErrNotInZone},
	aitem{"a.b.c.d.", "bit", "", "", "", merr.ErrNotInZone},
	aitem{"a.b.c.d.bit.", "bit", "a.b.c", "d", "bit", nil},
	aitem{"d.bit.", "bit", "", "d", "bit", nil},
	aitem{"bit.", "bit", "", "", "bit", nil},
	aitem{"bit.x.y.z.", "bit", "", "", "bit.x.y.z", nil},
	aitem{"d.bit.x.y.z.", "bit", "", "d", "bit.x.y.z", nil},
	aitem{"c.d.bit.x.y.z.", "bit", "c", "d", "bit.x.y.z", nil},
	aitem{"a.b.c.d.bit.x.y.z.", "bit", "a.b.c", "d", "bit.x.y.z", nil},
}

func TestSplitDomainByFloatingAnchor(t *testing.T) {
	for i, it := range aitems {
		subname, basename, rootname, err := util.SplitDomainByFloatingAnchor(it.input, it.anchor)
		if subname != it.expectedSubname {
			t.Errorf("Item %d: subname \"%s\" does not equal expected value \"%s\"", i, subname, it.expectedSubname)
		}
		if basename != it.expectedBasename {
			t.Errorf("Item %d: basename \"%s\" does not equal expected value \"%s\"", i, basename, it.expectedBasename)
		}
		if rootname != it.expectedRootname {
			t.Errorf("Item %d: rootname \"%s\" does not equal expected value \"%s\"", i, basename, it.expectedRootname)
		}
		if err != it.expectedError {
			t.Errorf("Item %d: error \"%s\" does not equal expected error \"%s\"", i, err, it.expectedError)
		}
	}
}

func TestBasenameToNamecoinKey(t *testing.T) {
	var items = []struct {
		basename, ncKey string
		error           bool
	}{
		{"example", "d/example", false},
		{"examp-le", "d/examp-le", false},
		{"ex-amp-le", "d/ex-amp-le", false},
		{"xn--j6w193g", "d/xn--j6w193g", false},
		{"ex_ample", "", true},
		{"ex\x00ample", "", true},
	}

	for i, item := range items {
		ncKey, err := util.BasenameToNamecoinKey(item.basename)
		if (err != nil) != item.error {
			t.Errorf("Item %d: got error %v, expectedError=%v", i, err, item.error)
			continue
		}
		if ncKey != item.ncKey {
			t.Errorf("Item %d: got %q, expected %q", i, ncKey, item.ncKey)
		}

		if err != nil {
			continue
		}

		basename2, err := util.NamecoinKeyToBasename(item.ncKey)
		if err != nil {
			t.Errorf("Item %d: couldn't reverse mapping, got error: %v", i, err)
			continue
		}

		if basename2 != item.basename {
			t.Errorf("Item %d: non-isomorphic mapping: %q -> %q -> %q", i, item.basename, ncKey, basename2)
		}
	}
}

func TestNamecoinKeyToBasename(t *testing.T) {
	var items = []struct {
		ncKey, basename string
		error           bool
	}{
		{"d/example", "example", false},
		{"d/examp-le", "examp-le", false},
		{"d/xn--j6w193g", "xn--j6w193g", false},
		{"d/ex_ample", "", true},
		{"d/ex\x00ample", "", true},
		{"example", "", true},
	}

	for i, item := range items {
		basename, err := util.NamecoinKeyToBasename(item.ncKey)
		if (err != nil) != item.error {
			t.Errorf("Item %d: got error %v, expectedError=%v", i, err, item.error)
			continue
		}
		if basename != item.basename {
			t.Errorf("Item %d: got %q, expected %q", i, basename, item.basename)
		}

		if err != nil {
			continue
		}

		ncKey2, err := util.BasenameToNamecoinKey(basename)
		if err != nil {
			t.Errorf("Item %d: couldn't reverse mapping, got error: %v", i, err)
			continue
		}

		if ncKey2 != item.ncKey {
			t.Errorf("Item %d: non-isomorphic mapping: %q -> %q -> %q", i, item.ncKey, basename, ncKey2)
		}
	}
}

func TestValidateNameLength(t *testing.T) {
	var items = []struct {
		name string
		ok   bool
	}{
		{"e", true},
		{"example.name.example", true},

		// 255 chars:
		{"aaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},

		// 256 chars, bad:
		{"aaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},

		// 256 chars, good:
		{"aaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.", true},
	}

	for i, item := range items {
		ok := util.ValidateNameLength(item.name)
		if ok != item.ok {
			t.Errorf("Item %d: expected ok=%v, got ok=%v", i, item.ok, ok)
		}
	}
}

func TestValidateNames(t *testing.T) {
	var items = []struct {
		name            string
		isOwnerLabel    bool
		isOwnerName     bool
		isRelOwnerName  bool
		isDomainLabel   bool
		isDomainName    bool
		isRelDomainName bool
		isHostLabel     bool
		isHostName      bool
		isRelHostName   bool
		isServiceName   bool
	}{
		{"e", true, true, true, true, true, true, true, true, true, true},
		{"example", true, true, true, true, true, true, true, true, true, true},
		{"example-with-dashes", true, true, true, true, true, true, true, true, true, true},
		{"xn--j6w193g", true, true, true, true, true, true, true, true, true, true},

		// Valid owner label, but not a valid domain label or host label.
		{"e_xample", true, true, true, false, false, false, false, false, false, true},

		// Valid owner label and host label, but not a valid domain label.
		{"ab--cd", true, true, true, false, false, false, true, true, true, true},
		{"examp--le", true, true, true, false, false, false, true, true, true, true},

		// Valid fully-qualified name.
		{"example.", false, true, true, false, true, true, false, true, true, false},
		{"foo.example.", false, true, true, false, true, true, false, true, true, false},
		{"e.", false, true, true, false, true, true, false, true, true, false},
		{"f.e.", false, true, true, false, true, true, false, true, true, false},

		// Malformed.
		{"a..b", false, false, false, false, false, false, false, false, false, false},

		// Valid as relative name only.
		{"", false, false, true, false, false, true, false, false, true, false},

		// Not a valid owner label.
		{"a$b", false, false, false, false, false, false, false, false, false, false},

		// 63-character labels, 256 character total length with trailing dot.
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.", false, true, true, false, true, true, false, true, true, false},

		// Relative syntax.
		{"@", false, false, true, false, false, true, false, false, true, false},
		{"foo.@", false, false, true, false, false, true, false, false, true, false},
		{"@.", false, false, false, false, false, false, false, false, false, false},

		// All domain names processed are expected to be lowercase.
		{"EXAMPLE", false, false, false, false, false, false, false, false, false, false},
		{"Example", false, false, false, false, false, false, false, false, false, false},
		{"examPle", false, false, false, false, false, false, false, false, false, false},

		// Service names are limited to 62 characters because they are preceded by a _.
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, true, true, true, true, true, true, true, true, false},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true, true, true, true, true, true, true, true, true, true},
	}

	chk := func(i int, f func(n string) bool, name string, shouldBeOK bool, desc string) {
		isOK := f(name)
		if isOK != shouldBeOK {
			t.Errorf("Item %d: %q: expected ok=%v, got ok=%v for %s", i, name, shouldBeOK, isOK, desc)
		}
	}

	for i, item := range items {
		chk(i, util.ValidateOwnerLabel, item.name, item.isOwnerLabel, "owner label")
		chk(i, util.ValidateOwnerName, item.name, item.isOwnerName, "owner name")
		chk(i, util.ValidateRelOwnerName, item.name, item.isRelOwnerName, "rel owner name")
		chk(i, util.ValidateDomainLabel, item.name, item.isDomainLabel, "domain label")
		chk(i, util.ValidateDomainName, item.name, item.isDomainName, "domain name")
		chk(i, util.ValidateRelDomainName, item.name, item.isRelDomainName, "rel domain name")
		chk(i, util.ValidateHostLabel, item.name, item.isHostLabel, "host label")
		chk(i, util.ValidateHostName, item.name, item.isHostName, "hostname")
		chk(i, util.ValidateRelHostName, item.name, item.isRelHostName, "rel hostname")
		chk(i, util.ValidateServiceName, item.name, item.isServiceName, "service name")
	}
}

func TestValidateEmail(t *testing.T) {
	// Basic testing only as the bulk of this is handled by mail.ParseAddress.
	var items = []struct {
		address string
		ok      bool
	}{
		{"foo@example.com", true},
		{"foo", false},
		{"John Smith <foo@example.com>", false},
		{"", false},
	}

	for i, item := range items {
		ok := util.ValidateEmail(item.address)
		if ok != item.ok {
			t.Errorf("Item %d: %q: expected ok=%v, got ok=%v", i, item.address, item.ok, ok)
		}
	}
}

func TestParseFuzzyDomainName(t *testing.T) {
	var items = []struct {
		in, out, outNC string
		ok             bool
	}{
		{"d/example", "example", "d/example", true},
		{"example.bit", "example", "d/example", true},
		{"example.bit.", "example", "d/example", true},
		{"examp--le.bit.", "", "", false},
		{"un.known", "", "", false},
		{"un/known", "", "", false},
		{"bareword", "", "", false},
	}

	for i, item := range items {
		out, outNC, err := util.ParseFuzzyDomainNameNC(item.in)
		if (err == nil) != item.ok {
			t.Errorf("Item %d: %q: expected ok=%v, got ok=%v (%v)", i, item.in, item.ok, err == nil, err)
		}

		if out != item.out {
			t.Errorf("Item %d: %q: expected output %q, got %q", i, item.in, item.out, out)
		}

		if outNC != item.outNC {
			t.Errorf("Item %d: %q: expected output %q, got %q", i, item.in, item.outNC, outNC)
		}
	}
}
