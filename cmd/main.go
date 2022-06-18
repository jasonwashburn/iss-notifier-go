package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	var isDarkOutside bool = isDark(MY_LAT, MY_LONG)
	fmt.Println("Is dark outside:", isDarkOutside)
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

func isDark(lat float64, lon float64) bool {
	params := url.Values{}
	params.Add("lat", strconv.FormatFloat(lat, 'G', -1, 32))
	params.Add("lng", strconv.FormatFloat(lon, 'G', -1, 32))
	params.Add("formatted", "0")
	api_url := url.URL{
		Scheme:   "https",
		Host:     "api.sunrise-sunset.org",
		Path:     "json",
		RawQuery: params.Encode(),
	}

	type Times struct {
		AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
		AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
	}

	type SunriseSunsetResponse struct {
		Status  string
		Results Times `json:"results"`
	}

	resp, err := http.Get(api_url.String())
	if err != nil {
		log.Fatalln(err)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}


	var ssResponse SunriseSunsetResponse
	json.Unmarshal(responseData, &ssResponse)
	dateFormat := "2006-01-02T15:04:05-07:00"
	twilightBegin, err := time.Parse(dateFormat, ssResponse.Results.AstronomicalTwilightBegin)
	if err != nil {
		log.Fatalln(err)
	}
	twilightEnd, err := time.Parse(dateFormat, ssResponse.Results.AstronomicalTwilightEnd)
	if err != nil {
		log.Fatalln(err)
	}
	utcNow := time.Now().UTC()

	if utcNow.Hour() >= twilightBegin.Hour() && utcNow.Hour() <= twilightEnd.Hour() {
		return true
	} else {
		return false
	}
}
