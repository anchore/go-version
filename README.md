# Versioning Library for Go

[![Validations](https://github.com/anchore/go-version/actions/workflows/validations.yaml/badge.svg)](https://github.com/anchore/go-version/actions/workflows/validations.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/anchore/go-version.svg)](https://pkg.go.dev/github.com/anchore/go-version)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/anchore/go-version.svg)](https://github.com/anchore/go-version)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://github.com/anchore/go-version/blob/main/LICENSE)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack)](https://anchore.com/slack)

This is an Anchore-maintained fork of [hashicorp/go-version](https://github.com/hashicorp/go-version).
It tracks upstream while keeping Anchore's additions: `||` (or) constraints, the `^` and `~`
constraint operators, JSON marshalling for `Version` and `Constraints`, and prerelease handling
that follows [semver spec item 11](https://semver.org/#spec-item-11).

go-version is a library for parsing versions and version constraints,
and verifying versions against a set of constraints. go-version
can sort a collection of versions properly, handles prerelease/beta
versions, can increment versions, etc.

Versions used with go-version must follow [SemVer](http://semver.org/).

## Installation and Usage

Package documentation can be found on
[Go Reference](https://pkg.go.dev/github.com/anchore/go-version).

Installation can be done with a normal `go get`:

```
$ go get github.com/anchore/go-version
```

#### Version Parsing and Comparison

```go
v1, err := version.NewVersion("1.2")
v2, err := version.NewVersion("1.5+metadata")

// Comparison example. There is also GreaterThan, Equal, and just
// a simple Compare that returns an int allowing easy >=, <=, etc.
if v1.LessThan(v2) {
    fmt.Printf("%s is less than %s", v1, v2)
}
```

#### Version Parsing and Comparison with Prefixes

The library also supports parsing versions with a custom prefix.
Using the `WithPrefix` option, you can specify a prefix to strip
before parsing the version.

Use `WithPrefix` when your input strings carry a known release prefix such as
`deployment-`, `controller-`, etc.

After parsing, the prefix is not part of the canonical version value. This
means the regular comparison methods such as `Compare`, `LessThan`, `Equal`,
and `GreaterThan` compare only the stripped version. If you compare versions
from different prefixes with these methods, the prefixes are ignored. If you
need to reject cross-prefix comparisons, inspect the parsed prefixes before
comparing the versions.

```go
v1, _ := version.NewVersion("deployment-v1.2.3-beta+metadata", version.WithPrefix("deployment-"))
v2, _ := version.NewVersion("deployment-v1.2.4", version.WithPrefix("deployment-"))

if v1.LessThan(v2) {
    fmt.Printf("%s (%s) is less than %s (%s)\n", v1, v1.Original(), v2, v2.Original())
    // Outputs: 1.2.3-beta+metadata (deployment-v1.2.3-beta+metadata) is less than 1.2.4 (deployment-v1.2.4)
}
```

#### Version Constraints

```go
v1, err := version.NewVersion("1.2")

// Constraints example.
constraints, err := version.NewConstraint(">= 1.0, < 1.4")
if constraints.Check(v1) {
	fmt.Printf("%s satisfies constraints %s", v1, constraints)
}
```

#### Version Sorting

```go
versionsRaw := []string{"1.1", "0.7.1", "1.4-beta", "1.4", "2"}
versions := make([]*version.Version, len(versionsRaw))
for i, raw := range versionsRaw {
    v, _ := version.NewVersion(raw)
    versions[i] = v
}

// After this, the versions are properly sorted
sort.Sort(version.Collection(versions))
```

#### Or (`||`) Constraints

```go
constraints, err := version.NewConstraint(">= 1.0, < 1.4 || >= 2.0")
```

#### Caret and Tilde Constraints

```go
// ^1.2.0 allows any 1.x version at or above 1.2.0
caret, err := version.NewConstraint("^1.2.0")

// ~1.2.0 allows any 1.2.x version at or above 1.2.0
tilde, err := version.NewConstraint("~1.2.0")
```

## Issues and Contributing

If you find an issue with this library, please report an issue. If you'd
like, we welcome any contributions. See [CONTRIBUTING.md](CONTRIBUTING.md)
for details.

Changes that also apply upstream are best sent to
[hashicorp/go-version](https://github.com/hashicorp/go-version) first, so this
fork can pick them up on the next sync.
