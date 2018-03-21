package enigma

import (
	"context"
	"encoding/json"
	"sync"
)

type (
	sessionMessageChannelEntry struct {
		topics  []string
		channel chan SessionMessage
	}

	// SessionMessage is a notification regarding the session coming from Qlik Associative Engine.
	// The content is stored as a raw json structure.
	SessionMessage struct {
		Topic   string
		Content json.RawMessage
	}

	sessionMessages struct {
		history  []SessionMessage
		mutex    sync.Mutex
		channels map[*sessionMessageChannelEntry]bool
	}
)

func (entry *sessionMessageChannelEntry) emitSessionEvent(sessionEvent SessionMessage) {
	if len(entry.topics) > 0 && entry.topics[0] != "*" {
		for _, topic := range entry.topics {
			if sessionEvent.Topic == topic {
				entry.channel <- sessionEvent
				break
			}
		}
	} else {
		// Send all events if no limiting is supplied
		entry.channel <- sessionEvent
	}
}

func (e *sessionMessages) emitSessionMessage(topic string, value json.RawMessage) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	event := SessionMessage{Topic: topic, Content: value}
	e.history = append(e.history, event)
	for channelEntry := range e.channels {
		channelEntry.emitSessionEvent(event)
	}
}

// SessionMessageChannel returns a channel that receives notifications from Qlik Associative Engine. To only receive
// certain events a list of topics can be supplied. If no topics are supplied all events are received.
func (e *sessionMessages) SessionMessageChannel(topics ...string) chan SessionMessage {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	channelEntry := &sessionMessageChannelEntry{topics: topics, channel: make(chan SessionMessage, 16+len(e.history))}
	e.channels[channelEntry] = true
	for _, oldEvent := range e.history {
		channelEntry.emitSessionEvent(oldEvent)
	}
	return channelEntry.channel
}

// CloseSessionMessageChannel closes and unregisters the supplied event channel from the session.
func (e *sessionMessages) CloseSessionMessageChannel(channel chan SessionMessage) {
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

func (e *sessionMessages) closeAllSessionEventChannels() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for channelEntry := range e.channels {
		close(channelEntry.channel)
	}
	e.channels = make(map[*sessionMessageChannelEntry]bool)
}

func newSessionEvents() *sessionMessages {
	return &sessionMessages{channels: make(map[*sessionMessageChannelEntry]bool), mutex: sync.Mutex{}, history: make([]SessionMessage, 0)}
}

// SessionState returns either SESSION_CREATED or SESSION_ATTACHED to describe the status of the current websocket session
func (e *sessionMessages) SessionState(ctx context.Context) (string, error) {
	channel := e.SessionMessageChannel("OnConnected")
	defer e.CloseSessionMessageChannel(channel)
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case message := <-channel:
		connectedInfo := &onConnectedEvent{}
		err := json.Unmarshal(message.Content, connectedInfo)
		if err != nil {
			return "", err
		}
		return connectedInfo.SessionState, nil
	}
}
