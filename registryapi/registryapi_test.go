package registryapi_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/wimwenigerkind/odoopack-spec/registryapi"
)

const (
	hash   = "ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543"
	commit = "6b6c390a387c81ee604b602e49ca6c35646cc619"
)

func zip() registryapi.Version {
	return registryapi.Version{
		Version:   "19.0.1.2.3",
		Type:      registryapi.TypeZip,
		URL:       "https://registry.example.com/registry/v1/zipball/f1740dd9-ca0e-42fd-a63b-b553d3fd3882/" + hash,
		Shasum:    "sha256:" + hash,
		Reference: commit,
	}
}

func TestVersionValidate(t *testing.T) {
	valid := []struct {
		name string
		edit func(*registryapi.Version)
	}{
		{"zip", func(v *registryapi.Version) {}},
		{"dev version", func(v *registryapi.Version) { v.Version = "dev-main" }},
		{"zip without reference", func(v *registryapi.Version) { v.Reference = "" }},
		{"uppercase shasum", func(v *registryapi.Version) { v.Shasum = "sha256:" + strings.ToUpper(hash) }},
		{"git", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Shasum = "" }},
		{"git with shasum", func(v *registryapi.Version) { v.Type = registryapi.TypeGit }},
		{"zip sha256 object format reference", func(v *registryapi.Version) { v.Reference = hash }},
		{"git sha256 object format reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = hash }},
		{"http localhost", func(v *registryapi.Version) { v.URL = "http://localhost:6969/z" }},
		{"http LOCALHOST", func(v *registryapi.Version) { v.URL = "http://LOCALHOST/z" }},
		{"http 127.0.0.1", func(v *registryapi.Version) { v.URL = "http://127.0.0.1:6969/z" }},
		{"http ::1", func(v *registryapi.Version) { v.URL = "http://[::1]:6969/z" }},
		{"HTTPS scheme", func(v *registryapi.Version) { v.URL = "HTTPS://registry.example.com/z" }},
		{"depends", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{
				{Module: "acme_essentials", Package: "acme/essentials", Access: registryapi.AccessOK},
				{Module: "acme_warranty", Access: registryapi.AccessForbidden},
				{Module: "base", Access: registryapi.AccessExternal},
			}
		}},
	}
	for _, tc := range valid {
		v := zip()
		tc.edit(&v)
		if err := v.Validate(); err != nil {
			t.Errorf("%s: Validate() = %v, want nil", tc.name, err)
		}
	}

	invalid := []struct {
		name string
		edit func(*registryapi.Version)
	}{
		{"empty version", func(v *registryapi.Version) { v.Version = "" }},
		{"bad version", func(v *registryapi.Version) { v.Version = "1.0" }},
		{"non canonical version", func(v *registryapi.Version) { v.Version = " 19.0.1.2.3" }},
		{"empty type", func(v *registryapi.Version) { v.Type = "" }},
		{"unknown type", func(v *registryapi.Version) { v.Type = "tar" }},
		{"uppercase type", func(v *registryapi.Version) { v.Type = "ZIP" }},
		{"empty url", func(v *registryapi.Version) { v.URL = "" }},
		{"relative url", func(v *registryapi.Version) { v.URL = "/registry/v1/zipball/x" }},
		{"url without host", func(v *registryapi.Version) { v.URL = "https:///z" }},
		{"opaque url", func(v *registryapi.Version) { v.URL = "https:registry.example.com/z" }},
		{"http remote", func(v *registryapi.Version) { v.URL = "http://registry.example.com/z" }},
		{"http localhost lookalike", func(v *registryapi.Version) { v.URL = "http://localhost.example.com/z" }},
		{"http 127.0.0.2", func(v *registryapi.Version) { v.URL = "http://127.0.0.2/z" }},
		{"ftp", func(v *registryapi.Version) { v.URL = "ftp://registry.example.com/z" }},
		{"file", func(v *registryapi.Version) { v.URL = "file:///tmp/z.zip" }},
		{"unparsable url", func(v *registryapi.Version) { v.URL = "https://registry.example.com/%zz" }},
		{"zip without shasum", func(v *registryapi.Version) { v.Shasum = "" }},
		{"zip bare hex shasum", func(v *registryapi.Version) { v.Shasum = hash }},
		{"zip sha512 shasum", func(v *registryapi.Version) { v.Shasum = "sha512:" + hash + hash }},
		{"zip short shasum", func(v *registryapi.Version) { v.Shasum = "sha256:" + hash[:63] }},
		{"zip short reference", func(v *registryapi.Version) { v.Reference = commit[:7] }},
		{"zip uppercase reference", func(v *registryapi.Version) { v.Reference = strings.ToUpper(commit) }},
		{"git without reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = "" }},
		{"git branch reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = "main" }},
		{"git non hex reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = commit[:39] + "g" }},
		{"git uppercase reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = strings.ToUpper(commit) }},
		{"git bad shasum", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Shasum = "sha256:x" }},
		{"git long reference", func(v *registryapi.Version) { v.Type = registryapi.TypeGit; v.Reference = commit + "0" }},
		{"dep empty module", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{{Module: "", Access: registryapi.AccessExternal}}
		}},
		{"dep bad module", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{{Module: "acme-essentials", Access: registryapi.AccessExternal}}
		}},
		{"dep bad package", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{{Module: "acme_essentials", Package: "Acme/essentials", Access: registryapi.AccessOK}}
		}},
		{"dep empty access", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{{Module: "base", Access: ""}}
		}},
		{"dep unknown access", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{{Module: "base", Access: "OK"}}
		}},
		{"second dep bad", func(v *registryapi.Version) {
			v.Depends = []registryapi.Dep{
				{Module: "base", Access: registryapi.AccessExternal},
				{Module: "web", Access: "denied"},
			}
		}},
	}
	for _, tc := range invalid {
		v := zip()
		tc.edit(&v)
		err := v.Validate()
		if err == nil {
			t.Errorf("%s: Validate() = nil, want error", tc.name)
			continue
		}
		if !strings.HasPrefix(err.Error(), "registryapi: version ") {
			t.Errorf("%s: Validate() = %q, want prefix %q", tc.name, err, "registryapi: version ")
		}
	}
}

func TestAddonValidate(t *testing.T) {
	bad := zip()
	bad.Version = "19.0.1.2.4"
	bad.Type = "tar"
	worse := zip()
	worse.Version = "dev-x"
	worse.Shasum = ""

	tests := []struct {
		name      string
		addon     registryapi.Addon
		wantValid []string
		wantErrs  []string
	}{
		{
			name:      "valid",
			addon:     registryapi.Addon{Name: "acme/essentials", Versions: []registryapi.Version{zip()}},
			wantValid: []string{"19.0.1.2.3"},
		},
		{
			name:  "no versions",
			addon: registryapi.Addon{Name: "acme/essentials"},
		},
		{
			name:      "bad name",
			addon:     registryapi.Addon{Name: "Acme/essentials", Versions: []registryapi.Version{zip()}},
			wantValid: []string{"19.0.1.2.3"},
			wantErrs:  []string{"registryapi: addon: name:"},
		},
		{
			name:      "bad versions are all reported",
			addon:     registryapi.Addon{Name: "acme/essentials", Versions: []registryapi.Version{bad, zip(), worse}},
			wantValid: []string{"19.0.1.2.3"},
			wantErrs:  []string{`version "19.0.1.2.4"`, `version "dev-x"`},
		},
	}
	for _, tc := range tests {
		err := tc.addon.Validate()
		if (err != nil) != (len(tc.wantErrs) > 0) {
			t.Errorf("%s: Validate() = %v, want errors %q", tc.name, err, tc.wantErrs)
		}
		for _, want := range tc.wantErrs {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("%s: Validate() = %v, want it to contain %q", tc.name, err, want)
			}
		}

		valid, _ := tc.addon.ValidVersions()
		got := []string{}
		for _, v := range valid {
			got = append(got, v.Version)
		}
		if want := append([]string{}, tc.wantValid...); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: ValidVersions() = %q, want %q", tc.name, got, want)
		}
	}
}

func TestDecodeRegistryResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
		want registryapi.Addon
	}{
		{
			name: "zip version",
			body: `{"name":"acme/essentials","versions":[{"version":"dev-main","type":"zip","url":"http://localhost:6969/registry/v1/zipball/f1740dd9-ca0e-42fd-a63b-b553d3fd3882/ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543","shasum":"sha256:ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543","reference":"6b6c390a387c81ee604b602e49ca6c35646cc619"}]}`,
			want: registryapi.Addon{
				Name: "acme/essentials",
				Versions: []registryapi.Version{{
					Version:   "dev-main",
					Type:      "zip",
					URL:       "http://localhost:6969/registry/v1/zipball/f1740dd9-ca0e-42fd-a63b-b553d3fd3882/" + hash,
					Shasum:    "sha256:" + hash,
					Reference: commit,
				}},
			},
		},
		{
			// The registry resolves each module of the manifest's depends:
			// a visible addon gets its package and "ok", a hidden one only
			// "forbidden", a module no addon provides "external".
			name: "with depends",
			body: `{"name":"acme/health-check","versions":[{"version":"19.0.1.0.0","type":"zip","url":"https://registry.example.com/registry/v1/zipball/f1740dd9-ca0e-42fd-a63b-b553d3fd3882/ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543","shasum":"sha256:ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543","reference":"6b6c390a387c81ee604b602e49ca6c35646cc619","depends":[{"module":"base","access":"external"},{"module":"acme_essentials","package":"acme/essentials","access":"ok"},{"module":"acme_warranty","access":"forbidden"}]}]}`,
			want: registryapi.Addon{
				Name: "acme/health-check",
				Versions: []registryapi.Version{{
					Version:   "19.0.1.0.0",
					Type:      "zip",
					URL:       "https://registry.example.com/registry/v1/zipball/f1740dd9-ca0e-42fd-a63b-b553d3fd3882/" + hash,
					Shasum:    "sha256:" + hash,
					Reference: commit,
					Depends: []registryapi.Dep{
						{Module: "base", Access: "external"},
						{Module: "acme_essentials", Package: "acme/essentials", Access: "ok"},
						{Module: "acme_warranty", Access: "forbidden"},
					},
				}},
			},
		},
	}
	for _, tc := range tests {
		var a registryapi.Addon
		if err := json.Unmarshal([]byte(tc.body), &a); err != nil {
			t.Fatalf("%s: Unmarshal: %v", tc.name, err)
		}
		if !reflect.DeepEqual(a, tc.want) {
			t.Errorf("%s: decoded %#v, want %#v", tc.name, a, tc.want)
		}
		if err := a.Validate(); err != nil {
			t.Errorf("%s: Validate() = %v, want nil", tc.name, err)
		}

		// Encoding gives back the same bytes, so a server can send these
		// types without changing the wire.
		out, err := json.Marshal(a)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", tc.name, err)
		}
		if string(out) != tc.body {
			t.Errorf("%s: Marshal = %s, want %s", tc.name, out, tc.body)
		}
	}
}
