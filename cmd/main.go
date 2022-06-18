package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

const MY_LAT = 41.151058
const MY_LONG = -95.900253

type ISSResponse struct {
	Timestamp   int
	Message     string
	ISSPosition Position `json:"iss_position"`
}

type Position struct {
	Latitude  float64
	Longitude float64
}

func main() {
	var withinFiveDeg bool = issWithinFiveDeg(MY_LAT, MY_LONG)
	fmt.Println("Within Five Degrees:", withinFiveDeg)
}

// Gets current position of the ISS and determines if it's position is
// within 5 degrees of the supplied location
func issWithinFiveDeg(lat float64, lon float64) bool {
	resp, err := http.Get("http://api.open-notify.org/iss-now.json")
	if err != nil {
		log.Fatalln(err)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("ISS API Response: ", string(responseData))

	var issResponse ISSResponse
	json.Unmarshal(responseData, &issResponse)

	var latDiff = math.Abs(issResponse.ISSPosition.Latitude - lat)
	var longDiff = math.Abs(issResponse.ISSPosition.Longitude - lon)
	if latDiff <= 5 && longDiff <= 5 {
		return true
	} else {
		return false
	}
}
