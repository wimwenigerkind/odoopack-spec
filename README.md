# odoopack-spec

The rules the odoopack CLI and the odoopack registry share, so that neither side keeps its own copy. The package `version` holds the Odoo-aware version parsing, comparison and resolution.

An Odoo module version follows the convention `<series>.<major>.<minor>.<patch>` where the series is the two-part Odoo version:

```
19.0.1.2.3
└──┘ └───┘
series  module semver (major.minor.patch)
```

## Two kinds of version

- **Release**: five numeric components, e.g. `19.0.1.2.3`. A component is ASCII digits without a sign and without leading zeros (`0` itself is fine) and must fit in an `int`, so `19.0.01.0.0`, `19.0.+1.0.0` and `19.0.-0.0.0` are rejected.
- **Dev**: a branch pseudo-version, e.g. `dev-19.0` or `dev-main`. The branch name uses only `[A-Za-z0-9._/-]`, contains no `..`, does not start with `-` or `/`, and does not end with `/` or `.lock`.

Anything else is rejected. Leading and trailing whitespace is trimmed. Every accepted version has exactly one spelling, so `String()` returns what `Parse` accepts and two versions that compare equal print the same.

## Ordering

Releases sort by series first, then by module semver; dev versions sort below all releases and among themselves by branch name, byte by byte. Comparison of releases is numeric, so `18.0.10.0.0` is greater than `18.0.9.9.9`.

## Constraints (basic)

| Constraint | Matches |
| --- | --- |
| `19.0.1.2.3` | that exact release |
| `19.0` | the highest release in series 19.0 |
| `dev-19.0` | the dev version for that branch |

## Usage

```go
import "github.com/wimwenigerkind/odoopack-spec/version"

v, err := version.Parse("19.0.1.2.3")

c, _ := version.ParseConstraint("19.0")
best, ok := version.Resolve(c, versions) // highest 19.0 release
```

## Names

The package `name` holds the rules for package and module names.

- **Package**: `vendor/name` or `name`. Each part uses lowercase ASCII letters, digits, `_` and `-` and starts with a letter or digit, e.g. `acme/health-check`.
- **Module**: the Odoo module directory, lowercase ASCII letters, digits and `_`.

Both are at most 128 characters. A package maps to its module by turning `/` and `-` into `_`, so `acme/health-check` installs as `acme_health_check`. Two packages such as `a-b` and `a_b` map to the same module and cannot be installed side by side.

```go
import "github.com/wimwenigerkind/odoopack-spec/name"

if err := name.ValidPackage("acme/health-check"); err != nil {
	// reject
}
dir := name.ModuleName("acme/health-check") // acme_health_check
```

## Digests

The package `digest` holds the content digest of an archive, written `sha256:<64 hex characters>`. Only sha256 is supported. Upper case hex is accepted and normalized to lower case, so `String()` returns one spelling per digest and two digests of the same content compare equal with `==`.

```go
import "github.com/wimwenigerkind/odoopack-spec/digest"

d, err := digest.Parse("sha256:ece96e65881389c3ef6538ea26b9d7149e8e0af042c828d17283c6b76c633543")
err = d.Verify(file) // reports the expected and the actual digest on a mismatch
```

## Registry API

The package `registryapi` holds the response of `GET /registry/v1/addons/{name}` as `Addon`, `Version` and `Dep`, with the JSON field names the registry sends today. Decoding keeps every field as sent; `Validate` checks it:

- **Addon**: the name passes `name.ValidPackage` and every version is valid.
- **Version**: the version is canonical, `type` is `zip` or `git`, `url` is absolute `https` (`http` only for `localhost`, `127.0.0.1` and `::1`). A zip has a `shasum` that `digest.Parse` accepts; a git version has a `reference` that is a full commit id (40 lower case hex characters, or 64 for a repository in the sha256 object format), and a zip that sends a reference follows the same rule.
- **Dep**: the module passes `name.ValidModule`, the package (sent only when `access` is `ok`) passes `name.ValidPackage`, and `access` is `ok`, `forbidden` (the addon exists but the caller cannot see it) or `external` (no addon provides the module, e.g. `base`).

`Validate` reports the problems of all versions at once, each naming its version. `ValidVersions` returns the valid versions and the same problems, for a caller that skips a broken version instead of rejecting the addon.

```go
import "github.com/wimwenigerkind/odoopack-spec/registryapi"

var a registryapi.Addon
err := json.Unmarshal(body, &a)
versions, problems := a.ValidVersions()
```
