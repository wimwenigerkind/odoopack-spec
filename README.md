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
