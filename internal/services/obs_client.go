package services

import (
	"crypto/sha256"
	"encoding/base64"
	//"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// OBSClient - klient WebSocket dla OBS Studio
type OBSClient struct {
	conn              *websocket.Conn
	url               string
	password          string
	connected         bool
	mu                sync.RWMutex
	reconnectTimer    chan bool
	messageHandlers   map[string]func(map[string]interface{})
	requestID         int
	pendingRequests   map[string]chan map[string]interface{}
	pendingRequestsMu sync.RWMutex
}

// OBSMessage - struktura wiadomości OBS WebSocket
type OBSMessage struct {
	Op int                    `json:"op"`
	D  map[string]interface{} `json:"d"`
}

// NewOBSClient - tworzy nowego klienta OBS
func NewOBSClient(url, password string) *OBSClient {
	return &OBSClient{
		url:             url,
		password:        password,
		reconnectTimer:  make(chan bool),
		messageHandlers: make(map[string]func(map[string]interface{})),
		pendingRequests: make(map[string]chan map[string]interface{}),
	}
}

// Connect - łączy się z serwerem OBS WebSocket
func (c *OBSClient) Connect() {
	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	log.Println("OBS: Próba połączenia z serwerem OBS WebSocket...")

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		log.Printf("OBS: Błąd połączenia: %v. Ponowna próba za 5s...", err)
		c.scheduleReconnect()
		return
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	log.Println("OBS: Połączono z serwerem WebSocket")

	go c.receiveMessages()
}

// scheduleReconnect - planuje ponowne połączenie po 5 sekundach
func (c *OBSClient) scheduleReconnect() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	go func() {
		log.Println("OBS: Zaplanowano ponowne połączenie za 5 sekund...")
		select {
		case <-c.reconnectTimer:
			log.Println("OBS: Anulowano ponowne połączenie")
			return
		case <-time.After(5 * time.Second):
			log.Println("OBS: Próba ponownego połączenia...")
			c.Connect()
		}
	}()
}

// receiveMessages - odbiera wiadomości z WebSocket
func (c *OBSClient) receiveMessages() {
	defer c.handleDisconnect()

	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		var msg OBSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("OBS: Błąd odczytu wiadomości: %v", err)
			return
		}

		c.handleMessage(msg)
	}
}

// handleMessage - obsługuje odebraną wiadomość
func (c *OBSClient) handleMessage(msg OBSMessage) {
	switch msg.Op {
	case 0:
		c.handleHello(msg.D)
	case 2:
		log.Println("OBS: Zidentyfikowano pomyślnie")
		c.onConnected()
	case 5:
		c.handleEvent(msg.D)
	case 7:
		c.handleRequestResponse(msg.D)
	case 9:
		c.handleRequestResponse(msg.D)
	}
}

// handleHello - obsługuje wiadomość Hello (autoryzacja)
func (c *OBSClient) handleHello(data map[string]interface{}) {
	log.Println("OBS: Otrzymano Hello, rozpoczynam autoryzację...")

	authData, hasAuth := data["authentication"].(map[string]interface{})

	if hasAuth && c.password != "" {
		challenge := authData["challenge"].(string)
		salt := authData["salt"].(string)

		secret := c.generateAuthString(c.password, salt, challenge)

		identifyMsg := OBSMessage{
			Op: 1,
			D: map[string]interface{}{
				"rpcVersion":         1,
				"authentication":     secret,
				"eventSubscriptions": 33,
			},
		}
		c.sendMessage(identifyMsg)
	} else {
		identifyMsg := OBSMessage{
			Op: 1,
			D: map[string]interface{}{
				"rpcVersion":         1,
				"eventSubscriptions": 33,
			},
		}
		c.sendMessage(identifyMsg)
	}
}

// generateAuthString - generuje string autoryzacji
func (c *OBSClient) generateAuthString(password, salt, challenge string) string {
	h1 := sha256.New()
	h1.Write([]byte(password + salt))
	secret := base64.StdEncoding.EncodeToString(h1.Sum(nil))

	h2 := sha256.New()
	h2.Write([]byte(secret + challenge))
	return base64.StdEncoding.EncodeToString(h2.Sum(nil))
}

// handleEvent - obsługuje event z OBS
func (c *OBSClient) handleEvent(data map[string]interface{}) {
	eventType, ok := data["eventType"].(string)
	if !ok {
		return
	}

	eventData, _ := data["eventData"].(map[string]interface{})
	log.Printf("OBS Event: %s", eventType)

	if handler, exists := c.messageHandlers[eventType]; exists {
		handler(eventData)
	}
}

// handleRequestResponse - obsługuje odpowiedź na request
func (c *OBSClient) handleRequestResponse(data map[string]interface{}) {
	requestID, ok := data["requestId"].(string)
	if !ok {
		return
	}

	c.pendingRequestsMu.Lock()
	respChan, exists := c.pendingRequests[requestID]
	if exists {
		delete(c.pendingRequests, requestID)
	}
	c.pendingRequestsMu.Unlock()

	if exists {
		respChan <- data
		close(respChan)
	}
}

// sendMessage - wysyła wiadomość do OBS
func (c *OBSClient) sendMessage(msg OBSMessage) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil
	}

	return conn.WriteJSON(msg)
}

// SendRequest - wysyła request do OBS i czeka na odpowiedź
func (c *OBSClient) SendRequest(requestType string, requestData map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, nil
	}
	c.mu.RUnlock()

	c.requestID++
	requestID := string(rune(c.requestID))

	respChan := make(chan map[string]interface{})
	c.pendingRequestsMu.Lock()
	c.pendingRequests[requestID] = respChan
	c.pendingRequestsMu.Unlock()

	msg := OBSMessage{
		Op: 6,
		D: map[string]interface{}{
			"requestType": requestType,
			"requestId":   requestID,
			"requestData": requestData,
		},
	}

	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}

	response := <-respChan
	return response, nil
}

// OnEvent - rejestruje handler dla eventu
func (c *OBSClient) OnEvent(eventType string, handler func(map[string]interface{})) {
	c.messageHandlers[eventType] = handler
}

// onConnected - wywoływane po pomyślnym połączeniu
func (c *OBSClient) onConnected() {
	log.Println("OBS: Połączenie ustanowione, gotowy do pracy")
}

// handleDisconnect - obsługuje rozłączenie
func (c *OBSClient) handleDisconnect() {
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	wasConnected := c.connected
	c.connected = false
	c.mu.Unlock()

	if wasConnected {
		log.Println("OBS: Rozłączono. Próba ponownego połączenia za 5s...")
		c.scheduleReconnect()
	}
}

// IsConnected - sprawdza czy jest połączenie
func (c *OBSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Close - zamyka połączenie
func (c *OBSClient) Close() {
	c.reconnectTimer <- true
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
}

// StartRecording - rozpoczyna nagrywanie w OBS
func (c *OBSClient) StartRecording() error {
	_, err := c.SendRequest("StartRecord", nil)
	return err
}

// StopRecording - zatrzymuje nagrywanie w OBS
func (c *OBSClient) StopRecording() error {
	_, err := c.SendRequest("StopRecord", nil)
	return err
}

// GetRecordingStatus - pobiera status nagrywania
func (c *OBSClient) GetRecordingStatus() (bool, error) {
	response, err := c.SendRequest("GetRecordStatus", nil)
	if err != nil {
		return false, err
	}

	if responseData, ok := response["responseData"].(map[string]interface{}); ok {
		if outputActive, ok := responseData["outputActive"].(bool); ok {
			return outputActive, nil
		}
	}

	return false, nil
}

// SetCurrentScene - ustawia aktywną scenę
func (c *OBSClient) SetCurrentScene(sceneName string) error {
	_, err := c.SendRequest("SetCurrentProgramScene", map[string]interface{}{
		"sceneName": sceneName,
	})
	return err
}

// GetSceneList - pobiera listę scen
func (c *OBSClient) GetSceneList() ([]string, error) {
	response, err := c.SendRequest("GetSceneList", nil)
	if err != nil {
		return nil, err
	}

	scenes := []string{}
	if responseData, ok := response["responseData"].(map[string]interface{}); ok {
		if sceneList, ok := responseData["scenes"].([]interface{}); ok {
			for _, scene := range sceneList {
				if sceneMap, ok := scene.(map[string]interface{}); ok {
					if sceneName, ok := sceneMap["sceneName"].(string); ok {
						scenes = append(scenes, sceneName)
					}
				}
			}
		}
	}

	return scenes, nil
}