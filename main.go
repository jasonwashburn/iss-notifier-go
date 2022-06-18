package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

const MY_LAT = 41.151058
const MY_LONG = -95.900253

type ISSResponse struct {
	Timestamp   int
	Message     string
	ISSPosition Position `json:"iss_position"`
}

type Position struct {
	Latitude  string
	Longitude string
}

// A script that can be run repeatedly via something like cron to automatically
// send you an email if the ISS is currently overhead and it's dark outside
func main() {
	var withinFiveDeg bool = issWithinFiveDeg(MY_LAT, MY_LONG)
	log.Println("Within Five Degrees:", withinFiveDeg)
	var isDarkOutside bool = isDark(MY_LAT, MY_LONG)
	log.Println("Is dark outside:", isDarkOutside)

	type Config struct {
		FromEmail   string `yaml:"fromEmail"`
		Password    string `yaml:"password"`
		TargetEmail string `yaml:"targetEmail"`
	}

	if !withinFiveDeg && !isDarkOutside {
		f, err := os.Open("config.yml")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var cfg Config
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&cfg)
		if err != nil {
			log.Fatal(err)
		}

		issArt, err := ioutil.ReadFile("iss_art.txt")
		if err != nil {
			log.Fatal(err)
		}
		emailMessage := string(issArt) + "Look up! The ISS is Overhead!"

		log.Print("Sending Message;", emailMessage)
		sendEmail(cfg.TargetEmail, emailMessage, cfg.FromEmail, cfg.Password)
	}
}

// Gets current position of the ISS and determines if it's position is
// within 5 degrees of the supplied location
func issWithinFiveDeg(lat float64, lon float64) bool {
	log.Println("Retrieving ISS position...")
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
	log.Println("API Response:", string(responseData))
	issLatitude, err := strconv.ParseFloat(issResponse.ISSPosition.Latitude, 32)
	if err != nil {
		log.Fatal(err)
	}
	issLongitude, err := strconv.ParseFloat(issResponse.ISSPosition.Longitude, 32)
	if err != nil {
		log.Fatal(err)
	}
	var latDiff = math.Abs(issLatitude - lat)
	var longDiff = math.Abs(issLongitude - lon)
	if latDiff <= 5 && longDiff <= 5 {
		return true
	} else {
		return false
	}
}

// Retrieves the current UTC astronomical twilight beginning and end times, then compares
// it to the current time in UTC to see if it's dark outside
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
	log.Println("Fetching twilight start and end times...")
	resp, err := http.Get(api_url.String())
	if err != nil {
		log.Fatalln(err)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("API Response:", string(responseData))
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

// Sends an email using the gmail smtp server
func sendEmail(toAddress string, message string, fromAddress string, password string) {
	log.Println("Sending email...")
	recipients := []string{toAddress}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	byteMessage := []byte(message)

	auth := smtp.PlainAuth("", fromAddress, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromAddress, recipients, byteMessage)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success!")
}
