package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestParseValid(t *testing.T) {
	in, err := ioutil.ReadFile("./test/valid.in")
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
		v, err := parse(str)
		if err != nil {
			t.Error(err.Error())
		}
		for _, k := range keys {
			if v[k] == "" {
				t.Errorf("No '%s' for valid version '%s'", k, str)
			}
		}
	}
}

func TestParseInvalid(t *testing.T) {
	in, err := ioutil.ReadFile("./test/invalid.in")
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
		v, err := parse(str)
		if err == nil {
			t.Error("Expected error but was nil.")
		}
		all := true
		for _, k := range keys {
			_, ok := v[k]
			all = all && ok
		}
		if all {
			t.Errorf("Invalid version '%s' had all M.m.p set", str)
		}
	}
	b := version{"major": ""}.valid()
	b = b || version{"major": "1", "minor": ""}.valid()
	b = b || version{"major": "1", "minor": "1", "patch": ""}.valid()
	b = !b && version{"major": "1", "minor": "1", "patch": "1"}.valid()
	if !b {
		t.Error("Expected 'b' to be true.")
	}
}

func TestCompare(t *testing.T) {
	tests := map[[2]string]int{
		[2]string{"2.0.0", "1.0.0"}:            1,
		[2]string{"1.0.0", "2.0.0"}:            -1,
		[2]string{"1.9.0", "1.0.0"}:            1,
		[2]string{"1.0.0", "1.10.0"}:           -1,
		[2]string{"1.0.1", "1.0.0"}:            1,
		[2]string{"1.0.0", "1.0.1"}:            -1,
		[2]string{"1.0.0", "1.0.0-beta"}:       1,
		[2]string{"1.0.0-beta", "1.0.0"}:       -1,
		[2]string{"1.0.0-beta", "1.0.0-alpha"}: 0,
		[2]string{"1.0.0-rc", "1.0.0"}:         -1,
		[2]string{"1.0.1", "1.0.0"}:            1,
		[2]string{"1.0.0", "1.0.1"}:            -1,
		[2]string{"1.0.1", "1.1.0"}:            -1,
		[2]string{"2.0.0", "1.1.0"}:            1,
		[2]string{"2.1.0", "2.1.1"}:            -1,
		[2]string{"2.1.1-hello", "2.1.1"}:      -1,
	}
	for test, res := range tests {
		v, _ := parse(test[0])
		w, _ := parse(test[1])
		c := compare(v, w)
		if c != res {
			t.Errorf("compare '%s' vs. '%s' => '%d' expected '%d'",
				test[0], test[1], c, res)
		}
	}
}

func TestBump(t *testing.T) {
	v, _ := parse("v1.9.7-beta+alfa")
	w, _ := parse("2.1.1")
	err := v.bump("foo")
	exp := "invalid field"
	if !strings.Contains(err.Error(), exp) {
		t.Errorf("error did not contain substring '%s'", exp)
	}
	err = v.bump("buildmetadata")
	exp = "cannot bump"
	if !strings.Contains(err.Error(), exp) {
		t.Errorf("error did not contain substring '%s'", exp)
	}
	bumps := []string{"minor", "major", "patch", "patch", "minor", "patch"}
	for _, b := range bumps {
		_ = v.bump(b)
	}
	if compare(v, w) != 0 {
		t.Errorf("expected '%s' to be equal to '%s'", v.String(), w.String())
	}
}

func TestString(t *testing.T) {
	tests := []string{
		"1.0.0",
		"0.5.0-bro",
		"0.0.1-rc+meta",
	}

	for _, test := range tests {
		v, err := parse(test)
		if err != nil {
			t.Error(err.Error())
		}
		if v.String() != test {
			t.Errorf("expected '%s' got '%s'", test, v.String())
		}
	}
}

func TestRun(t *testing.T) {
	pass := [][]string{
		[]string{"1.0.1", "1.0.1-beta", "1"},
		[]string{"1.9.0", "1.9.0", "0"},
		[]string{"2.0.0", "1.9.1-beta", "1"},
		[]string{"2.0.0", "2.1.0-beta+alfa", "-1"},
	}
	fail := [][]string{
		[]string{},
		[]string{"1.0.0"},
		[]string{"1.9.1-beta"},
		[]string{"1"},
		[]string{"1", "1.0.0"},
		[]string{"bump", "1.0.0"},
		[]string{"bump", "major"},
		[]string{"bump", "major", "1"},
		[]string{"bump", "major", "1.0.0", "1.0.0"},
	}
	cmd := exec.Command("go", "build")
	err := cmd.Run()
	if err != nil {
		t.Error("error when building executable: " + err.Error())
	}

	bin := "./verval"

	for _, args := range pass {
		cmd := exec.Command(bin, args[:2]...)
		out, err := cmd.Output()
		if err != nil {
			t.Error(err.Error())
		}
		if err != nil {
			t.Error(err.Error())
		}
		if res := string(out); res != args[2] {
			t.Errorf("expected '%s' but got '%s'", args[2], res)
		}
	}
	for _, args := range fail {
		cmd := exec.Command(bin, args...)
		err := cmd.Run()
		if err == nil {
			t.Errorf("cmd: %v passed but should've failed.", args)
		}
	}

	bumps := [][3]string{
		[3]string{"minor", "v1.0.0", "1.1.0"},
		[3]string{"minor", "0.5.7", "0.6.0"},
		[3]string{"minor", "1.7.9-beta+asdasd", "1.8.0"},
		[3]string{"patch", "0.5.7", "0.5.8"},
		[3]string{"major", "0.5.7", "1.0.0"},
	}
	for _, test := range bumps {
		cmd := exec.Command(bin, "bump", test[0], test[1])
		out, err := cmd.Output()
		if err != nil {
			t.Error(err.Error())
		}
		res := string(out)
		if res != test[2] {
			t.Errorf("expected '%s' to be '%s'", res, test[2])
		}
	}
	os.Remove(bin)
}
