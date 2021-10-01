package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oschwald/geoip2-golang"
)

func parser(line string) string {
	foundIp := strings.SplitN(line, "-", 2)[0]
	return foundIp
}

type longLat struct {
	Long float64
	Lat  float64
}

func fileIn(ch chan []byte) {

	scanner := bufio.NewScanner(os.Stdin)
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err)
		return
	}

	// buf := bytes.NewBuffer(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ip := parser(line)
		// fmt.Println(ip)
		if ip == "" {
			continue
		}
		ipNew := net.ParseIP(strings.TrimSpace(ip))
		results, err := db.City(ipNew)
		if err != nil {
			fmt.Println(err)
			return
		}
		long := results.Location.Longitude
		lat := results.Location.Latitude
		// fmt.Println(long, lat)
		ch <- []byte(fmt.Sprintf("%f:%f", long, lat))
	}
}

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func main() {

	ch := make(chan []byte)

	// fileIn(ch)

	// ch := make(chan []byte)
	// wg.Add(1)
	go func() {
		// defer wg.Done()
		fileIn(ch)

	}()
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "ws://localhost:8080" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		val := <-ch
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 1000):
			// Send an echo packet every second
			// x := <-ch
			// y := IntToByteArray(x)
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
			err := conn.WriteMessage(websocket.TextMessage, val)
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}

}
