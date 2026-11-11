package version

func Resolve(c Constraint, versions []Version) (Version, bool) {
	var best Version
	found := false
	for _, v := range versions {
		if !c.Matches(v) {
			continue
		}
		if !found || Compare(v, best) > 0 {
			best = v
			found = true
		}
	}
	return best, found
}
