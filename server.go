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
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/oschwald/geoip2-golang"
	"github.com/thanhpk/randstr"
)

// Globals
var clients map[string]chan []byte
var clients_lock sync.RWMutex

var upgrader = websocket.Upgrader{} // use default options
var interrupt2 chan os.Signal
var done2 chan interface{}

func getIp(line string) string {
	foundIp := strings.SplitN(line, "-", 2)[0]
	return foundIp
}

func fileIn(clients map[string]chan []byte) {
	db, err := geoip2.Open("logs/GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err)
		return
	}
	distList := []string{"alpine", "archlinux", "archlinux32", "artix-linux", "blender", "centos", "clonezilla", "cpan", "cran", "ctan", "cygwin", "debian", "debian-cd", "eclipse", "freebsd", "gentoo", "gentoo-portage", "gparted", "ipfire", "isabelle", "linux", "linuxmint", "manjaro", "msys2", "odroid", "openbsd", "opensuse", "parrot", "raspbian", "RebornOS", "ros", "sabayon", "serenity", "slackware", "slitaz", "tdf", "templeos", "ubuntu", "ubuntu-cdimage", "ubuntu-ports", "ubuntu-releases", "videolan", "voidlinux", "zorinos"}
	distMap := make(map[string]int)
	for i, dist := range distList {
		distMap[dist] = i
	}
	// Create a map of dists and give them an id, hashing a map is quicker than an array

	scanner := bufio.NewScanner(os.Stdin)
	// Iterate through stdin
	for scanner.Scan() {
		line := scanner.Text()
		ip := getIp(line)
		if ip == "" {
			continue
		}
		//make sure ip is valid ip
		ipNew := net.ParseIP(strings.TrimSpace(ip))
		results, err := db.City(ipNew)
		if err != nil {
			fmt.Println(err)
			return
		}
		// get distro, use regex to find the distro
		reDist := regexp.MustCompile(`\/(.*?)\/`)
		listDistro := reDist.FindAllString(line, -1)
		nfoundDistro := ""
		if len(listDistro) < 2 {
			continue
		}
		foundDistro := strings.SplitN(listDistro[1], " ", -1)
		nfoundDistro = strings.Join(foundDistro, "")
		nfoundDistro = strings.Replace(nfoundDistro, "/", "", -1)
		// do some formating to distro to make it so I can hash it
		long := results.Location.Longitude
		lat := results.Location.Latitude
		fmt.Println(long, lat)
		//convert lat to string

		distByte := byte(distMap[nfoundDistro])
		longByte := byte(long)
		latByte := byte(lat)

		// convert long to uint64

		// convert dist, lat, long to byte

		// turn dist, long, and lat to byte array to send
		msg := []byte{distByte, longByte, latByte}

		clients_lock.Lock()
		for _, ch := range clients {
			// Add msg to channel for sending messages
			// Have to do it this way as websocket handler is seperate function
			ch <- msg
		}
		clients_lock.Unlock()
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Handles the websocket
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		w.WriteHeader(404)
		return
	}

	// get the channel
	ch := clients[id]

	log.Printf("%s connected!\n", id)

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}

	defer func() {
		// Close connection gracefully
		conn.Close()
		log.Printf("%s disconnected!", id)
	}()

	for {
		// Reciever byte array
		val := <-ch
		// Send message across websocket
		err = conn.WriteMessage(2, val)
		if err != nil {
			// If err, lock client list while removing from it
			clients_lock.Lock()
			log.Printf("Error sending message %s : %s", id, err)
			delete(clients, id)
			clients_lock.Unlock()
			return
		}
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	id := randstr.Hex(16)
	// Create UUID but badly
	// Should work as we arent serving enough clients were psuedo random will mess us up

	clients_lock.Lock()
	clients[id] = make(chan []byte)
	// When someone registers add them to the client list
	clients_lock.Unlock()
	log.Printf("new connection registered: %s\n", id)

	// Send id to client
	w.WriteHeader(200)
	w.Write([]byte(id))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Return list of active clients
	// Mostly for diagnostic purposes
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprint(len(clients))))
}

type HTMLStrippingFileSystem struct {
	http.FileSystem
}

func main() {
	// Create a type safe Map for strings to channels
	clients = make(map[string]chan []byte)

	interrupt := make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(1)
	}()

	// Read from standard in and pass cordinates to each client
	go fileIn(clients)

	// gorilla/mux router
	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/register", registerHandler)
	r.HandleFunc("/socket/{id}", socketHandler)

	// Handle homepage, ugly but works
	r.PathPrefix("/").Handler(http.FileServer(HTMLStrippingFileSystem{http.Dir("static")})).Methods("GET")

	// Serve on 8080
	s := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	log.Printf("Serving on localhost:%d", 8000)
	log.Fatalf("%s", s.ListenAndServe())
}
