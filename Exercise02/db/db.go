package db

import (
	"fmt"
	"strings"
	"context"
	"encoding/json"
	"log"
	"time"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Result struct {
	ScrollID string `json:"_scroll_id"`
	Took     int
	Hits     struct {
		Total struct {
			Value int
		}
		Hits []struct {
			Source Place `json:"_source"`
		}
	}
	Aggregations map[string]interface{} `json:"aggregations"`
}

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

type DataBase struct {
	Name string
	Lat  float64
	Lon  float64
}

type CountResp struct {
	Count int `json:"count"`
}

func (db *DataBase) GetPlaces(limit int, offset int) ([]Place, int, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal("error in db.go 36: ", err)
	}

	query := getQuery(limit, offset, db.Lat, db.Lon)

	res, _ := es.Search(
		es.Search.WithIndex(db.Name),
		es.Search.WithSort("_doc"),
		es.Search.WithSize(limit),
		es.Search.WithScroll(time.Minute),
		es.Search.WithBody(strings.NewReader(query)),
	)

	defer res.Body.Close()

	results := Result{}
	err = json.NewDecoder(res.Body).Decode(&results)
	if err != nil {
		log.Fatal("places json decoding: ", err)
	}

	places := make([]Place, 0, 10)
	for _, elem := range results.Hits.Hits {
		places = append(places, elem.Source)
	}

	countReq := esapi.CountRequest{
		Index: []string{db.Name},
	}

	re, err := countReq.Do(context.Background(), es)
	if err != nil {
		log.Printf("Count request err: %s", err)
	}
	defer re.Body.Close()

	var countResp CountResp
	if err = json.NewDecoder(re.Body).Decode(&countResp); err != nil {
		log.Printf("Count decode err: %s", err)
	}

	return places, countResp.Count, nil
}


func getQuery(limit, offset int, lat, lon float64) string {
	if limit == 3 {
		return fmt.Sprintf(`{"sort": [
    {
      "_geo_distance": {
        "location": {
          "lat": %f,
          "lon": %f
        	},
        "order": "asc",
        "unit": "km",
        "mode": "min",
        "distance_type": "arc",
        "ignore_unmapped": true
      }
    }
		]}`, lat, lon)
	}
	return fmt.Sprintf(`{
		"query": {
			"range": {
				"_seq_no": {
					"gte": %d,
					"lte": %d
				}
			}	
		}
	}`, limit*offset, limit*offset+limit-1)
}
