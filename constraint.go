// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	constraintRegexp     *regexp.Regexp
	constraintRegexpOnce sync.Once
)

func getConstraintRegexp() *regexp.Regexp {
	constraintRegexpOnce.Do(func() {
		// This heavy lifting only happens the first time this function is called
		constraintRegexp = regexp.MustCompile(fmt.Sprintf(
			`^\s*(%s)\s*(%s)\s*$`,
			`<=|>=|!=|~>|\^|~|<|>|=|`,
			VersionRegexpRaw,
		))
	})
	return constraintRegexp
}

// Constraint represents a single constraint for a version, such as
// ">= 1.0".
type Constraint struct {
	f        constraintFunc
	op       operator
	check    *Version
	original string
}

func (c *Constraint) Equals(con *Constraint) bool {
	return c.op == con.op && c.check.Equal(con.check)
}

// Constraints is a 2D slice of constraints. The outer slice is a
// disjunction (||) of conjunctions (,) of constraints.
type Constraints [][]*Constraint

type constraintFunc func(v, c *Version) bool

type constraintOperation struct {
	op operator
	f  constraintFunc
}

// NewConstraint will parse one or more constraints from the given
// constraint string. The string must be a comma or pipe separated
// list of constraints.
func NewConstraint(cs string) (Constraints, error) {
	ors := strings.Split(cs, "||")
	or := make([][]*Constraint, len(ors))
	for k, v := range ors {
		vs := strings.Split(v, ",")
		result := make([]*Constraint, len(vs))
		for i, single := range vs {
			c, err := parseSingle(single)
			if err != nil {
				return nil, err
			}

			result[i] = c
		}
		or[k] = result
	}

	return Constraints(or), nil
}

// MustConstraints is a helper that wraps a call to a function
// returning (Constraints, error) and panics if error is non-nil.
func MustConstraints(c Constraints, err error) Constraints {
	if err != nil {
		panic(err)
	}

	return c
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v *Version) bool {
	for _, o := range cs {
		ok := true
		for _, c := range o {
			if !c.Check(v) {
				ok = false
				break
			}
		}

		if ok {
			return true
		}
	}

	return false
}

// Equals compares Constraints with other Constraints
// for equality. This may not represent logical equivalence
// of compared constraints.
// e.g. even though '>0.1,>0.2' is logically equivalent
// to '>0.2' it is *NOT* treated as equal.
//
// Missing operator is treated as equal to '=', whitespaces
// are ignored and constraints are sorted before comparison.
func (cs Constraints) Equals(c Constraints) bool {
	if len(cs) != len(c) {
		return false
	}

	// make copies to retain order of the original slices
	left := cs.sorted()
	right := c.sorted()

	// compare sorted slices
	for i, group := range left {
		if len(group) != len(right[i]) {
			return false
		}

		for j, con := range group {
			if !con.Equals(right[i][j]) {
				return false
			}
		}
	}

	return true
}

// sorted returns a copy of the constraints with each conjunction group
// sorted, and the groups themselves sorted, so that two logically
// identical sets of constraints compare element by element.
func (cs Constraints) sorted() Constraints {
	out := make(Constraints, len(cs))
	for i, group := range cs {
		g := make([]*Constraint, len(group))
		copy(g, group)
		sort.Stable(constraintGroup(g))
		out[i] = g
	}
	sort.Stable(out)

	return out
}

// constraintGroup is a single conjunction (comma separated) of constraints.
type constraintGroup []*Constraint

func (g constraintGroup) Len() int { return len(g) }

func (g constraintGroup) Less(i, j int) bool {
	if g[i].op != g[j].op {
		return g[i].op < g[j].op
	}

	return g[i].check.LessThan(g[j].check)
}

func (g constraintGroup) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

// Len, Less and Swap sort the disjunction (||) groups. Use sorted to order
// the constraints within each group as well.
func (cs Constraints) Len() int {
	return len(cs)
}

func (cs Constraints) Less(i, j int) bool {
	a, b := cs[i], cs[j]
	for k := 0; k < len(a) && k < len(b); k++ {
		if a[k].op != b[k].op {
			return a[k].op < b[k].op
		}
		if !a[k].check.Equal(b[k].check) {
			return a[k].check.LessThan(b[k].check)
		}
	}

	return len(a) < len(b)
}

func (cs Constraints) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

// Returns the string format of the constraints
func (cs Constraints) String() string {
	orStr := make([]string, len(cs))
	for i, o := range cs {
		csStr := make([]string, len(o))
		for j, c := range o {
			csStr[j] = c.String()
		}

		orStr[i] = strings.Join(csStr, ",")
	}

	return strings.Join(orStr, "||")
}

// Check tests if a constraint is validated by the given version.
func (c *Constraint) Check(v *Version) bool {
	return c.f(v, c.check)
}

// Prerelease returns true if the version underlying this constraint
// contains a prerelease field.
func (c *Constraint) Prerelease() bool {
	return len(c.check.Prerelease()) > 0
}

func (c *Constraint) String() string {
	return c.original
}

func parseSingle(v string) (*Constraint, error) {
	matches := getConstraintRegexp().FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("malformed constraint: %s", v)
	}

	check, err := NewVersion(matches[2])
	if err != nil {
		return nil, err
	}

	var cop constraintOperation
	switch matches[1] {
	case "=":
		cop = constraintOperation{op: equal, f: constraintEqual}
	case "!=":
		cop = constraintOperation{op: notEqual, f: constraintNotEqual}
	case ">":
		cop = constraintOperation{op: greaterThan, f: constraintGreaterThan}
	case "<":
		cop = constraintOperation{op: lessThan, f: constraintLessThan}
	case ">=":
		cop = constraintOperation{op: greaterThanEqual, f: constraintGreaterThanEqual}
	case "<=":
		cop = constraintOperation{op: lessThanEqual, f: constraintLessThanEqual}
	case "~>":
		cop = constraintOperation{op: pessimistic, f: constraintPessimistic}
	case "^":
		cop = constraintOperation{op: caret, f: constraintCaret}
	case "~":
		cop = constraintOperation{op: tilde, f: constraintTilde}
	default:
		cop = constraintOperation{op: equal, f: constraintEqual}
	}

	return &Constraint{
		f:        cop.f,
		op:       cop.op,
		check:    check,
		original: v,
	}, nil
}

func prereleaseCheck(v, c *Version) bool {
	switch vPre, cPre := v.Prerelease() != "", c.Prerelease() != ""; {
	case cPre && vPre:
		// A constraint with a pre-release can only match a pre-release version
		// with the same base segments.
		return v.equalSegments(c)
	case !cPre && vPre:
		// OK, per https://semver.org/#spec-item-11 (#3)
	case cPre && !vPre:
		// OK, except with the pessimistic operator
	case !cPre && !vPre:
		// OK
	}
	return true
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

type operator rune

const (
	equal            operator = '='
	notEqual         operator = '≠'
	greaterThan      operator = '>'
	lessThan         operator = '<'
	greaterThanEqual operator = '≥'
	lessThanEqual    operator = '≤'
	pessimistic      operator = '~'
	caret            operator = '^'
	tilde            operator = '≈'
)

func constraintEqual(v, c *Version) bool {
	return v.Equal(c)
}

func constraintNotEqual(v, c *Version) bool {
	return !v.Equal(c)
}

func constraintGreaterThan(v, c *Version) bool {
	return v.Compare(c) == 1
}

func constraintLessThan(v, c *Version) bool {
	return v.Compare(c) == -1
}

func constraintGreaterThanEqual(v, c *Version) bool {
	return v.Compare(c) >= 0
}

func constraintLessThanEqual(v, c *Version) bool {
	return v.Compare(c) <= 0
}

func constraintPessimistic(v, c *Version) bool {
	// Using a pessimistic constraint with a pre-release, restricts versions to pre-releases
	if !prereleaseCheck(v, c) || (c.Prerelease() != "" && v.Prerelease() == "") {
		return false
	}

	// If the version being checked is naturally less than the constraint, then there
	// is no way for the version to be valid against the constraint
	if v.LessThan(c) {
		return false
	}
	// We'll use this more than once, so grab the length now so it's a little cleaner
	// to write the later checks
	cs := len(c.segments)

	// If the version being checked has less specificity than the constraint, then there
	// is no way for the version to be valid against the constraint
	if cs > len(v.segments) {
		return false
	}

	// Check the segments in the constraint against those in the version. If the version
	// being checked, at any point, does not have the same values in each index of the
	// constraints segments, then it cannot be valid against the constraint.
	for i := 0; i < c.si-1; i++ {
		if v.segments[i] != c.segments[i] {
			return false
		}
	}

	// Check the last part of the segment in the constraint. If the version segment at
	// this index is less than the constraints segment at this index, then it cannot
	// be valid against the constraint
	if c.segments[cs-1] > v.segments[cs-1] {
		return false
	}

	// If nothing has rejected the version by now, it's valid
	return true
}

func constraintCaret(v, c *Version) bool {
	if !prereleaseCheck(v, c) || v.LessThan(c) {
		return false
	}

	if v.segments[0] != c.segments[0] {
		return false
	}

	return true
}

func constraintTilde(v, c *Version) bool {
	if !prereleaseCheck(v, c) || v.LessThan(c) {
		return false
	}

	if v.segments[0] != c.segments[0] {
		return false
	}

	if v.segments[1] != c.segments[1] && c.si > 1 {
		return false
	}

	return true
}

// MarshalJSON implements the encoding/json.Marshaler interface.
func (cs *Constraints) MarshalJSON() ([]byte, error) {
	return json.Marshal(cs.String())
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface.
func (cs *Constraints) UnmarshalJSON(data []byte) error {
	var constraintStr string
	if err := json.Unmarshal(data, &constraintStr); err != nil {
		return err
	}

	nc, err := NewConstraint(constraintStr)
	if err != nil {
		return err
	}
	*cs = nc

	return nil
}
