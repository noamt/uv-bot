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
	measurementReporter := &uv.STDOutMeasurementReporter{}
	measurerAndReporter := uv.GetMeasureAndReportFunction(measurementProvider, measurementReporter)
	measurementSettings := &uv.MeasurementSettings{ExitChan: exitChan, LoopInterval: 2 * time.Second, PollInterval: 2 * time.Minute}

	uv.MeasureAndReport(measurerAndReporter, measurementSettings)
}
