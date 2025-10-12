package http

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/log"
	"github.com/gorilla/websocket"
)

type hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	in      chan []byte
	reg     chan *client
	unreg   chan *client
}

type client struct {
	ws   *websocket.Conn
	send chan []byte
}

var (
	wsHub     *hub
	wsOnce    sync.Once
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	logWriter *broadcastWriter
)

func getHub() *hub {
	wsOnce.Do(func() {
		wsHub = &hub{
			clients: map[*client]struct{}{},
			in:      make(chan []byte, 1024),
			reg:     make(chan *client),
			unreg:   make(chan *client),
		}
		go wsHub.run()
	})
	return wsHub
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.reg:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unreg:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()
		case msg := <-h.in:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

type broadcastWriter struct {
	h   *hub
	mu  sync.Mutex
	buf []byte
}

func (w *broadcastWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.buf = append(w.buf, p...)
	start := 0
	for {
		i := bytes.IndexByte(w.buf[start:], '\n')
		if i < 0 {
			break
		}
		end := start + i
		line := make([]byte, end-start)
		copy(line, w.buf[start:end])
		w.h.in <- line
		start = end + 1
	}
	if start > 0 {
		w.buf = append([]byte{}, w.buf[start:]...)
	}
	w.mu.Unlock()
	return len(p), nil
}

func LogWriter() io.Writer {
	getHub()
	if logWriter == nil {
		logWriter = &broadcastWriter{h: wsHub}
	}
	return logWriter
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	h := getHub()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &client{ws: conn, send: make(chan []byte, 256)}
	log.Tracef("WebSocket client connected: %s", r.RemoteAddr)
	h.reg <- c
	go writePump(c)
	readPump(c, h)
}

func writePump(c *client) {
	defer c.ws.Close()
	for msg := range c.send {
		c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
	_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{})
}

func readPump(c *client, h *hub) {
	defer func() { h.unreg <- c }()
	for {
		if _, _, err := c.ws.ReadMessage(); err != nil {
			return
		}
	}
}
