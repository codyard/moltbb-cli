package api

// SignalR JSON protocol WebSocket client for TowerHub.
//
// Protocol reference:
//   - Messages are delimited by ASCII 0x1e (Unit Separator)
//   - Type 1: Invocation (client→server call, or server→client push)
//   - Type 3: Completion (server→client result for an invocation)
//   - Type 6: Ping / Pong
//   - Type 7: Close

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	signalrDelimiter = byte(0x1e)
	signalrHubPath   = "/api/tower"
)

// signalrMsg is the common envelope for all SignalR messages.
type signalrMsg struct {
	Type         int               `json:"type"`
	InvocationId string            `json:"invocationId,omitempty"`
	Target       string            `json:"target,omitempty"`
	Arguments    []json.RawMessage `json:"arguments"`
	Result       json.RawMessage   `json:"result,omitempty"`
	Error        string            `json:"error,omitempty"`
}

type invocationResult struct {
	data json.RawMessage
	err  error
}

// PushHandler is called when the server pushes an event.
type PushHandler func(args []json.RawMessage)

// SignalRConn is a single SignalR connection to TowerHub.
type SignalRConn struct {
	conn      *websocket.Conn
	handlers  map[string]PushHandler
	pending   map[string]chan invocationResult
	mu        sync.RWMutex
	counter   atomic.Int64
	done      chan struct{}
	closeOnce sync.Once
	closeErr  atomic.Value
}

// negotiate performs the SignalR negotiate handshake (POST /negotiate) and
// returns the connectionToken to use when opening the WebSocket.
func (c *Client) negotiate(ctx context.Context, token string) (string, error) {
	negotiateURL := c.baseURL + signalrHubPath + "/negotiate?negotiateVersion=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, negotiateURL, nil)
	if err != nil {
		return "", fmt.Errorf("build negotiate request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Length", "0")
	req.ContentLength = 0

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("negotiate: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("negotiate failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ConnectionToken string `json:"connectionToken"`
		ConnectionId    string `json:"connectionId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse negotiate response: %w", err)
	}
	if result.ConnectionToken != "" {
		return result.ConnectionToken, nil
	}
	return result.ConnectionId, nil
}

// ConnectToHub establishes a SignalR WebSocket connection to TowerHub.
// Performs the negotiate step first to obtain a connectionToken, which
// is required by ASP.NET Core SignalR to bind authentication context.
func (c *Client) ConnectToHub(ctx context.Context, token string) (*SignalRConn, error) {
	// Step 1: negotiate → get connectionToken
	connToken, err := c.negotiate(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("SignalR negotiate: %w", err)
	}

	// Step 2: open WebSocket with id=connectionToken and access_token for JWT auth
	wsBase := strings.Replace(c.baseURL, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)

	u, err := url.Parse(wsBase + signalrHubPath)
	if err != nil {
		return nil, fmt.Errorf("parse hub URL: %w", err)
	}
	q := u.Query()
	q.Set("id", connToken)
	q.Set("access_token", token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}
	wsHeaders := http.Header{}
	wsHeaders.Set("Authorization", "Bearer "+token)
	conn, _, err := dialer.DialContext(ctx, u.String(), wsHeaders)
	if err != nil {
		return nil, fmt.Errorf("connect to TowerHub: %w", err)
	}

	sc := &SignalRConn{
		conn:     conn,
		handlers: make(map[string]PushHandler),
		pending:  make(map[string]chan invocationResult),
		done:     make(chan struct{}),
	}

	// SignalR JSON handshake
	handshake := fmt.Sprintf(`{"protocol":"json","version":1}%c`, signalrDelimiter)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(handshake)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send SignalR handshake: %w", err)
	}

	// Read handshake response — must arrive within 10 s
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, raw, err := conn.ReadMessage()
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read SignalR handshake response: %w", err)
	}
	for _, part := range splitSignalR(raw) {
		if len(bytes.TrimSpace(part)) > 2 {
			var hsErr struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(part, &hsErr) == nil && hsErr.Error != "" {
				conn.Close()
				return nil, fmt.Errorf("SignalR handshake rejected: %s", hsErr.Error)
			}
		}
	}

	go sc.readLoop()
	go sc.pingLoop()
	return sc, nil
}

// On registers a handler for server-push events (type 1 with no invocationId).
func (sc *SignalRConn) On(target string, handler PushHandler) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.handlers[target] = handler
}

// Invoke calls a hub method and waits for its completion result.
// Returns the raw JSON result, or an error if the hub returned an error.
func (sc *SignalRConn) Invoke(ctx context.Context, target string, args ...any) (json.RawMessage, error) {
	id := fmt.Sprintf("%d", sc.counter.Add(1))

	rawArgs := make([]json.RawMessage, 0, len(args))
	for _, a := range args {
		b, err := json.Marshal(a)
		if err != nil {
			return nil, fmt.Errorf("marshal argument: %w", err)
		}
		rawArgs = append(rawArgs, json.RawMessage(b))
	}

	msg := signalrMsg{
		Type:         1,
		InvocationId: id,
		Target:       target,
		Arguments:    rawArgs,
	}

	ch := make(chan invocationResult, 1)
	sc.mu.Lock()
	sc.pending[id] = ch
	sc.mu.Unlock()

	defer func() {
		sc.mu.Lock()
		delete(sc.pending, id)
		sc.mu.Unlock()
	}()

	if err := sc.writeMsg(msg); err != nil {
		return nil, err
	}

	select {
	case res := <-ch:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-sc.done:
		if err, ok := sc.closeErr.Load().(error); ok && err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("connection closed")
	}
}

// InvokeVoid calls a hub method that returns void (no result expected).
func (sc *SignalRConn) InvokeVoid(ctx context.Context, target string, args ...any) error {
	_, err := sc.Invoke(ctx, target, args...)
	return err
}

// Close gracefully closes the WebSocket connection.
func (sc *SignalRConn) Close() {
	sc.closeOnce.Do(func() {
		_ = sc.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		sc.conn.Close()
	})
}

// Done returns a channel closed when the connection is lost or closed.
func (sc *SignalRConn) Done() <-chan struct{} {
	return sc.done
}

// ── internal ──────────────────────────────────────────────────────────────────

func (sc *SignalRConn) writeMsg(msg signalrMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal SignalR message: %w", err)
	}
	data = append(data, signalrDelimiter)
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteMessage(websocket.TextMessage, data)
}

func (sc *SignalRConn) readLoop() {
	defer close(sc.done)
	for {
		_, data, err := sc.conn.ReadMessage()
		if err != nil {
			sc.closeErr.Store(fmt.Errorf("signalr read failed: %w", err))
			return
		}
		for _, part := range splitSignalR(data) {
			part = bytes.TrimSpace(part)
			if len(part) == 0 {
				continue
			}
			var msg signalrMsg
			if err := json.Unmarshal(part, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case 1: // Invocation / push
				if msg.InvocationId != "" {
					// Client-side invocation echo — ignore
					break
				}
				sc.mu.RLock()
				h, ok := sc.handlers[msg.Target]
				sc.mu.RUnlock()
				if ok {
					h(msg.Arguments)
				}
			case 3: // Completion
				sc.mu.RLock()
				ch, ok := sc.pending[msg.InvocationId]
				sc.mu.RUnlock()
				if ok {
					if msg.Error != "" {
						ch <- invocationResult{err: fmt.Errorf("%s", msg.Error)}
					} else {
						ch <- invocationResult{data: msg.Result}
					}
				}
			case 7: // Close
				err := fmt.Errorf("signalr close frame: %s", string(part))
				sc.closeErr.Store(err)
				sc.conn.Close()
				return
			}
		}
	}
}

func (sc *SignalRConn) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ping := fmt.Sprintf(`{"type":6}%c`, signalrDelimiter)
			sc.mu.Lock()
			_ = sc.conn.WriteMessage(websocket.TextMessage, []byte(ping))
			sc.mu.Unlock()
		case <-sc.done:
			return
		}
	}
}

func splitSignalR(data []byte) [][]byte {
	return bytes.Split(data, []byte{signalrDelimiter})
}
