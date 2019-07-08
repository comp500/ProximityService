package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

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
	list := clientList{}

	http.Handle("/", http.FileServer(box))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		messageChannel := make(chan clientMessage)
		list.Lock()
		list.clients = append(list.clients, messageChannel)
		list.Unlock()

		for {
			msg := <-messageChannel
			err = conn.WriteJSON(msg)
			if err != nil {
				list.Lock()
				index := -1
				for i, v := range list.clients {
					if v == messageChannel {
						index = i
						break
					}
				}
				if index > -1 {
					list.clients[index] = list.clients[len(list.clients)-1]
					list.clients = list.clients[:len(list.clients)-1]
				}
				list.Unlock()
				return
			}
		}
	})

	dataChannel := make(chan []byte)
	done := make(chan bool)

	go handleData(dataChannel, done, &list)
	go startBluetooth(dataChannel, done)

	fmt.Printf("Starting server on port %d\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

	done <- true
}

func handleData(dataChannel chan []byte, done chan bool, list *clientList) {
	rcvPart1 := false
	var part1Analog int
	binaryValue := false
	for {
		select {
		case <-done:
			return
		case data := <-dataChannel:
			for _, b := range data {
				if (b & 0x80) == 0x80 {
					binaryValue = (b & 0x40) == 0x40

					rcvPart1 = true
					part1Analog = int(b&0x07) << 7
				} else {
					if rcvPart1 {
						rcvPart1 = false
						analogValue := int(b) | part1Analog

						msg := clientMessage{binaryValue, analogValue}
						list.Lock()
						for _, c := range list.clients {
							c <- msg
						}
						list.Unlock()
					}
				}
			}
		}
	}
}

type clientMessage struct {
	Bin    bool
	Analog int
}

type clientList struct {
	clients []chan clientMessage
	sync.Mutex
}
