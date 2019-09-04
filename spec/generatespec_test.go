package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	overwrite := false
	actual := generateSpec("./mockapi", "mockapi")
	if overwrite {
		_ = ioutil.WriteFile("mockapi/mockspec.json", actual, 0644)
	} else {
		reference, _ := ioutil.ReadFile("mockapi/mockspec.json")
		assert.Equal(t, reference, actual, "Generated mock spec differs from the reference spec")
	}
}
