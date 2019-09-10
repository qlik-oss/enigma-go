package main

import (
  "bytes"
  "io/ioutil"
  "testing"
)

func TestParseValid(t *testing.T) {
  in, err := ioutil.ReadFile("./valid.in")
  if err != nil {
    t.Error(err.Error())
  }
  keys := []string{"major", "minor", "patch"}
  lines := bytes.Split(in, []byte{'\n'})
  for _, line := range lines {
    if len(line) == 0 {
      continue
    }
    str := string(line)
    v := parse(str)
    for _, k := range keys {
      if v[k] == "" {
        t.Errorf("No '%s' for valid version '%s'", k, str)
      }
    }
  }
}

func TestParseInvalid(t *testing.T) {
  in, err := ioutil.ReadFile("./invalid.in")
  if err != nil {
    t.Error(err.Error())
  }
  keys := []string{"major", "minor", "patch"}
  lines := bytes.Split(in, []byte{'\n'})
  for _, line := range lines {
    if len(line) == 0 {
      continue
    }
    str := string(line)
    v := parse(str)
    all := true
    for _, k := range keys {
      _, ok := v[k]
      all = all && ok
    }
    if all {
      t.Errorf("Invalid version '%s' had all M.m.p set", str)
    }
  }
}
