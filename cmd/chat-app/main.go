package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
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
	defer c.Close()

	mt, message, err := c.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	if message == nil {
		return
	}
	log.Printf("recv: %s\n at: %s", message, time.Now().Format(time.RFC1123))

	var chatMessage chatMessage
	err = json.Unmarshal([]byte(message), &chatMessage)
	if err != nil {
		log.Println(err)
		return
	}

	err = c.WriteMessage(mt, []byte(fmt.Sprintf(`<ul id="chat" hx-swap-oob="beforeend">%s</ul>`, chatMessage.Message)))
	if err != nil {
		log.Println(err)
		return
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
	wsh := webSocketHandler{upgrader: websocket.Upgrader{}}

	http.HandleFunc("/chat", wsh.WSHandler)
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8000", nil)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
