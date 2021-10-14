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
	"time"

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
	distList := []string{"alpine", "archlinux", "archlinux32", "artix-linux", "blender", "centos", "clonezilla", "cpan", "cran", "ctan", "cygwin", "debian", "debian-cd", "debian-security", "eclipse", "fedora", "fedora-epel", "freebsd", "gentoo", "gentoo-portage", "gnu", "gparted", "ipfire", "isabelle", "linux", "linuxmint", "manjaro", "msys2", "odroid", "openbsd", "opensuse", "parrot", "raspbian", "RebornOS", "ros", "sabayon", "serenity", "slackware", "slitaz", "tdf", "templeos", "ubuntu", "ubuntu-cdimage", "ubuntu-ports", "ubuntu-releases", "videolan", "voidlinux", "zorinos"}
	distMap := make(map[string]int)
	for i, dist := range distList {
		distMap[dist] = i
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ip := getIp(line)
		if ip == "" {
			continue
		}
		ipNew := net.ParseIP(strings.TrimSpace(ip))
		results, err := db.City(ipNew)
		if err != nil {
			fmt.Println(err)
			return
		}
		reGet := regexp.MustCompile(`\/(.*?)\/`)
		listDistro := reGet.FindAllString(line, -1)
		nfoundDistro := ""
		foundDistro := strings.SplitN(listDistro[1], " ", -1)
		nfoundDistro = strings.Join(foundDistro, "")
		nfoundDistro = strings.Replace(nfoundDistro, "/", "", -1)

		long := results.Location.Longitude
		lat := results.Location.Latitude
		// fmt.Println(long, lat)

		clients_lock.Lock()
		for _, ch := range clients {
			ch <- []byte(fmt.Sprintf("%f:%f:%d", long, lat, distMap[nfoundDistro]))
		}
		clients_lock.Unlock()
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		w.WriteHeader(404)
		return
	}

	// get the channel
	ch := clients[id]

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	for {
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 500):
			val := <-ch
			err = conn.WriteMessage(1, val)
			if err != nil {
				delete(clients, id)
				return
			}
		}
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	id := randstr.Hex(16)

	clients_lock.Lock()
	clients[id] = make(chan []byte)
	clients_lock.Unlock()
	fmt.Printf("id created: %s\n", id)

	// Send id to client
	w.WriteHeader(200)
	w.Write([]byte(id))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Send id to client
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprint(len(clients))))
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
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
	// r.HandleFunc("/", home)

	r.PathPrefix("/").Handler(http.FileServer(HTMLStrippingFileSystem{http.Dir("static")})).Methods("GET")
	// Serve on 8080
	s := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	log.Fatalf("%s", s.ListenAndServe())
}
