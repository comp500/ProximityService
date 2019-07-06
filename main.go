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

	box := packr.NewBox("./web/dist")

	http.Handle("/", http.FileServer(box))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		_ = conn
	})

	dataChannel := make(chan []byte)
	done := make(chan bool)

	go handleData(dataChannel, done)
	go startBluetooth(dataChannel, done)

	fmt.Printf("Starting server on port %d\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

	done <- true
}

func handleData(dataChannel chan []byte, done chan bool) {
	rcvPart1 := false
	var part1Analog int
	for {
		select {
		case <-done:
			return
		case data := <-dataChannel:
			fmt.Printf("%X\n", data)
			for _, b := range data {
				if (b & 0x80) == 0x80 {
					binaryValue := (b & 0x40) == 0x40
					if binaryValue {
						fmt.Println("Read binary: HIGH")
					} else {
						fmt.Println("Read binary: LOW")
					}

					rcvPart1 = true
					part1Analog = int(b&0x07) << 7
				} else {
					if rcvPart1 {
						rcvPart1 = false
						analogValue := int(b) | part1Analog
						fmt.Println("Read analog: %d", analogValue)
					}
				}
			}
		}
	}
}
