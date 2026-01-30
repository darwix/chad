package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

// Message types
const (
	TypeChat     = "chat"
	TypePresence = "presence"
	TypeTyping   = "typing"
)

type Message struct {
	ID        int64     `json:"id,omitempty"`
	Type      string    `json:"type"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type PresenceUpdate struct {
	Type  string   `json:"type"`
	Users []string `json:"users"`
}

type TypingUpdate struct {
	Type   string `json:"type"`
	Sender string `json:"sender"`
	IsTyping bool `json:"is_typing"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	name string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	db         *sql.DB
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.broadcastPresence()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.broadcastPresence()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) broadcastPresence() {
	h.mu.Lock()
	users := []string{}
	for client := range h.clients {
		if client.name != "" {
			users = append(users, client.name)
		}
	}
	h.mu.Unlock()

	presence := PresenceUpdate{
		Type:  TypePresence,
		Users: users,
	}
	data, _ := json.Marshal(presence)
	// Use a goroutine to avoid deadlock if called from Hub.Run
	go func() {
		h.broadcast <- data
	}()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == TypeChat {
				msg.Timestamp = time.Now()
				// Save to DB
		res, err := c.hub.db.Exec("INSERT INTO messages (sender, content, timestamp) VALUES (?, ?, ?)",
					msg.Sender, msg.Content, msg.Timestamp)
		if err == nil {
			msg.ID, _ = res.LastInsertId()
		}
				if err != nil {
					log.Printf("db error: %v", err)
				}

				finalMsg, _ := json.Marshal(msg)
				c.hub.broadcast <- finalMsg
			} else if msg.Type == TypeTyping {
				c.hub.broadcast <- message
			} else if msg.Type == TypePresence {
				// Initial name setting
				c.hub.mu.Lock()
				c.name = msg.Sender
				c.hub.mu.Unlock()
				c.hub.broadcastPresence()
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite", "./chat.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender TEXT,
		content TEXT,
		timestamp DATETIME
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
	return db
}

func getHistory(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
	rows, err := db.Query("SELECT id, sender, content, timestamp FROM messages ORDER BY timestamp DESC LIMIT 50")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		m.Type = TypeChat
		err := rows.Scan(&m.ID, &m.Sender, &m.Content, &m.Timestamp)
		if err != nil {
			log.Println(err)
			continue
		}
		messages = append(messages, m)
	}

    // Reverse messages to get chronological order
    for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
        messages[i], messages[j] = messages[j], messages[i]
    }

	json.NewEncoder(w).Encode(messages)
}

func main() {
	db := initDB()
	defer db.Close()

	hub := NewHub(db)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		getHistory(db, w, r)
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
