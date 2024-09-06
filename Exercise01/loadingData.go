package main

import (
	"bytes"
	"context"
	"log"
	"strings"

	"encoding/csv"
	"encoding/json"
	"fmt"
	esapi "github.com/elastic/go-elasticsearch/esapi"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"os"
	"strconv"
)

type Place struct {
	ID       int       `json:"id" csv:"omitempty"`
	Name     string    `json:"name"`
	Address  string    `json:"address"`
	Phone    string    `json:"phone"`
	Location geo_point `json:"location"`
}

type geo_point struct {
	Latitude  float64 `json:"lat" csv:"Latitude"`
	Longitude float64 `json:"lon" csv:"Longitude"`
}

func main() {

	file, err := os.Open("data.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	places := []*Place{}

	r := csv.NewReader(file)
	r.Comma = '\t'
	r.Read()
	for record, _ := r.Read(); record != nil; record, _ = r.Read() {
		id, _ := strconv.Atoi(record[0])
		lon, _ := strconv.ParseFloat(record[4], 64)
		lat, _ := strconv.ParseFloat(record[5], 64)
		places = append(places, &Place{ID: id, Name: record[1], Address: record[2], Phone: record[3], Location: geo_point{lat, lon}})
	}

	mapping := `{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
      },
      "mappings": {
  "properties": {
    "name": {
        "type":  "text"
    },
    "address": {
        "type":  "text"
    },
    "phone": {
        "type":  "text"
    },
    "location": {
      "type": "geo_point"
    }
  }
}
  }`

	indexName := "places"

	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	es.Indices.Delete([]string{indexName})

	indexReq := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(string(mapping)),
	}

	resp, err := indexReq.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error doing request: %s", err)
	}
	if resp.IsError() {
		log.Fatalf("Resp is error: %s", resp)
	}

	batch := 250
	count := len(places)
	var buf bytes.Buffer
	baseID := 0

	for i, a := range places {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d" } }%s`, baseID, "\n"))
		baseID++

		data, err := json.Marshal(a)
		if err != nil {
			log.Fatalf("Cannot encode article %d: %s", a.ID, err)
		}
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)

		if i > 0 && i%batch == 0 || i == count-1 {
			res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(indexName))
			if err != nil {
				log.Fatalf("Failure indexing batch : %s", err)
			}
			res.Body.Close()
			buf.Reset()
			if i%1000 == 0 || i == count-1 {
				fmt.Printf("Indexing %d places\n", i)
			}
		}
	}
}
