package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

func parser(line string) string {
	foundIp := strings.SplitN(line, "-", 2)[0]
	return foundIp
}

func fileIn() []string {
	file, err := os.Open("access.log.2")
	if err != nil {
		fmt.Println("Error opening file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	pList := []string{}
	i := 0
	for scanner.Scan() {
		p := parser(scanner.Text())
		pList = append(pList, p)
		if i%100000 == 0 {
			fmt.Println(i)
		}
		i++
	}
	return pList
}

func main() {
	start := time.Now()

	ip := fileIn()
	fmt.Println("File read and parsed")
	// fmt.Println(ip)
	type LongLat struct {
		long float64
		lat  float64
	}

	// db, err := ip2location.OpenDB("./IP-LATITUDE-LONGITUDE.BIN")
	listLongLat := []LongLat{}
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		fmt.Print(err)
		return
	}
	for i := 0; i < len(ip); i++ {
		ipNew := net.ParseIP(strings.TrimSpace(ip[i]))
		results, err := db.City(ipNew)
		if err != nil {
			fmt.Print(err)
			return
		}
		q := LongLat{results.Location.Latitude, results.Location.Longitude}
		listLongLat = append(listLongLat, q)
		if i%100000 == 0 {
			fmt.Println(i)
		}

	}

	fmt.Println("Done")
	elasped := time.Since(start)
	fmt.Println("")
	fmt.Printf("Took %s", elasped)
	fmt.Println("")
}
