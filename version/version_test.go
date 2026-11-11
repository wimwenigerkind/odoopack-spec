package version_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/wimwenigerkind/odoopack-spec/version"
)

var (
	maxInt     = strconv.Itoa(math.MaxInt)
	maxIntPlus = strconv.FormatUint(uint64(math.MaxInt)+1, 10)
)

func release(a, b, c, d, e int) version.Version {
	return version.Version{Kind: version.KindRelease, Series: [2]int{a, b}, Semver: [3]int{c, d, e}}
}

func dev(branch string, series [2]int) version.Version {
	return version.Version{Kind: version.KindDev, Branch: branch, Series: series}
}

var noSeries = [2]int{-1, -1}

func mustParse(t *testing.T, s string) version.Version {
	t.Helper()
	v, err := version.Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return v
}

func TestParse(t *testing.T) {
	valid := []struct {
		in   string
		want version.Version
		str  string
	}{
		{"19.0.1.2.3", release(19, 0, 1, 2, 3), "19.0.1.2.3"},
		{"0.0.0.0.0", release(0, 0, 0, 0, 0), "0.0.0.0.0"},
		{"18.0.10.20.30", release(18, 0, 10, 20, 30), "18.0.10.20.30"},
		{" 19.0.1.2.3\n", release(19, 0, 1, 2, 3), "19.0.1.2.3"},
		{"19.0.0.0." + maxInt, release(19, 0, 0, 0, math.MaxInt), "19.0.0.0." + maxInt},
		{"dev-main", dev("main", noSeries), "dev-main"},
		{"dev-19.0", dev("19.0", [2]int{19, 0}), "dev-19.0"},
		{"dev-19.0-fix", dev("19.0-fix", noSeries), "dev-19.0-fix"},
		{"dev-feature/x_y.z", dev("feature/x_y.z", noSeries), "dev-feature/x_y.z"},
		{"dev-a-", dev("a-", noSeries), "dev-a-"},
		{"dev-.hidden", dev(".hidden", noSeries), "dev-.hidden"},
		{"dev-dev-main", dev("dev-main", noSeries), "dev-dev-main"},
		{"dev-19.00", dev("19.00", noSeries), "dev-19.00"},
		{" dev-main ", dev("main", noSeries), "dev-main"},
	}
	for _, tc := range valid {
		t.Run(tc.in, func(t *testing.T) {
			v := mustParse(t, tc.in)
			if v != tc.want {
				t.Errorf("Parse(%q) = %#v, want %#v", tc.in, v, tc.want)
			}
			if got := v.String(); got != tc.str {
				t.Errorf("String() = %q, want %q", got, tc.str)
			}
			if v.IsRelease() == v.IsDev() {
				t.Errorf("IsRelease() = %v, IsDev() = %v", v.IsRelease(), v.IsDev())
			}
		})
	}

	invalid := []string{
		"",
		"   ",
		"19",
		"19.0",
		"19.0.1",
		"19.0.1.2",
		"19.0.1.2.3.4",
		"19.0.1.2.",
		".19.0.1.2",
		"19..0.1.2",
		"19.0.1.2.x",
		"v19.0.1.2.3",
		"19.0.1.2.3-beta",
		"19.0.+1.2.3",
		"19.0.-1.2.3",
		"19.0.-0.2.3",
		"19.0.01.2.3",
		"19.00.1.2.3",
		"019.0.1.2.3",
		"19.0.1 .2.3",
		"19.0.1_0.2.3",
		"19.0.0x1.2.3",
		"19.0.١.2.3",
		"19.0.0.0." + maxIntPlus,
		"19.0.0.0.99999999999999999999999",
		"dev-",
		"dev",
		"Dev-main",
		"dev-feature branch",
		"dev-a\tb",
		"dev-a..b",
		"dev--x",
		"dev-/x",
		"dev-x/",
		"dev-x.lock",
		"dev-a~1",
		"dev-a^",
		"dev-a:b",
		"dev-a?",
		"dev-a*",
		"dev-a[b",
		"dev-a\\b",
		"dev-a@{1}",
		"dev-ä",
	}
	for _, in := range invalid {
		t.Run("invalid/"+in, func(t *testing.T) {
			if v, err := version.Parse(in); err == nil {
				t.Errorf("Parse(%q) = %#v, want error", in, v)
			}
		})
	}
}

func TestSeriesString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"19.0.1.2.3", "19.0"},
		{"dev-19.0", "19.0"},
		{"dev-main", ""},
	}
	for _, tc := range tests {
		if got := mustParse(t, tc.in).SeriesString(); got != tc.want {
			t.Errorf("SeriesString(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseConstraint(t *testing.T) {
	valid := []struct {
		in  string
		str string
	}{
		{"19.0.1.2.3", "19.0.1.2.3"},
		{"19.0", "19.0"},
		{"0.0", "0.0"},
		{"dev-main", "dev-main"},
		{"dev-19.0", "dev-19.0"},
		{" 19.0 ", "19.0"},
	}
	for _, tc := range valid {
		t.Run(tc.in, func(t *testing.T) {
			c, err := version.ParseConstraint(tc.in)
			if err != nil {
				t.Fatalf("ParseConstraint(%q): %v", tc.in, err)
			}
			if got := c.String(); got != tc.str {
				t.Errorf("String() = %q, want %q", got, tc.str)
			}
		})
	}

	invalid := []string{
		"",
		"  ",
		"19",
		"19.0.1",
		"19.0.1.2",
		"19.0.1.2.3.4",
		"19.",
		".0",
		"19.x",
		"19.00",
		"019.0",
		"+19.0",
		"19.-0",
		"19.0.01.2.3",
		"*",
		">=19.0",
		"dev-",
		"dev-a b",
		"dev-a..b",
		"dev--x",
		"dev-x.lock",
		maxIntPlus + ".0",
	}
	for _, in := range invalid {
		t.Run("invalid/"+in, func(t *testing.T) {
			if c, err := version.ParseConstraint(in); err == nil {
				t.Errorf("ParseConstraint(%q) = %v, want error", in, c)
			}
		})
	}
}

func TestConstraintMatches(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{"19.0.1.2.3", "19.0.1.2.3", true},
		{"19.0.1.2.3", "19.0.1.2.4", false},
		{"19.0.1.2.3", "18.0.1.2.3", false},
		{"19.0.1.2.3", "dev-19.0", false},
		{"19.0", "19.0.0.0.0", true},
		{"19.0", "19.0.5.1.0", true},
		{"19.0", "19.1.0.0.0", false},
		{"19.0", "18.0.1.0.0", false},
		{"19.0", "dev-19.0", false},
		{"dev-main", "dev-main", true},
		{"dev-main", "dev-master", false},
		{"dev-19.0", "dev-19.0", true},
		{"dev-19.0", "19.0.1.2.3", false},
	}
	for _, tc := range tests {
		t.Run(tc.constraint+"~"+tc.version, func(t *testing.T) {
			c, err := version.ParseConstraint(tc.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint(%q): %v", tc.constraint, err)
			}
			if got := c.Matches(mustParse(t, tc.version)); got != tc.want {
				t.Errorf("Matches = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"19.0.1.2.3", "19.0.1.2.3", 0},
		{"19.0.0.0.0", "18.0.9.9.9", 1},
		{"18.1.0.0.0", "18.0.9.9.9", 1},
		{"18.0.10.0.0", "18.0.9.9.9", 1},
		{"18.0.1.10.0", "18.0.1.9.0", 1},
		{"18.0.1.0.10", "18.0.1.0.9", 1},
		{"18.0.1.0.0", "18.0.0.99.99", 1},
		{"0.0.0.0.0", "dev-main", 1},
		{"0.0.0.0.0", "dev-19.0", 1},
		{"dev-main", "dev-main", 0},
		// Dev versions order by branch name, byte by byte, so "19.0" < "9.0".
		{"dev-19.0", "dev-9.0", -1},
		{"dev-a", "dev-b", -1},
		{"dev-Main", "dev-main", -1},
	}
	for _, tc := range tests {
		t.Run(tc.a+"~"+tc.b, func(t *testing.T) {
			a, b := mustParse(t, tc.a), mustParse(t, tc.b)
			if got := version.Compare(a, b); got != tc.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
			if got := version.Compare(b, a); got != -tc.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", tc.b, tc.a, got, -tc.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	versions := []string{
		"18.0.3.0.0",
		"19.0.1.2.3",
		"19.0.1.10.0",
		"19.0.1.9.9",
		"19.1.0.0.0",
		"dev-19.0",
		"dev-main",
	}
	tests := []struct {
		constraint string
		want       string
		ok         bool
	}{
		{"19.0", "19.0.1.10.0", true},
		{"18.0", "18.0.3.0.0", true},
		{"19.1", "19.1.0.0.0", true},
		{"17.0", "", false},
		{"19.0.1.2.3", "19.0.1.2.3", true},
		{"19.0.1.2.4", "", false},
		{"dev-main", "dev-main", true},
		{"dev-19.0", "dev-19.0", true},
		{"dev-17.0", "", false},
	}
	var parsed []version.Version
	for _, s := range versions {
		parsed = append(parsed, mustParse(t, s))
	}
	for _, tc := range tests {
		t.Run(tc.constraint, func(t *testing.T) {
			c, err := version.ParseConstraint(tc.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint(%q): %v", tc.constraint, err)
			}
			got, ok := version.Resolve(c, parsed)
			if ok != tc.ok {
				t.Fatalf("Resolve ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.String() != tc.want {
				t.Errorf("Resolve = %s, want %s", got, tc.want)
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		c, _ := version.ParseConstraint("19.0")
		if v, ok := version.Resolve(c, nil); ok {
			t.Errorf("Resolve(nil) = %s, want no match", v)
		}
	})
}

// Before strict parsing "19.0.01.0.0" and "19.0.1.0.0" compared equal but
// printed differently, so the winner of a tie depended on the input order.
func TestResolveTieIndependentOfOrder(t *testing.T) {
	c, err := version.ParseConstraint("19.0")
	if err != nil {
		t.Fatal(err)
	}
	a := mustParse(t, "19.0.1.0.0")
	b := mustParse(t, " 19.0.1.0.0 ")
	older := mustParse(t, "19.0.0.9.0")
	orders := [][]version.Version{
		{a, b, older},
		{b, a, older},
		{older, b, a},
		{older, a, b},
	}
	for _, in := range orders {
		got, ok := version.Resolve(c, in)
		if !ok {
			t.Fatalf("Resolve(%v) found nothing", in)
		}
		if got != a || got.String() != "19.0.1.0.0" {
			t.Errorf("Resolve(%v) = %#v, want %#v", in, got, a)
		}
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"19.0.1.2.3",
		"0.0.0.0.0",
		"19.0.01.2.3",
		"19.0.+1.2.3",
		"19.0.-0.2.3",
		"19.0.0.0." + maxIntPlus,
		"dev-main",
		"dev-19.0",
		"dev-feature/x",
		"dev-a..b",
		" dev-main ",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := version.Parse(s)
		if err != nil {
			return
		}
		again, err := version.Parse(v.String())
		if err != nil {
			t.Fatalf("Parse(%q) ok, but Parse(String() = %q): %v", s, v.String(), err)
		}
		if again != v {
			t.Fatalf("Parse(%q) = %#v, Parse(%q) = %#v", s, v, v.String(), again)
		}
		if again.String() != v.String() {
			t.Fatalf("String() changed from %q to %q", v.String(), again.String())
		}
		if version.Compare(again, v) != 0 {
			t.Fatalf("Compare(%q, %q) != 0", again, v)
		}
	})
}
