package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

func newWebSocketHandler() *webSocketHandler {
	return &webSocketHandler{
		upgrader: websocket.Upgrader{},
		clients:  make(map[*websocket.Conn]bool),
	}
}

type chatMessage struct {
	Message string `json:"message"`
}

func (wsh webSocketHandler) WSHandler(w http.ResponseWriter, r *http.Request) {
	c, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	wsh.clients[c] = true
	defer c.Close()
	defer delete(wsh.clients, c)

	// keep reading messages from the client until the connection is closed
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		if message == nil {
			continue
		}
		log.Printf("recv: %s", message)

		var chatMessage chatMessage
		err = json.Unmarshal(message, &chatMessage)
		if err != nil {
			log.Println(err)
			continue
		}

		// broadcast the message to all clients
		for client := range wsh.clients {
			err = client.WriteMessage(mt, []byte(fmt.Sprintf(`<ul id="chat" hx-swap-oob="beforeend">%s</ul>`, chatMessage.Message)))
			if err != nil {
				log.Println(err)
				client.Close()
				delete(wsh.clients, client)
			}
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/static/chat.html")
	}
}

func init() {
	log.Println("Listening on port 8000...")
}

func main() {
	wsh := newWebSocketHandler()

	http.HandleFunc("/chat", wsh.WSHandler)
	http.HandleFunc("/", homeHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
