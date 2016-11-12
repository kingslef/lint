// Package testutil contains methods to test checkers.
package testutil

import (
	"fmt"

	"path/filepath"

	"strings"

	"regexp"

	"reflect"

	"github.com/sridharv/fakegopath"
	"github.com/surullabs/lint"
	"github.com/surullabs/lint/checkers"
)

// StaticCheckTest is a table-driven test for a checker.
type StaticCheckTest struct {
	// File is a src file to use instead of Content.
	File string
	// Content is the content of the created file.
	Content []byte
	// Checker is the checker to run on the package.
	Checker lint.Checker
	// Validate returns nil if err is what is expected.
	Validate func(err error) error
}

// Test runs the test for pkg.
func (s StaticCheckTest) Test(pkg string) error {
	checkers.Unload(pkg)
	tmp, err := fakegopath.NewTemporaryWithFiles(pkg, []fakegopath.SourceFile{
		{Src: s.File, Content: s.Content, Dest: filepath.Join(pkg, "file.go")},
	})
	if err != nil {
		return fmt.Errorf("failed to create temporary go path: %v", err)
	}
	defer tmp.Reset()
	return s.Validate(s.Checker.Check(pkg))
}

// Errorer is used to report Errors. testing.T can be used as an Errorer.
type Errorer interface {
	Error(args ...interface{})
}

// Test runs the provided StaticCheckTests for pkg. Errors are reported using
// Errorer.
func Test(t Errorer, pkg string, tests []StaticCheckTest) {
	for i, test := range tests {
		if err := test.Test(pkg); err != nil {
			t.Error("Check", i, err)
		}
	}
}

// SkippedErrors returns a function that skips all errors matching pattern and verifies
// there are no errors left.
func SkippedErrors(pattern string) func(error) error {
	return Skip(lint.RegexpMatch(pattern), NoError)
}

// Skip returns a function that skips errors using s and verifies the result using then
func Skip(s lint.Skipper, then func(error) error) func(err error) error {
	return func(err error) error {
		return then(lint.Skip(err, s))
	}
}

// NoError returns err. Use this if you expect the operation to have no error.
func NoError(err error) error { return err }

// HasSuffix returns a function that checks that an error is returned which has the provided suffix.
func HasSuffix(suffix string) func(err error) error {
	return func(err error) error {
		if err == nil {
			return fmt.Errorf("no error found when expecting error with suffix %s", suffix)
		}
		if !strings.HasSuffix(err.Error(), suffix) {
			return err
		}
		return nil
	}
}

// MatchesRegexp returns a function that checks that an error is return which matches the provided regexp.
func MatchesRegexp(re string) func(err error) error {
	return func(err error) error {
		if err == nil {
			return fmt.Errorf("no error found when expecting error matching RE %s", re)
		}
		if matches, matchErr := regexp.MatchString(re, err.Error()); matchErr != nil {
			return matchErr
		} else if !matches {
			return fmt.Errorf("error %v does not match re %s", err, re)
		}
		return nil
	}
}

// Contains contains
func Contains(str string) func(err error) error {
	return func(err error) error {
		if err == nil {
			return fmt.Errorf("no error found when expecting error containing %s", str)
		}
		if !strings.Contains(err.Error(), str) {
			return err
		}
		return nil
	}
}

// Arger holds the Args method that returns a list of command line arguments
type Arger interface {
	Args() []string
}

// ArgTest verifies that the expected command line arguments are returned
type ArgTest struct {
	A        Arger
	Expected []string
}

// Test runs ArgTest and returns an error if the arguments generated by a.A do not
// match a.Expected
func (a ArgTest) Test() error {
	args := a.A.Args()
	if !reflect.DeepEqual(args, a.Expected) {
		return fmt.Errorf("Expected: %v, got %v", a.Expected, args)
	}
	return nil
}

// TestArgs runs the provided ArgTests. Errors are reported using
// Errorer.
func TestArgs(t Errorer, tests []ArgTest) {
	for i, test := range tests {
		if err := test.Test(); err != nil {
			t.Error("Args", i, err)
		}
	}
}

// SkipTest is a table driven test for skippers
type SkipTest struct {
	S    lint.Skipper
	Line string
	Skip bool
}

// Test returns an error if the s.S.Skip(s.Line) != s.Skip
func (s SkipTest) Test() error {
	if s.S.Skip(s.Line) != s.Skip {
		return fmt.Errorf("expected %v for %s", s.Skip, s.Line)
	}
	return nil
}

// TestSkips runs the provided SkipTests. Errors are reported using
// Errorer.
func TestSkips(t Errorer, tests []SkipTest) {
	for i, test := range tests {
		if err := test.Test(); err != nil {
			t.Error("Skips", i, err)
		}
	}
}
