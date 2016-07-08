// Package assert contains functions for making assertions in unit tests
package assert

import (
<<<<<<< HEAD
=======
	"fmt"
	"path/filepath"
	"runtime"
>>>>>>> 12a5469... start on swarm services; move to glade
	"strings"
)

// TestingT is an interface which defines the methods of testing.T that are
// required by this package
type TestingT interface {
	Fatalf(string, ...interface{})
}

// Equal compare the actual value to the expected value and fails the test if
// they are not equal.
func Equal(t TestingT, actual, expected interface{}) {
	if expected != actual {
<<<<<<< HEAD
		t.Fatalf("Expected '%v' (%T) got '%v' (%T)", expected, expected, actual, actual)
=======
		fatal(t, fmt.Sprintf("Expected '%v' (%T) got '%v' (%T)", expected, expected, actual, actual))
>>>>>>> 12a5469... start on swarm services; move to glade
	}
}

//EqualStringSlice compares two slices and fails the test if they do not contain
// the same items.
func EqualStringSlice(t TestingT, actual, expected []string) {
	if len(actual) != len(expected) {
		t.Fatalf("Expected (length %d): %q\nActual (length %d): %q",
			len(expected), expected, len(actual), actual)
	}
	for i, item := range actual {
		if item != expected[i] {
			t.Fatalf("Slices differ at element %d, expected %q got %q",
				i, expected[i], item)
		}
	}
}

// NilError asserts that the error is nil, otherwise it fails the test.
func NilError(t TestingT, err error) {
	if err != nil {
<<<<<<< HEAD
		t.Fatalf("Expected no error, got: %s", err.Error())
=======
		fatal(t, fmt.Sprintf("Expected no error, got: %s", err.Error()))
>>>>>>> 12a5469... start on swarm services; move to glade
	}
}

// Error asserts that error is not nil, and contains the expected text,
// otherwise it fails the test.
func Error(t TestingT, err error, contains string) {
	if err == nil {
<<<<<<< HEAD
		t.Fatalf("Expected an error, but error was nil")
	}

	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("Expected error to contain '%s', got '%s'", contains, err.Error())
=======
		fatal(t, "Expected an error, but error was nil")
	}

	if !strings.Contains(err.Error(), contains) {
		fatal(t, fmt.Sprintf("Expected error to contain '%s', got '%s'", contains, err.Error()))
>>>>>>> 12a5469... start on swarm services; move to glade
	}
}

// Contains asserts that the string contains a substring, otherwise it fails the
// test.
func Contains(t TestingT, actual, contains string) {
	if !strings.Contains(actual, contains) {
<<<<<<< HEAD
		t.Fatalf("Expected '%s' to contain '%s'", actual, contains)
	}
}
=======
		fatal(t, fmt.Sprintf("Expected '%s' to contain '%s'", actual, contains))
	}
}

func fatal(t TestingT, msg string) {
	_, file, line, _ := runtime.Caller(2)
	t.Fatalf("%s:%d: %s", filepath.Base(file), line, msg)
}
>>>>>>> 12a5469... start on swarm services; move to glade
