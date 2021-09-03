package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
)

//https://golangdocs.com/reading-files-in-golang

type ParsedQuery struct {
	ip         string
	date       time.Time
	distro     string
	sourcename string
}

func parser(line string) (ParsedQuery, error) {

	// fmt.Println(line)
	reQuotes := regexp.MustCompile(`"(.*?)"`)
	quoteList := reQuotes.FindAllString(line, 3)
	reGet := regexp.MustCompile(`\/(.*?)\/`)
	// fmt.Println(line)
	listDistro := reGet.FindAllString(line, -1)
	nfoundDistro := ""
	if len(listDistro) < 2 {
		return ParsedQuery{"", time.Now(), "", ""}, errors.New("wrong Distro")
	} else {
		foundDistro := strings.SplitN(listDistro[1], " ", -1)
		nfoundDistro = strings.Join(foundDistro, "")
		nfoundDistro = strings.Replace(nfoundDistro, "/", "", -1)
	}
	// getCommand := quoteList[0]
	userAgent := quoteList[2]
	foundIp := strings.SplitN(line, "-", 2)[0]
	reDateTime := regexp.MustCompile(`\[.*\]`)
	foundDate := reDateTime.FindString(line)

	t := "[02/Jan/2006:15:04:05 -0700]"
	tm, err := time.Parse(t, foundDate)
	if err != nil {
		fmt.Print("time date ")
		fmt.Println(err)
	}

	// fmt.Println(nfoundDistro)

	// findIpv4 := regexp.MustCompile(`\b((([0-2]\d[0-5])|(\d{2})|(\d))\.){3}(([0-2]\d[0-5])|(\d{2})|(\d))\b`)
	// findIPv6 := regexp.MustCompile(`(([a-fA-F0-9]{1,4}|):){1,7}([a-fA-F0-9]{1,4}|:)`)
	// // fmt.Println(findIpv4.FindString(line))
	// // fmt.Println(findIPv6.FindString(line))
	// foundipv4 := findIpv4.FindString(line)
	// foundipv6 := findIPv6.FindString(line)
	// if foundipv4 != "" {
	// 	fmt.Println(foundipv4)
	// }
	// if foundipv6 != "" {
	// 	fmt.Println(foundipv6)
	// }
	FinishedQuery := ParsedQuery{
		ip:         foundIp,
		date:       tm,
		distro:     nfoundDistro,
		sourcename: userAgent,
	}
	return FinishedQuery, nil

}

func fileIn() []ParsedQuery {
	file, err := os.Open("access.log.2")
	if err != nil {
		fmt.Println("Error opening file")
	}
	defer file.Close()

	var pList = []ParsedQuery{}

	scanner := bufio.NewScanner(file)

	i := 0
	for scanner.Scan() {
		if i > 10 {
			break
		}
		// fmt.Println(scanner.Text())
		p, err := parser(scanner.Text())
		if err != nil {
			fmt.Print("from 87 ")
			fmt.Println(err)
			i++
			continue
		} else {
			pList = append(pList, p)
			i++
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Print("From 97 ")
		fmt.Println(err)
	}
	return pList
}

func main() {

	pList := fileIn()
	fmt.Println(pList[0].date)

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some Error Occured. Err: %s", err)
	}

	token := os.Getenv("TOKEN")
	bucket := os.Getenv("BUCKET")
	url := os.Getenv("URL")
	org := os.Getenv("ORG")

	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(url, token)
	// Use blocking write client for writes to desired bucket
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// get non-blocking write client
	// writeAPI := client.WriteAPI(org, bucket)

	// write line protocol
	for i := 0; i < len(pList); i++ {
		p := influxdb2.NewPointWithMeasurement("stat").
			AddTag("unit", "download").
			AddField("distro", pList[i].distro).
			AddField("date", pList[i].date).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), p)
		// Flush writes
		// writeAPI.Flush()
	}

	/*
		// Get query client
		queryAPI := client.QueryAPI(org)
		// Get parser flux query result
		result, err := queryAPI.Query(context.Background(), `from(bucket:"MirrorData")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
		if err == nil {
			// Use Next() to iterate over query result lines
			for result.Next() {
				// Observe when there is new grouping key producing new table
				if result.TableChanged() {
					fmt.Printf("table: %s\n", result.TableMetadata().String())
				}
				// read result
				fmt.Printf("row: %s\n", result.Record().String())
			}
			if result.Err() != nil {
				fmt.Printf("Query error: %s\n", result.Err().Error())
			}
		}
		// Ensures background processes finishes
		client.Close()
	*/
}
