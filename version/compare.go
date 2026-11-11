package version

import "strings"

func Compare(a, b Version) int {
	if a.Kind != b.Kind {
		if a.Kind == KindRelease {
			return 1
		}
		return -1
	}
	if a.Kind == KindDev {
		return strings.Compare(a.Branch, b.Branch)
	}
	for i := range 2 {
		if c := cmpInt(a.Series[i], b.Series[i]); c != 0 {
			return c
		}
	}
	for i := range 3 {
		if c := cmpInt(a.Semver[i], b.Semver[i]); c != 0 {
			return c
		}
	}
	return 0
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
