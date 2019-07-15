package main

import (
  "testing"
)

func test(t *testing.T, arg string, flag int, expected string, fail bool) {
  v, err := parse(arg)
  if err != nil {
    if fail {
      return
    }
    t.Fatal("got unexpected error:", err)
  }
  version := bump(v, flag)
  if expected != version {
    t.Fatalf("expected: '%s', actual: '%s'", expected, version)
  }
}

func TestParse(t *testing.T) {
  test(t, "0.0.0", major, "1.0.0", false)
  test(t, "0.0.0", minor, "0.1.0", false)
  test(t, "0.0.0", patch, "0.0.1", false)
  test(t, "1.1.1", major, "2.0.0", false)
  test(t, "1.1.1", minor, "1.2.0", false)
  test(t, "1.1.1", patch, "1.1.2", false)
  test(t, "20.23.400", major, "21.0.0", false)
  test(t, "20.23.400", minor, "20.24.0", false)
  test(t, "20.23.400", patch, "20.23.401", false)

  test(t, "v1.2.0-18-gac70419", minor, "1.3.0", false)

  test(t, "vv1.2.0-18-gac70419", minor, "", true)
  test(t, "v1.2.0+meta", minor, "", true)
}
