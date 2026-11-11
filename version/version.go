package version

import (
	"fmt"
	"strconv"
	"strings"
)

const devPrefix = "dev-"

type Kind int

const (
	KindRelease Kind = iota
	KindDev
)

// Version keeps no copy of its input: parsing accepts only the canonical
// spelling, so String rebuilds it and versions that compare equal print equally.
type Version struct {
	Kind   Kind
	Series [2]int
	Semver [3]int
	Branch string
}

func Parse(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("version: empty version")
	}

	if strings.HasPrefix(s, devPrefix) {
		branch := s[len(devPrefix):]
		if err := validateBranch(branch); err != nil {
			return Version{}, fmt.Errorf("version: invalid dev branch in %q: %w", s, err)
		}
		v := Version{Kind: KindDev, Branch: branch, Series: [2]int{-1, -1}}
		if series, ok := parseSeries(branch); ok {
			v.Series = series
		}
		return v, nil
	}

	parts := strings.Split(s, ".")
	if len(parts) != 5 {
		return Version{}, fmt.Errorf("version: %q must have 5 numeric components (series.major.minor.patch), e.g. 19.0.1.2.3", s)
	}
	var nums [5]int
	for i, p := range parts {
		n, err := parseComponent(p)
		if err != nil {
			return Version{}, fmt.Errorf("version: invalid component %q in %q: %w", p, s, err)
		}
		nums[i] = n
	}
	return Version{
		Kind:   KindRelease,
		Series: [2]int{nums[0], nums[1]},
		Semver: [3]int{nums[2], nums[3], nums[4]},
	}, nil
}

func parseComponent(p string) (int, error) {
	if p == "" {
		return 0, fmt.Errorf("empty component")
	}
	for i := 0; i < len(p); i++ {
		if p[i] < '0' || p[i] > '9' {
			return 0, fmt.Errorf("component must be ASCII digits only")
		}
	}
	if len(p) > 1 && p[0] == '0' {
		return 0, fmt.Errorf("leading zero")
	}
	// The digits are already checked, so the only error left is a value out of range for int.
	n, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// validateBranch applies a simple subset of git's rules for a branch name:
// the characters [A-Za-z0-9._/-], no "..", no leading "-" or "/", and no
// trailing "/" or ".lock". Branch names outside this subset are not
// expected in an Odoo addon repository.
func validateBranch(b string) error {
	if b == "" {
		return fmt.Errorf("empty branch")
	}
	for i := 0; i < len(b); i++ {
		c := b[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '.' && c != '_' && c != '/' && c != '-' {
			return fmt.Errorf("character %q not allowed, use [A-Za-z0-9._/-]", c)
		}
	}
	switch {
	case strings.Contains(b, ".."):
		return fmt.Errorf(`branch must not contain ".."`)
	case b[0] == '-' || b[0] == '/':
		return fmt.Errorf(`branch must not start with "-" or "/"`)
	case strings.HasSuffix(b, "/") || strings.HasSuffix(b, ".lock"):
		return fmt.Errorf(`branch must not end with "/" or ".lock"`)
	}
	return nil
}

func parseSeries(s string) ([2]int, bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return [2]int{}, false
	}
	a, errA := parseComponent(parts[0])
	b, errB := parseComponent(parts[1])
	if errA != nil || errB != nil {
		return [2]int{}, false
	}
	return [2]int{a, b}, true
}

func (v Version) IsRelease() bool { return v.Kind == KindRelease }

func (v Version) IsDev() bool { return v.Kind == KindDev }

func (v Version) SeriesString() string {
	if v.Series[0] < 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d", v.Series[0], v.Series[1])
}

func (v Version) String() string {
	if v.Kind == KindDev {
		return devPrefix + v.Branch
	}
	return fmt.Sprintf("%d.%d.%d.%d.%d", v.Series[0], v.Series[1], v.Semver[0], v.Semver[1], v.Semver[2])
}
