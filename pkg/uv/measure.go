package uv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var latestIndexForLocation = map[string]float32{}
var lastPoll time.Time

type MeasurementSettings struct {
	ExitChan     <-chan bool
	LoopInterval time.Duration
	PollInterval time.Duration
}

func MeasureAndReport(measurerReporter MeasurerReporter, measurementSettings *MeasurementSettings) {
Loop:
	for {
		select {
		case <-measurementSettings.ExitChan:
			log.Println("Received exit signal")
			break Loop
		default:
			if lastPoll.IsZero() || time.Since(lastPoll) >= measurementSettings.PollInterval {
				log.Println("Measuring UV index")
				MeasureAndReportLocations(Locations, measurerReporter)
				lastPoll = time.Now()
			}
			time.Sleep(measurementSettings.LoopInterval)
		}
	}
}

type MeasurerReporter func(location *Location) error

func MeasureAndReportLocations(locations []*Location, measurerReporter MeasurerReporter) {
	for i := 0; i < len(locations); i++ {
		locationToMeasure := locations[i]
		err := measurerReporter(locationToMeasure)
		if err != nil {
			log.Println(fmt.Errorf("failed to measurer and report %s: %w", locationToMeasure.DisplayName, err))
			continue
		}
	}
}

func GetMeasureAndReportFunction(measurementProvider MeasurementProvider, reporter MeasurementReporter) MeasurerReporter {
	return func(location *Location) error {
		uvIndex, measurementError := measurementProvider.Measure(location)
		if measurementError != nil {
			return fmt.Errorf("failed to get UV index for %s: %w", location.DisplayName, measurementError)
		}
		lastMeasurement := latestIndexForLocation[location.DisplayName]
		severityChangedSinceLastMeasurement := IndexHasChanged(lastMeasurement, uvIndex)
		if severityChangedSinceLastMeasurement {
			reportError := reporter.Report(location, uvIndex)
			if reportError != nil {
				return fmt.Errorf("failed to report UV index for %s: %w", location.DisplayName, reportError)
			}
			latestIndexForLocation[location.DisplayName] = uvIndex
		}
		return nil
	}
}

type OneCallCurrent struct {
	UVI float32 `json:"uvi"`
}

type OneCallResponse struct {
	Current *OneCallCurrent `json:"current"`
}

type MeasurementProvider interface {
	Measure(locationToMeasure *Location) (float32, error)
}

type OpenWeatherMap struct {
	Host  string
	AppID string
}

func (openweathermap *OpenWeatherMap) Measure(locationToPoll *Location) (float32, error) {
	client := http.DefaultClient

	req, requestError := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/data/2.5/onecall", openweathermap.Host), nil)
	if requestError != nil {
		return 0, fmt.Errorf("failed to prepare HTTP request: %w", requestError)
	}

	q := req.URL.Query()

	q.Add("lat", locationToPoll.Latitude)
	q.Add("lon", locationToPoll.Longitude)

	q.Add("appid", openweathermap.AppID)
	q.Add("exclude", "minutely,hourly,alerts,daily")
	req.URL.RawQuery = q.Encode()
	resp, requestExecError := client.Do(req)
	if requestExecError != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", requestExecError)
	}
	dec := json.NewDecoder(resp.Body)
	ocr := OneCallResponse{}
	jsonErr := dec.Decode(&ocr)
	if jsonErr != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w", jsonErr)
	}
	return ocr.Current.UVI, nil
}

func IndexHasChanged(latestUVIndex float32, newIndex float32) bool {
	if latestUVIndex == 0 {
		return true
	} else if newIndex < 3.0 && latestUVIndex > 3.0 {
		return true
	} else if (newIndex >= 3.0 && newIndex < 8.0) && (latestUVIndex < 3.0 || latestUVIndex >= 8.0) {
		return true
	} else if newIndex >= 8 && latestUVIndex < 8.0 {
		return true
	}
	return false
}

type MeasurementReporter interface {
	Report(locationToReport *Location, uvIndex float32) error
}

type STDOutMeasurementReporter struct{}

func (measurementReporter *STDOutMeasurementReporter) Report(locationToReport *Location, uvIndex float32) error {
	alert := getAlert(locationToReport, uvIndex)
	fmt.Println(alert)
	return nil
}

func NewTwitterMeasurementReporter(twitterAuth *TwitterAuth) *TwitterMeasurementReporter {
	config := oauth1.NewConfig(twitterAuth.ConsumerKey, twitterAuth.ConsumerSecret)
	token := oauth1.NewToken(twitterAuth.AccessToken, twitterAuth.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	return &TwitterMeasurementReporter{client: client}
}

type TwitterAuth struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type TwitterMeasurementReporter struct {
	client *twitter.Client
}

func (t *TwitterMeasurementReporter) Report(locationToReport *Location, uvIndex float32) error {
	alert := getAlert(locationToReport, uvIndex)
	_, response, tweetError := t.client.Statuses.Update(alert, nil)
	if tweetError != nil {
		return fmt.Errorf("failed to tweet '%s': %w", alert, tweetError)
	}
	if response.StatusCode >= http.StatusBadRequest {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("failed to tweet '%s'. Response code: %d. Body: %s", alert, response.StatusCode, string(body))
	}
	return nil
}

func getAlert(locationToReport *Location, uvIndex float32) string {
	alerts := AltertsByLocation[locationToReport.DisplayName]
	if uvIndex < 3.0 {
		return alerts.Low(uvIndex)
	} else if uvIndex < 8.0 {
		return alerts.Moderate(uvIndex)
	}
	return alerts.High(uvIndex)
}
