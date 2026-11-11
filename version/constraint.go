package version

import (
	"fmt"
	"strings"
)

type constraintKind int

const (
	cExact constraintKind = iota
	cSeries
	cDev
)

type Constraint struct {
	kind   constraintKind
	exact  Version
	series [2]int
	branch string
	raw    string
}

func ParseConstraint(s string) (Constraint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Constraint{}, fmt.Errorf("odoosemver: empty constraint")
	}

	if strings.HasPrefix(s, devPrefix) {
		branch := s[len(devPrefix):]
		if branch == "" {
			return Constraint{}, fmt.Errorf("odoosemver: empty dev branch in %q", s)
		}
		return Constraint{kind: cDev, branch: branch, raw: s}, nil
	}

	switch strings.Count(s, ".") {
	case 1:
		series, ok := parseSeries(s)
		if !ok {
			return Constraint{}, fmt.Errorf("odoosemver: invalid series constraint %q", s)
		}
		return Constraint{kind: cSeries, series: series, raw: s}, nil
	case 4:
		v, err := Parse(s)
		if err != nil {
			return Constraint{}, err
		}
		return Constraint{kind: cExact, exact: v, raw: s}, nil
	default:
		return Constraint{}, fmt.Errorf("odoosemver: constraint %q must be an exact version (a.b.c.d.e), a series (a.b), or dev-<branch>", s)
	}
}

func (c Constraint) Matches(v Version) bool {
	switch c.kind {
	case cExact:
		return v.Kind == KindRelease && Compare(v, c.exact) == 0
	case cSeries:
		return v.Kind == KindRelease && v.Series == c.series
	case cDev:
		return v.Kind == KindDev && v.Branch == c.branch
	}
	return false
}

func (c Constraint) String() string {
	return c.raw
}
