// package main

// import (
// 	"bufio"
// 	"context"
// 	"errors"
// 	"fmt"
// 	"os"
// 	"regexp"
// 	"strings"
// 	"time"

// 	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
// )

// //https://golangdocs.com/reading-files-in-golang

// type ParsedQuery struct {
// 	ip         string
// 	date       time.Time
// 	distro     string
// 	sourcename string
// }

// func parser(line string) (ParsedQuery, error) {

// 	reQuotes := regexp.MustCompile(`"(.*?)"`)
// 	quoteList := reQuotes.FindAllString(line, 3)
// 	reGet := regexp.MustCompile(`\/(.*?)\/`)

// 	listDistro := reGet.FindAllString(line, -1)
// 	nfoundDistro := ""
// 	if len(listDistro) < 2 {
// 		return ParsedQuery{"", time.Now(), "", ""}, errors.New("wrong Distro")
// 	} else {
// 		foundDistro := strings.SplitN(listDistro[1], " ", -1)
// 		nfoundDistro = strings.Join(foundDistro, "")
// 		nfoundDistro = strings.Replace(nfoundDistro, "/", "", -1)
// 	}
// 	userAgent := quoteList[2]
// 	foundIp := strings.SplitN(line, "-", 2)[0]
// 	reDateTime := regexp.MustCompile(`\[.*\]`)
// 	foundDate := reDateTime.FindString(line)

// 	t := "[02/Jan/2006:15:04:05 -0700]"
// 	tm, _ := time.Parse(t, foundDate)

// 	FinishedQuery := ParsedQuery{
// 		ip:         foundIp,
// 		date:       tm,
// 		distro:     nfoundDistro,
// 		sourcename: userAgent,
// 	}
// 	return FinishedQuery, nil

// }

// var token = os.Getenv("TOKEN")
// var bucket = os.Getenv("BUCKET")
// var url = os.Getenv("URL")
// var org = os.Getenv("ORG")

// // Create a new client using an InfluxDB server base URL and an authentication token
// var client = influxdb2.NewClient(url, token)

// // Use blocking write client for writes to desired bucket
// var writeAPI = client.WriteAPIBlocking(org, bucket)

// func sendToDb(q ParsedQuery) {
// 	pList := q

// 	// write line protocol
// 	p := influxdb2.NewPointWithMeasurement("stat").
// 		AddTag("unit", "download").
// 		AddField("distro", pList.distro).
// 		AddField("date", pList.date).
// 		SetTime(time.Now())
// 	writeAPI.WritePoint(context.Background(), p)

// }

// func fileIn() []ParsedQuery {
// 	file, err := os.Open("access.log.2")
// 	if err != nil {
// 		fmt.Println("Error opening file")
// 	}
// 	defer file.Close()

// 	var pList = []ParsedQuery{}

// 	scanner := bufio.NewScanner(file)

// 	// i := 0
// 	for scanner.Scan() {
// 		p, _ := parser(scanner.Text())
// 		sendToDb(p)
// 	}
// 	return pList
// }

// func main() {

// 	start := time.Now()

// 	fileIn()

// 	elasped := time.Since(start)
// 	fmt.Println("")
// 	fmt.Printf("Took %s", elasped)
// 	fmt.Println("")
// }
