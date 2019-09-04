package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitDocShouldCatchExperimental(t *testing.T) {
	x := splitDoc("This is a nice function\nStability: Experimental\n")
	assert.False(t, x.Deprecated)
	assert.Equal(t, "Experimental", x.Stability)
	assert.Equal(t, "This is a nice function", x.Descr)
}

func TestSplitDocShouldSkipDepractionMidText(t *testing.T) {
	x := splitDoc("This is a nice function with Deprecated: in the description")
	assert.False(t, x.Deprecated)
	assert.Equal(t, "", x.Stability)
	assert.Equal(t, "This is a nice function with Deprecated: in the description", x.Descr)
}
func TestSplitDocOnlyDeprecation(t *testing.T) {
	x := splitDoc("Deprecated: This will be removed in a future version")
	assert.True(t, x.Deprecated)
	assert.Equal(t, "", x.Descr)
}
func TestSplitDocWithAllOfThem(t *testing.T) {
	x := splitDoc("This is a nice feature\nDeprecated: This will be removed in a future version\nStability: Experimental")
	assert.True(t, x.Deprecated)
	assert.Equal(t, "Experimental", x.Stability)
	assert.Equal(t, "This is a nice feature", x.Descr)
}
