package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Listening on port 8000")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "../../web/static/chat.html")
	case http.MethodPost:
		fmt.Printf("got message: { %s } at { %s }\n", r.FormValue("message"), time.Now().Format(time.RFC1123))
		http.ServeFile(w, r, "../../web/static/chat.html")
	}
}
