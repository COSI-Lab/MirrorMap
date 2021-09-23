package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	log "github.com/schollz/logger"
	"github.com/schollz/websocket"
	"github.com/schollz/websocket/wsjson"
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*120000)
	defer cancel()

	for {
		var v interface{}
		err = wsjson.Read(ctx, c, &v)
		if err != nil {
			break
		}
		log.Debugf("received: %v", v)
		// log.Printf("recieved: %v", v)
		err = wsjson.Write(ctx, c, struct{ Message string }{
			"hello, browser",
		})
		if err != nil {
			break
		}
	}
	if websocket.CloseStatus(err) == websocket.StatusGoingAway {
		err = nil
	}
	c.Close(websocket.StatusNormalClosure, "")
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%v %v %v %s\n", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

func handle(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// very special paths
	if r.URL.Path == "/ws" {
		return handleWebsocket(w, r)
	} else {
		b, _ := ioutil.ReadFile("index.html")
		w.Write(b)
	}

	return
}

func parser(line string) string {
	foundIp := strings.SplitN(line, "-", 2)[0]
	return foundIp
}

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
		fmt.Println(long, lat)
	}
}

// func stuff() {
// 	// fileIn()
// 	fmt.Println("File read and parsed")
// 	// fmt.Println(ip)
// 	type LongLat struct {
// 		long float64
// 		lat  float64
// 	}

// 	// db, err := ip2location.OpenDB("./IP-LATITUDE-LONGITUDE.BIN")
// 	listLongLat := []LongLat{}

// 	if err != nil {
// 		fmt.Print(err)
// 		return
// 	}
// 	for i := 0; i < len(ip); i++ {
// 		ipNew := net.ParseIP(strings.TrimSpace(ip[i]))
// 		results, err := db.City(ipNew)
// 		if err != nil {
// 			fmt.Print(err)
// 			return
// 		}
// 		q := LongLat{results.Location.Latitude, results.Location.Longitude}
// 		listLongLat = append(listLongLat, q)
// 		if i%100000 == 0 {
// 			fmt.Println(i)
// 		}

// 	}
// }

//d0dfc00032243627562af6aaa7821189
func main() {
	// scanner := bufio.NewReader(os.Stdin)
	// for {

	// 	line, _ := scanner.ReadString('\n')
	// 	fmt.Println(line)
	// 	p := parser(line)
	// 	fmt.Println(p)
	// }
	fileIn()
	// start := time.Now()
	// log.SetLevel("debug")
	// port := 8098
	// log.Infof("listening on :%d", port)
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	// // fmt.Println(listLongLat)

	// fmt.Println("Done")
	// elasped := time.Since(start)
	// fmt.Println("")
	// fmt.Printf("Took %s", elasped)
	// fmt.Println("")
}
