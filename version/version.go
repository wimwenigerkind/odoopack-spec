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

type Version struct {
	Kind   Kind
	Series [2]int
	Semver [3]int
	Branch string
	raw    string
}

func Parse(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("odoosemver: empty version")
	}

	if strings.HasPrefix(s, devPrefix) {
		branch := s[len(devPrefix):]
		if branch == "" {
			return Version{}, fmt.Errorf("odoosemver: empty dev branch in %q", s)
		}
		v := Version{Kind: KindDev, Branch: branch, raw: s, Series: [2]int{-1, -1}}
		if series, ok := parseSeries(branch); ok {
			v.Series = series
		}
		return v, nil
	}

	parts := strings.Split(s, ".")
	if len(parts) != 5 {
		return Version{}, fmt.Errorf("odoosemver: %q must have 5 numeric components (series.major.minor.patch), e.g. 19.0.1.2.3", s)
	}
	var nums [5]int
	for i, p := range parts {
		n, err := parseComponent(p)
		if err != nil {
			return Version{}, fmt.Errorf("odoosemver: invalid component %q in %q: %w", p, s, err)
		}
		nums[i] = n
	}
	return Version{
		Kind:   KindRelease,
		Series: [2]int{nums[0], nums[1]},
		Semver: [3]int{nums[2], nums[3], nums[4]},
		raw:    s,
	}, nil
}

func parseComponent(p string) (int, error) {
	if p == "" {
		return 0, fmt.Errorf("empty component")
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative component")
	}
	return n, nil
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
	if v.raw != "" {
		return v.raw
	}
	if v.Kind == KindDev {
		return devPrefix + v.Branch
	}
	return fmt.Sprintf("%d.%d.%d.%d.%d", v.Series[0], v.Series[1], v.Semver[0], v.Semver[1], v.Semver[2])
}
