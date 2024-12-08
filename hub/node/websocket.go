package node

import "github.com/gorilla/websocket"

type WebSocketConn struct {
	Conn *websocket.Conn
}

func (wsc *WebSocketConn) Read(p []byte) (n int, err error) {
	_, msg, err := wsc.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(p, msg)
	return len(msg), nil
}

func (wsc *WebSocketConn) Write(p []byte) (n int, err error) {
	err = wsc.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (wsc *WebSocketConn) Close() error {
	return wsc.Conn.Close()
}
