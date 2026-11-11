package registryapi

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/wimwenigerkind/odoopack-spec/digest"
	"github.com/wimwenigerkind/odoopack-spec/name"
	"github.com/wimwenigerkind/odoopack-spec/version"
)

const (
	TypeZip = "zip"
	TypeGit = "git"
)

// The registry resolves each dependency against the addons the caller can
// see: ok names the package, forbidden means it exists but is hidden from the
// caller, external means no addon provides the module (e.g. Odoo's base).
const (
	AccessOK        = "ok"
	AccessForbidden = "forbidden"
	AccessExternal  = "external"
)

// Addon is the response of GET /registry/v1/addons/{name}. The fields stay
// plain strings so a response decodes as sent; Validate checks them.
type Addon struct {
	Name     string    `json:"name"`
	Versions []Version `json:"versions"`
}

type Version struct {
	Version   string `json:"version"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Shasum    string `json:"shasum,omitempty"`
	Reference string `json:"reference,omitempty"`
	Depends   []Dep  `json:"depends,omitempty"`
}

type Dep struct {
	Module  string `json:"module"`
	Package string `json:"package,omitempty"`
	Access  string `json:"access"`
}

func (a Addon) Validate() error {
	var errs []error
	if err := name.ValidPackage(a.Name); err != nil {
		errs = append(errs, fmt.Errorf("registryapi: addon: %w", err))
	}
	_, err := a.ValidVersions()
	return errors.Join(append(errs, err)...)
}

// ValidVersions returns the versions that pass Validate and the problems of
// the others joined, so a caller can skip a broken version instead of failing
// the whole addon. It does not check the addon name.
func (a Addon) ValidVersions() ([]Version, error) {
	valid := make([]Version, 0, len(a.Versions))
	var errs []error
	for _, v := range a.Versions {
		if err := v.Validate(); err != nil {
			errs = append(errs, err)
			continue
		}
		valid = append(valid, v)
	}
	return valid, errors.Join(errs...)
}

func (v Version) Validate() error {
	if err := v.validate(); err != nil {
		return fmt.Errorf("registryapi: version %q: %w", v.Version, err)
	}
	return nil
}

func (v Version) validate() error {
	parsed, err := version.Parse(v.Version)
	if err != nil {
		return err
	}
	// Parse trims whitespace; the wire value must already be canonical
	// because callers use it as a key.
	if parsed.String() != v.Version {
		return fmt.Errorf("version %q is not canonical, want %q", v.Version, parsed.String())
	}
	if err := validURL(v.URL); err != nil {
		return err
	}
	switch v.Type {
	case TypeZip:
		if _, err := digest.Parse(v.Shasum); err != nil {
			return fmt.Errorf("shasum: %w", err)
		}
		if v.Reference != "" && !isCommit(v.Reference) {
			return fmt.Errorf("reference %q must be a full commit id, 40 or 64 lowercase hex characters", v.Reference)
		}
	case TypeGit:
		if !isCommit(v.Reference) {
			return fmt.Errorf("reference %q must be a full commit id, 40 or 64 lowercase hex characters", v.Reference)
		}
		if v.Shasum != "" {
			if _, err := digest.Parse(v.Shasum); err != nil {
				return fmt.Errorf("shasum: %w", err)
			}
		}
	default:
		return fmt.Errorf("type %q must be %q or %q", v.Type, TypeZip, TypeGit)
	}
	for _, d := range v.Depends {
		if err := d.validate(); err != nil {
			return fmt.Errorf("dependency %q: %w", d.Module, err)
		}
	}
	return nil
}

func (d Dep) validate() error {
	if err := name.ValidModule(d.Module); err != nil {
		return err
	}
	if d.Package != "" {
		if err := name.ValidPackage(d.Package); err != nil {
			return err
		}
	}
	switch d.Access {
	case AccessOK, AccessForbidden, AccessExternal:
		return nil
	}
	return fmt.Errorf("access %q must be %q, %q or %q", d.Access, AccessOK, AccessForbidden, AccessExternal)
}

// validURL allows plain http only for a registry on the local machine, so a
// development setup works without a certificate.
func validURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("url %q: %w", s, err)
	}
	if !u.IsAbs() || u.Host == "" {
		return fmt.Errorf("url %q must be absolute", s)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		switch strings.ToLower(u.Hostname()) {
		case "localhost", "127.0.0.1", "::1":
			return nil
		}
		return fmt.Errorf("url %q must use https, http is only allowed for localhost", s)
	}
	return fmt.Errorf("url %q must use https", s)
}

// isCommit accepts a full commit id as git prints it: 40 hex characters for
// sha1 repositories, 64 for repositories in the sha256 object format.
func isCommit(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
