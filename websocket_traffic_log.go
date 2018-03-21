package enigma

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type (
	trafficLogRow struct {
		Sent     json.RawMessage `json:"Sent,omitempty"`
		Received json.RawMessage `json:"Received,omitempty"`
	}

	fileTrafficLog struct {
		FileName string
		Messages []trafficLogRow
		mutex    sync.Mutex
	}
)

// Opened implements the TrafficLogger interface
func (t *fileTrafficLog) Opened() {

}

// Sent implements the TrafficLogger interface
func (t *fileTrafficLog) Sent(message []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Messages = append(t.Messages, trafficLogRow{Sent: message})
}

// Received implements the TrafficLogger interface
func (t *fileTrafficLog) Received(message []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Messages = append(t.Messages, trafficLogRow{Received: message})
}

// Closed implements the TrafficLogger interface
func (t *fileTrafficLog) Closed() {
	bytes, _ := json.MarshalIndent(t.Messages, "", "\t")
	ioutil.WriteFile(t.FileName, bytes, 0644)
}

func newFileTrafficLogger(filename string) *fileTrafficLog {
	return &fileTrafficLog{FileName: filename, Messages: make([]trafficLogRow, 0, 1000)}
}

func readTrafficLog(fileName string) []trafficLogRow {
	file, _ := ioutil.ReadFile(fileName)
	result := []trafficLogRow{}
	_ = json.Unmarshal(file, &result)
	return result
}
