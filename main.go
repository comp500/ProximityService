package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins
		return true
	},
}

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	box := packr.New("WebFiles", "./web/dist")
	manager := clientManager{
		make(map[*client]bool),
		make(chan clientMessage),
		make(chan *client),
		make(chan *client),
	}

	http.Handle("/", http.FileServer(box))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		c := client{make(chan clientMessage)}
		manager.register <- &c

		for {
			msg, ok := <-c.send
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err = conn.WriteJSON(msg)
			if err != nil {
				manager.unregister <- &c
				fmt.Println(err)
				return
			}
		}
	})

	dataChannel := make(chan []byte)

	go manager.run()
	go handleData(dataChannel, &manager)
	go startBluetooth(dataChannel)

	fmt.Printf("Starting server on port %d\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}

func handleData(dataChannel chan []byte, manager *clientManager) {
	rcvPart1 := false
	var part1Analog int
	binaryValue := false
	for {
		data := <-dataChannel
		for _, b := range data {
			if (b & 0x80) == 0x80 {
				binaryValue = (b & 0x40) == 0x40

				rcvPart1 = true
				part1Analog = int(b&0x07) << 7
			} else {
				if rcvPart1 {
					rcvPart1 = false
					analogValue := int(b) | part1Analog

					manager.broadcast <- clientMessage{binaryValue, analogValue}
				}
			}
		}
	}
}

type clientMessage struct {
	Bin    bool
	Analog int
}

type client struct {
	send chan clientMessage
}

type clientManager struct {
	clients    map[*client]bool
	broadcast  chan clientMessage
	register   chan *client
	unregister chan *client
}

func (m *clientManager) run() {
	for {
		select {
		case c := <-m.register:
			m.clients[c] = true
		case c := <-m.unregister:
			if _, ok := m.clients[c]; ok {
				delete(m.clients, c)
				close(c.send)
			}
		case message := <-m.broadcast:
			for c := range m.clients {
				c.send <- message
			}
		}
	}
}
