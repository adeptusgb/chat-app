package main

import (
	"bytes"
	"encoding/json"
	"html/template"
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

type chatMessageTemplate struct {
	Username string
	Message  string
}

var messageTemplate *template.Template

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
		if len(message) == 0 {
			continue
		}
		log.Printf("recv: %s", message)

		var chatMessage chatMessage
		err = json.Unmarshal(message, &chatMessage)
		if err != nil {
			log.Println(err)
			continue
		}

		// render the message using the template
		chatMessageTemplate := chatMessageTemplate{
			Username: r.RemoteAddr,
			Message:  chatMessage.Message,
		}

		var tpl bytes.Buffer
		if err := messageTemplate.Execute(&tpl, chatMessageTemplate); err != nil {
			log.Println(err)
			continue
		}
		formattedMessage := tpl.String()

		// broadcast the formatted message to all clients
		for client := range wsh.clients {
			err = client.WriteMessage(mt, []byte(formattedMessage))

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

	var err error
	messageTemplate, err = template.ParseFiles("web/template/chat-message.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
}

func main() {
	wsh := newWebSocketHandler()

	http.HandleFunc("/chat", wsh.WSHandler)
	http.HandleFunc("/", homeHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
