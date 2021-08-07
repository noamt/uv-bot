package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noamt/uv-bot/pkg/uv"
)

func main() {
	exitChan := make(chan bool)

	go func() {
		log.Println("Listening for signals...")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		<-c
		exitChan <- true
	}()

	appID := os.Getenv("OPENWEATHER_MAP_APP_ID")
	if appID == "" {
		log.Fatalln("An OpenWeather Map app ID is required. Please set the OPENWEATHER_MAP_APP_ID env var")
	}
	measurementProvider := &uv.OpenWeatherMap{Host: "https://api.openweathermap.org", AppID: appID}

	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	if consumerKey == "" {
		log.Fatalln("A Twitter consumer key is required. Please set the TWITTER_CONSUMER_KEY env var")
	}

	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	if consumerSecret == "" {
		log.Fatalln("A Twitter consumer secret is required. Please set the TWITTER_CONSUMER_SECRET env var")
	}

	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalln("A Twitter access token is required. Please set the TWITTER_ACCESS_TOKEN env var")
	}

	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")
	if accessSecret == "" {
		log.Fatalln("A Twitter access secret is required. Please set the TWITTER_ACCESS_SECRET env var")
	}
	twitterAuth := &uv.TwitterAuth{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    accessToken,
		AccessSecret:   accessSecret,
	}
	measurementReporter := uv.NewTwitterMeasurementReporter(twitterAuth)
	measurerAndReporter := uv.GetMeasureAndReportFunction(measurementProvider, measurementReporter)
	measurementSettings := &uv.MeasurementSettings{ExitChan: exitChan, LoopInterval: 2 * time.Second, PollInterval: 2 * time.Minute}

	uv.MeasureAndReport(measurerAndReporter, measurementSettings)
}
