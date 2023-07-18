package main

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var updateSpecGoldenfile = flag.Bool("update", false, "updateSpecGoldenfile golden files")

func Test(t *testing.T) {

	info := &info{
		Name:                "mockapi",
		GoPackageImportPath: "github.com/qlik-oss/enigma-go/v4/spec/mockapi",
		GoPackageName:       "mockapi",
		Version:             "no particular version",
		License:             "MIT",
		Description:         "mockapi is a package used to verify the spec generator itself",
	}
	actual := generateSpec("./mockapi", info)
	if *updateSpecGoldenfile {
		_ = ioutil.WriteFile("mockapi/mockspec.json", actual, 0644)
	} else {
		reference, _ := ioutil.ReadFile("mockapi/mockspec.json")
		assert.Equal(t, string(reference), string(actual), "Generated mock spec differs from the reference spec")
	}
}
