package enigma

import (
	"sync"
)

type (
	sessionChangeLists struct {
		mutex    sync.Mutex
		channels map[*sessionChangeListEntry]bool
	}

	sessionChangeListEntry struct {
		pushedOnly bool
		channel    chan ChangeLists
	}
)

func (e *sessionChangeLists) emitChangeLists(changed []int, closed []int, pushed bool) {
	if len(changed) > 0 || len(closed) > 0 {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		for channelEntry := range e.channels {
			if pushed || !channelEntry.pushedOnly {
				channelEntry.channel <- ChangeLists{Changed: changed, Closed: closed}
			}
		}
	}
}

// ChangeListsChannel returns a channel that receives change and close notifications from Qlik Associative Engine. if the pushedOnly argument is set to true
// only pushed change lists are put into the channel - not changes that are returned as a response to an API call.
func (e *sessionChangeLists) ChangeListsChannel(pushedOnly bool) chan ChangeLists {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	channelEntry := &sessionChangeListEntry{channel: make(chan ChangeLists, 16), pushedOnly: pushedOnly}
	e.channels[channelEntry] = true
	return channelEntry.channel
}

// CloseChangeListsChannel closes and unregisters the supplied event channel from the session.
func (e *sessionChangeLists) CloseChangeListsChannel(channel chan ChangeLists) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for channelEntry := range e.channels {
		if channelEntry.channel == channel {
			close(channelEntry.channel)
			delete(e.channels, channelEntry)
			break
		}
	}
}

func (e *sessionChangeLists) closeAllChangeListChannels() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for channelEntry := range e.channels {
		close(channelEntry.channel)
	}
	e.channels = make(map[*sessionChangeListEntry]bool)
}

func newSessionChangeLists() *sessionChangeLists {
	return &sessionChangeLists{channels: make(map[*sessionChangeListEntry]bool), mutex: sync.Mutex{}}
}
