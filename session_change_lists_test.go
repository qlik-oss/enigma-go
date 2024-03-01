package enigma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionChangeLists(t *testing.T) {
	// Set up two channels - on for all change lists, and one for all
	s := newSessionChangeLists()
	allChangeListsChannel := s.ChangeListsChannel(false)
	pushedChangeListsChannel := s.ChangeListsChannel(true)

	// Emit two lists
	s.emitChangeLists([]int{1, 2}, []int{3, 4}, []int{9, 10}, true)
	s.emitChangeLists([]int{5, 6}, []int{7, 8}, []int{11, 12}, false)

	// Check that the lists appear as expected
	allData1 := <-allChangeListsChannel
	allData2 := <-allChangeListsChannel
	assert.Equal(t, 1, allData1.Changed[0])
	assert.Equal(t, 3, allData1.Closed[0])
	assert.Equal(t, 9, allData1.Suspended[0])
	assert.Equal(t, 5, allData2.Changed[0])
	assert.Equal(t, 7, allData2.Closed[0])
	assert.Equal(t, 11, allData2.Suspended[0])

	pushedData1 := <-pushedChangeListsChannel
	assert.Equal(t, 1, pushedData1.Changed[0])
	assert.Equal(t, 3, pushedData1.Closed[0])
	assert.Equal(t, 9, pushedData1.Suspended[0])

	// Check that the second non-pushed list doesn't appear
	var nothingInChangePushedChangeList bool
	select {
	case <-pushedChangeListsChannel:
		nothingInChangePushedChangeList = false
	default:
		nothingInChangePushedChangeList = true
	}
	assert.True(t, nothingInChangePushedChangeList)

	// Close listeners are check that they are actually deregistered
	s.CloseChangeListsChannel(pushedChangeListsChannel)
	s.CloseChangeListsChannel(allChangeListsChannel)
	assert.Equal(t, len(s.channels), 0)
}
