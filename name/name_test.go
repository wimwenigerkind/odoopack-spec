package name

import (
	"strings"
	"testing"
)

func TestValidPackage(t *testing.T) {
	valid := []string{
		"acme/essentials",
		"acme/health-check",
		"acme_publisher_warranty_offline",
		"a",
		"a/b",
		"0vendor/1name",
		"a-/b_",
		strings.Repeat("a", MaxLen),
		strings.Repeat("a", MaxLen/2) + "/" + strings.Repeat("b", MaxLen/2-1),
	}
	for _, s := range valid {
		if err := ValidPackage(s); err != nil {
			t.Errorf("ValidPackage(%q) = %v, want nil", s, err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"a/../b",
		"../a",
		"/abs",
		"/a",
		"a/",
		"a//b",
		"a/b/c",
		`a\b`,
		`vendor\name`,
		"Acme/essentials",
		"acme/Essentials",
		"a b",
		" a",
		"a\n",
		"-a",
		"_a",
		"vendor/-a",
		"vendor/_a",
		"a.b",
		strings.Repeat("a", MaxLen+1),
		"café",
		"ä/b",
		"а",
	}
	for _, s := range invalid {
		if err := ValidPackage(s); err == nil {
			t.Errorf("ValidPackage(%q) = nil, want error", s)
		}
	}
}

func TestModuleName(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{"acme/essentials", "acme_essentials"},
		{"acme/health-check", "acme_health_check"},
		{"acme_publisher_warranty_offline", "acme_publisher_warranty_offline"},
	}
	for _, tt := range tests {
		if got := ModuleName(tt.pkg); got != tt.want {
			t.Errorf("ModuleName(%q) = %q, want %q", tt.pkg, got, tt.want)
		}
		if err := ValidModule(ModuleName(tt.pkg)); err != nil {
			t.Errorf("ValidModule(ModuleName(%q)) = %v, want nil", tt.pkg, err)
		}
	}
}

// Distinct packages can share a module directory; the registry relies on
// equal module names to detect such collisions.
func TestModuleNameCollision(t *testing.T) {
	pairs := [][2]string{
		{"a-b", "a_b"},
		{"vendor/a-b", "vendor_a_b"},
	}
	for _, p := range pairs {
		if ModuleName(p[0]) != ModuleName(p[1]) {
			t.Errorf("ModuleName(%q) = %q, ModuleName(%q) = %q, want equal", p[0], ModuleName(p[0]), p[1], ModuleName(p[1]))
		}
	}
}

func TestValidModule(t *testing.T) {
	valid := []string{
		"acme_essentials",
		"a",
		"0",
		"_a",
		strings.Repeat("a", MaxLen),
	}
	for _, s := range valid {
		if err := ValidModule(s); err != nil {
			t.Errorf("ValidModule(%q) = %v, want nil", s, err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"a/b",
		"a-b",
		`a\b`,
		"A",
		"a b",
		"café",
		strings.Repeat("a", MaxLen+1),
	}
	for _, s := range invalid {
		if err := ValidModule(s); err == nil {
			t.Errorf("ValidModule(%q) = nil, want error", s)
		}
	}
}
