package main

import (
	_ "embed"
	"log"
	"net/http"
	"strconv"
	"time"

	httpsse "github.com/yinyin/go-util-http-sse"
)

//go:embed index.html
var mainContent []byte

type eventData struct {
	Cycle int
	Tick  time.Time
	Ident string
}

type sampleHandler struct {
}

func (h *sampleHandler) hndSentEvents(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	evtStream := httpsse.NewEventStream(w)
	evtStream.WriteHeaders()
	evtStream.AdviseRetry(time.Second * 5)
	lastEventID := httpsse.GetLastEventID(req)
	log.Printf("INFO: last event ID: [%s]", lastEventID)
	for cycle := 0; cycle < 10; cycle++ {
		select {
		case t := <-ticker.C:
			switch part := cycle % 5; part {
			case 0:
				if err := evtStream.SendJSON("part0", &eventData{
					Cycle: cycle,
					Tick:  t,
				}); nil != err {
					log.Printf("ERROR: cannot sent part-0: %v", err)
					return
				}
			case 1:
				identCode := "p1-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.FormatInt(int64(cycle), 10)
				if err := evtStream.SendJSONWithID(identCode, "part1", &eventData{
					Cycle: cycle,
					Tick:  t,
					Ident: identCode,
				}); nil != err {
					log.Printf("ERROR: cannot sent part-1: %v", err)
					return
				}
			case 2:
				if err := evtStream.SendString(
					"part2",
					"cycle="+strconv.FormatInt(int64(cycle), 10)+", t="+t.Format(time.RFC3339)); nil != err {
					log.Printf("ERROR: cannot sent part-2: %v", err)
					return
				}
			case 3:
				identCode := "p3-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.FormatInt(int64(cycle), 10)
				if err := evtStream.SendStringWithID(
					identCode,
					"part3",
					"cycle="+strconv.FormatInt(int64(cycle), 10)+", t="+t.Format(time.RFC3339)+", ident="+identCode); nil != err {
					log.Printf("ERROR: cannot sent part-3: %v", err)
					return
				}
			default:
				evtStream.Heartbeat()
			}
		case <-ctx.Done():
			log.Printf("WARN: link lost: %v", ctx.Err())
			return
		}
	}
}

func (h *sampleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet, http.MethodPost:
		break
	default:
		http.Error(w, "not allow: "+req.Method, http.StatusMethodNotAllowed)
		return
	}
	log.Printf("DEBUG: method=%s; path=%s", req.Method, req.URL.Path)
	switch req.URL.Path {
	case "/endpoint/sse":
		h.hndSentEvents(w, req)
		return
	case "/":
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(mainContent)), 10))
		w.WriteHeader(http.StatusOK)
		w.Write(mainContent)
		return
	}
	http.NotFound(w, req)
}
