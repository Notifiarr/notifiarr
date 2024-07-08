package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nxadm/tail"
)

const startFileBytes = 1500

//nolint:gochecknoglobals
var upgrader = websocket.Upgrader{
	ReadBufferSize:  mnd.Kilobyte,
	WriteBufferSize: mnd.Kilobyte,
}

// Websockets cannot use our apachelog package, so we have to produce our own log line for those reqs.
func (c *Client) socketLog(code int, r *http.Request) {
	_, _ = c.Logger.HTTPLog.Writer().Write([]byte(fmt.Sprintf(`%s - - [%s] "%s %s %s" %d 0 "%s" "%s"`+"\n",
		r.Header.Get("X-Forwarded-For"), time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		r.Method, r.RequestURI, r.Proto, code, r.Header.Get("Referer"), r.Header.Get("User-Agent"))))
}

func (c *Client) handleWebSockets(response http.ResponseWriter, request *http.Request) {
	defer c.CapturePanic()

	var fileInfos *logs.LogFileInfos

	switch src := mux.Vars(request)["source"]; src {
	case fileSourceLogs:
		fileInfos = c.Logger.GetAllLogFilePaths()
	default:
		http.Error(response, "invalid source: "+src, http.StatusBadRequest)
		c.socketLog(http.StatusBadRequest, request)

		return
	}

	fileID := mux.Vars(request)["fileId"]

	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != fileID {
			continue
		}

		offset := int64(startFileBytes)
		if fileInfo.Size < startFileBytes {
			offset = fileInfo.Size
		}

		fileTail, err := tail.TailFile(fileInfo.Path,
			tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: -offset, Whence: io.SeekEnd}})
		if err != nil {
			http.Error(response, "tail error: "+err.Error(), http.StatusBadRequest)
			c.socketLog(http.StatusInternalServerError, request)

			return
		}

		socket, err := upgrader.Upgrade(response, request, nil)
		if err != nil {
			c.Errorf("[gui requested] Creating Websocket: %v", err)
			c.socketLog(http.StatusInternalServerError, request)

			return
		}

		go c.webSocketWriter(socket, fileTail)
		c.socketLog(http.StatusOK, request)
		c.webSocketReader(socket)

		return
	}

	http.Error(response, "file for ID not found: "+fileID, http.StatusBadRequest)
	c.socketLog(http.StatusBadRequest, request)
}

func (c *Client) webSocketWriter(socket *websocket.Conn, fileTail *tail.Tail) {
	var (
		lastError  = ""
		pingTicker = time.NewTicker(29 * time.Second) //nolint:mnd
		writeWait  = 10 * time.Second
	)

	defer func() {
		c.CapturePanic()
		fileTail.Stop() //nolint:errcheck
		pingTicker.Stop()
		socket.Close()
	}()

	for {
		select {
		case line := <-fileTail.Lines:
			if line == nil {
				c.Debugf("file lines return empty, dropping websocket (did the file rotate?)")
				return
			}

			if line.Num == 1 {
				continue // first line is never complete.
			}

			text := line.Text

			if line.Err != nil {
				if lineErr := line.Err.Error(); lineErr != lastError {
					lastError = lineErr
					text = lineErr
				} else {
					c.Debugf("closing websocket due to errors: %v", lineErr)
					return // two errors.
				}
			} else {
				lastError = ""
			}

			_ = socket.SetWriteDeadline(time.Now().Add(writeWait))

			if err := socket.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
				c.Debugf("websocket closed, write error: %v", err)
				return // dead sock
			}
		case <-pingTicker.C:
			_ = socket.SetWriteDeadline(time.Now().Add(writeWait))

			if err := socket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.Debugf("websocket closed, ping error: %v", err)
				return
			}
		}
	}
}

func (c *Client) webSocketReader(socket *websocket.Conn) {
	defer func() {
		c.CapturePanic()
		socket.Close()
	}()

	socket.SetReadLimit(1) // we don't read anything from here.
	_ = socket.SetReadDeadline(time.Now().Add(1 * time.Minute))
	socket.SetPongHandler(func(string) error {
		_ = socket.SetReadDeadline(time.Now().Add(1 * time.Minute))
		return nil
	})

	for {
		if _, _, err := socket.ReadMessage(); err != nil {
			break
		}
	}
}
