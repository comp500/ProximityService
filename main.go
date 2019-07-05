package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	box := packr.NewBox("./web")

	http.Handle("/", http.FileServer(box))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		_ = conn
	})

	fmt.Printf("Starting server on port %d\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
