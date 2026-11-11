package name

import (
	"fmt"
	"strings"
)

// MaxLen caps package and module names alike. Odoo installs a module into a
// directory of that name and the registry stores archives under it, so the cap
// stays well below the 255 byte file name limit of common file systems while
// leaving room for any name in use. ModuleName keeps the length, so a valid
// package always maps to a module name within the cap.
const MaxLen = 128

var moduleReplacer = strings.NewReplacer("/", "_", "-", "_")

func ValidPackage(s string) error {
	if s == "" {
		return fmt.Errorf("name: empty package name")
	}
	if len(s) > MaxLen {
		return fmt.Errorf("name: package name %q is longer than %d characters", s, MaxLen)
	}
	parts := strings.Split(s, "/")
	if len(parts) > 2 {
		return fmt.Errorf("name: package name %q must be vendor/name or name", s)
	}
	for _, p := range parts {
		if err := validPart(p); err != nil {
			return fmt.Errorf("name: package name %q: %w", s, err)
		}
	}
	return nil
}

func validPart(p string) error {
	if p == "" {
		return fmt.Errorf("empty part")
	}
	for i, r := range p {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case (r == '_' || r == '-') && i > 0:
		case r == '_' || r == '-':
			return fmt.Errorf("part %q must start with a lowercase letter or digit", p)
		default:
			return fmt.Errorf("invalid character %q in %q, only lowercase letters, digits, _ and - are allowed", r, p)
		}
	}
	return nil
}

// ModuleName maps a package name to the directory name of its Odoo module.
// It expects a name that passes ValidPackage; other input is not rejected but
// may map to a path that is not a single directory.
func ModuleName(pkg string) string {
	return moduleReplacer.Replace(pkg)
}

func ValidModule(s string) error {
	if s == "" {
		return fmt.Errorf("name: empty module name")
	}
	if len(s) > MaxLen {
		return fmt.Errorf("name: module name %q is longer than %d characters", s, MaxLen)
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
		default:
			return fmt.Errorf("name: invalid character %q in module name %q, only lowercase letters, digits and _ are allowed", r, s)
		}
	}
	return nil
}
