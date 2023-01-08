package httpsse

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const (
	guessChunkCount    = 8
	guessChunkOverhead = guessChunkCount * (6 + 2)
)

var dataFieldName = []byte("data: ")

const dataFieldSize = 6
const dataFieldOverhead = 8

func prepareEventIdentName(eventIdent, eventName string) []byte {
	return []byte("id: " + eventIdent + "\nevent: " + eventName + "\n")
}

func prepareEventName(eventName string) []byte {
	return []byte("event: " + eventName + "\n")
}

func prepareEventDataBuffer(prependBuf, data []byte) (result []byte) {
	prependBufSize := len(prependBuf)
	bufSize := prependBufSize + len(data) + guessChunkOverhead
	bufLimit := bufSize - dataFieldOverhead
	buf := make([]byte, bufSize)
	copy(buf, prependBuf)
	copy(buf[prependBufSize:], dataFieldName)
	targetIndex := prependBufSize + dataFieldSize
	needFieldPrefix := false
	for _, c := range data {
		if needFieldPrefix {
			copy(buf[targetIndex:], dataFieldName)
			targetIndex += dataFieldSize
			needFieldPrefix = false
		}
		if c == '\n' {
			buf[targetIndex] = '\n'
			targetIndex++
			needFieldPrefix = true
		} else if c < 0x20 {
			continue
		} else {
			buf[targetIndex] = c
			targetIndex++
		}
		if targetIndex >= bufLimit {
			bufSize = bufSize + guessChunkOverhead
			bufLimit = bufSize - dataFieldOverhead
			buf2 := make([]byte, bufSize)
			copy(buf2, buf)
			buf = buf2
		}
	}
	result = buf[:targetIndex]
	if needFieldPrefix {
		result = append(result, dataFieldName...)
	}
	result = append(result, []byte("\n\n")...)
	return
}

// EventStream support server-send event stream.
type EventStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewEventStream create new instance of event stream sender.
func NewEventStream(w http.ResponseWriter) (evtStream *EventStream) {
	flusher, _ := w.(http.Flusher)
	evtStream = &EventStream{
		flusher: flusher,
	}
	return
}

func (evtStream *EventStream) flush() {
	if evtStream.flusher == nil {
		return
	}
	evtStream.flusher.Flush()
}

// WriteHeadersWithStatusCode send event stream content headers and status line with
// given HTTP status code.
func (evtStream *EventStream) WriteHeadersWithStatusCode(statusCode int) {
	w := evtStream.w
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(statusCode)
}

// WriteHeaders send event stream content headers and status line.
func (evtStream *EventStream) WriteHeaders(statusCode int) {
	evtStream.WriteHeadersWithStatusCode(http.StatusOK)
}

// AdviseRetry emit retry message to client.
func (evtStream *EventStream) AdviseRetry(d time.Duration) (err error) {
	t := int64(d / time.Millisecond)
	if t < 1 {
		t = 1
	}
	msgBody := "retry: " + strconv.FormatInt(t, 10) + "\n\n"
	_, err = evtStream.w.Write([]byte(msgBody))
	evtStream.flush()
	return
}

// Heartbeat send a comment message with UNIX epoch in nano-seconds as heartbeat
// to keep connection alive.
func (evtStream *EventStream) Heartbeat() (err error) {
	msgBody := []byte(": " + strconv.FormatInt(time.Now().UnixNano(), 10) + "\n\n")
	_, err = evtStream.w.Write([]byte(msgBody))
	evtStream.flush()
	return
}

// SendString emit event message with given name and data string.
func (evtStream *EventStream) SendString(eventName, data string) (err error) {
	buf := prepareEventDataBuffer(prepareEventName(eventName), []byte(data))
	_, err = evtStream.w.Write(buf)
	evtStream.flush()
	return
}

// SendStringWithID emit event message with given ID, name and data string.
func (evtStream *EventStream) SendStringWithID(
	eventIdent, eventName, data string) (err error) {
	buf := prepareEventDataBuffer(
		prepareEventIdentName(eventIdent, eventName), []byte(data))
	_, err = evtStream.w.Write(buf)
	evtStream.flush()
	return
}

// SendJSON emit event message with given name and data.
// The data will encode into JSON before package into event message.
func (evtStream *EventStream) SendJSON(eventName string, data interface{}) (err error) {
	buf, err := json.Marshal(data)
	if nil != err {
		return
	}
	buf = prepareEventDataBuffer(prepareEventName(eventName), buf)
	_, err = evtStream.w.Write(buf)
	evtStream.flush()
	return
}

// SendJSONWithID emit event message with given ID, name and data.
// The data will encode into JSON before package into event message.
func (evtStream *EventStream) SendJSONWithID(eventIdent, eventName string, data interface{}) (err error) {
	buf, err := json.Marshal(data)
	if nil != err {
		return
	}
	buf = prepareEventDataBuffer(prepareEventIdentName(eventIdent, eventName), buf)
	_, err = evtStream.w.Write(buf)
	evtStream.flush()
	return
}
