package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/oschwald/geoip2-golang"

	log "github.com/schollz/logger"
)

func parser(line string) string {
	foundIp := strings.SplitN(line, "-", 2)[0]
	return foundIp
}

type longLat struct {
	Long float64
	Lat  float64
}

var clients = make(map[*websocket.Conn]bool)
var channel = make(chan longLat)

func fileIn() {

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
		fmt.Println(ip)
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
		channel <- longLat{long, lat}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	go fileIn()

	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/longlat", longLatHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	go echo()

	log.Fatal(http.ListenAndServe(":8844", router))

	// fileIn(channel)

	/*
		Map of websockets map is dict in python
		As websocket is created add to map
		have data iterate over websockets sending it to each

	*/

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "home")
}

func writer(coord longLat) {
	channel <- coord
}

func longLatHandler(w http.ResponseWriter, r *http.Request) {
	var coordinates longLat
	if err := json.NewDecoder(r.Body).Decode(&coordinates); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	defer r.Body.Close()
	go writer(coordinates)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// register client
	clients[ws] = true
}

func echo() {
	for {
		val := <-channel
		latlong := fmt.Sprintf("%f %f %s", val.Lat, val.Long)
		// send to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(latlong))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
