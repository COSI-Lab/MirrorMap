// server.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
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
	db, err := geoip2.Open("logs/GeoLite2-City.mmdb")
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

var upgrader = websocket.Upgrader{} // use default options
var interrupt2 chan os.Signal
var done2 chan interface{}

func socketHandler(w http.ResponseWriter, r *http.Request) {

	ch := make(chan []byte)

	go func() {
		// defer wg.Done()
		fileIn(ch)

	}()

	interrupt2 = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	done2 = make(chan interface{})
	signal.Notify(interrupt2, os.Interrupt)
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	for {
		val := <-ch
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 1000):
			conn.WriteMessage(1, val)

		case <-interrupt2:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done2:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}

	// // The event loop
	// for {
	// 	messageType, message, err := conn.ReadMessage()
	// 	if err != nil {
	// 		log.Println("Error during message reading:", err)
	// 		break
	// 	}
	// 	m := fmt.Sprintf("%s", message)
	// 	nums := strings.Split(m, ":")
	// 	num1, num2 := nums[0], nums[1]
	// 	log.Printf("Long: %s : Lat: %s, received on 2", num1, num2)

	// 	err = conn.WriteMessage(messageType, message)
	// 	if err != nil {
	// 		log.Println("Error during message writing:", err)
	// 		break
	// 	} else {
	// 		conn.WriteMessage(messageType, []byte("Test"))
	// 	}
	// }
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {

	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe("128.153.165.236:8080", nil))
}
