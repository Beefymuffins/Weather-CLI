package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempF float64 `json:"temp_f"`
		// FeelsLike float64 `json:"feelslike_f"` // Not used
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Hour []struct {
				TimeEpoch int64   `json:"time_epoch"`
				TempF     float64 `json:"temp_f"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain float64 `json:"chance_of_rain"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func main() {
	// Load ENV vars from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file couldn't be loaded")
	}

	// Set Vars
	api_key := os.Getenv("API_KEY")
	base_url := "http://api.weatherapi.com/v1/forecast.json"

	// Parse command line flags
	displayAll := flag.Bool("a", false, "Display all hours")
	flag.Parse()

	// Set default query param (no city supplied as an argument)
	q := os.Getenv("DEFAULT_CITY") // Default value

	// If user passes in a city as an argument set it to q (Also checks for flags, -a)
	// Example: go run main.go London
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		q = os.Args[1]
	}

	// Build the URL
	var url string
	if q != "" {
		url = base_url + "?key=" + api_key + "&q=" + q + "&days=1&aqi=no&alerts=no"
		if *displayAll {
			url += "&a=1"
		}
	} else {
		url = base_url + "?key=" + api_key + "&days=1&aqi=no&alerts=no"
	}

	// Get response
	res, err := http.Get(url)

	// Check for Error
	if err != nil {
		panic(err)
	}

	// Close response body
	defer res.Body.Close()

	// Error if not StatusCode 200
	if res.StatusCode != 200 {
		panic("Weather API not available")
	}

	// Read body
	body, err := io.ReadAll(res.Body)

	// Check for Error
	if err != nil {
		panic(err)
	}

	// Map json response to Weather Struct
	var weather Weather
	err = json.Unmarshal(body, &weather)

	// Check for Error
	if err != nil {
		panic(err)
	}

	location, current, hours := weather.Location, weather.Current, weather.Forecast.Forecastday[0].Hour

	// Change current.TempF text color based on temp
	var tempColorFunc func(format string, a ...interface{}) string
	if current.TempF > 80 {
		tempColorFunc = color.RedString
	} else if current.TempF < 32 {
		tempColorFunc = color.BlueString
	} else {
		tempColorFunc = fmt.Sprintf
	}

	// Print Current Weather
	fmt.Printf(
		"%s, %s: %s, %s\n",
		location.Name,
		location.Country,
		tempColorFunc("%.0fF", current.TempF),
		current.Condition.Text,
	)

	// Print 24 Hour Forecast
	for _, hour := range hours {
		date := time.Unix(hour.TimeEpoch, 0)

		// Display all hours if -a flag is provided
		if *displayAll || date.After(time.Now()) {
			// Change hour.TempF text color based on temp
			var tempColorFunc func(format string, a ...interface{}) string
			if hour.TempF > 80 {
				tempColorFunc = color.RedString
			} else if hour.TempF < 32 {
				tempColorFunc = color.BlueString
			} else {
				tempColorFunc = fmt.Sprintf
			}

			// Change hour.ChanceOfRain text color based on ChanceOfRain %
			var rainColor string
			if hour.ChanceOfRain > 40 {
				rainColor = color.RedString("%.0f%%", hour.ChanceOfRain)
			} else {
				rainColor = fmt.Sprintf("%.0f%%", hour.ChanceOfRain)
			}

			// Assign the values to message
			message := fmt.Sprintf(
				"%s - %s, %s, %s\n",
				date.Format("15:04"),
				tempColorFunc("%.0fF", hour.TempF),
				rainColor,
				hour.Condition.Text,
			)

			fmt.Print(message)
		}
	}
}
