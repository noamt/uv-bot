package uv_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/noamt/uv-bot/pkg/uv"
)

type testMeasurementProvider struct {
	FailOnLocation         map[string]bool
	MeasurementForLocation map[string]float32
	MeasuredLocations      []string
}

func (t *testMeasurementProvider) Measure(locationToMeasure *uv.Location) (float32, error) {
	if t.FailOnLocation[locationToMeasure.DisplayName] {
		return 0, errors.New("something happened")
	}
	t.MeasuredLocations = append(t.MeasuredLocations, locationToMeasure.DisplayName)
	return t.MeasurementForLocation[locationToMeasure.DisplayName], nil
}

type testMeasurementReporter struct {
	FailOnLocation    map[string]bool
	ReportedLocations map[string]float32
}

func (t *testMeasurementReporter) Report(locationToReport *uv.Location, uvIndex float32) error {
	if t.FailOnLocation[locationToReport.DisplayName] {
		return errors.New("something happened")
	}
	t.ReportedLocations[locationToReport.DisplayName] = uvIndex
	return nil
}

func TestMeasureAndReport(t *testing.T) {
	measurerReporterCalled := false
	measurerReporter := func(location *uv.Location) error {
		measurerReporterCalled = true
		return nil
	}

	exitChan := make(chan bool)
	measureAndReportExitChan := make(chan bool)
	settings := &uv.MeasurementSettings{LoopInterval: 100 * time.Millisecond, PollInterval: 100 * time.Millisecond, ExitChan: exitChan}
	go func() {
		uv.MeasureAndReport(measurerReporter, settings)
		measureAndReportExitChan <- true
	}()

	time.Sleep(400 * time.Millisecond)

	exitChan <- true

	select {
	case <-measureAndReportExitChan:
		t.Log("Loop successfully stopped")
	case <-time.After(5 * time.Second):
		t.Error("Never receive a loop exit message over the channel")
	}

	if !measurerReporterCalled {
		t.Error("Measurer-reporter was never called")
	}
}

func TestMeasureAndReportLocations(t *testing.T) {
	locations := []*uv.Location{
		{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"},
		{DisplayName: "test2", IANA: "Continent/City2", Latitude: "444.444", Longitude: "555.555"},
	}

	measuredLocations := []string{}

	measurerReporter := func(location *uv.Location) error {
		measuredLocations = append(measuredLocations, location.DisplayName)
		return nil
	}

	uv.MeasureAndReportLocations(locations, measurerReporter)

	if measuredLocations[0] != "test" {
		t.Errorf("Expected first measured location to be %s but got %s", "test", measuredLocations[0])
	}
	if measuredLocations[1] != "test2" {
		t.Errorf("Expected second measured location to be %s but got %s", "test2", measuredLocations[1])
	}
}

func TestMeasureAndReportLocations_FailOnFirst(t *testing.T) {
	locations := []*uv.Location{
		{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"},
		{DisplayName: "test2", IANA: "Continent/City2", Latitude: "444.444", Longitude: "555.555"},
	}

	measuredLocations := []string{}

	measurerReporter := func(location *uv.Location) error {
		if location.DisplayName == "test" {
			return errors.New("something happened")
		}
		measuredLocations = append(measuredLocations, location.DisplayName)
		return nil
	}

	uv.MeasureAndReportLocations(locations, measurerReporter)

	if measuredLocations[0] != "test2" {
		t.Errorf("Expected first measured location to be %s but got %s", "test2", measuredLocations[0])
	}
}

func TestMeasureAndReportFunction(t *testing.T) {
	provider := &testMeasurementProvider{MeasurementForLocation: make(map[string]float32)}
	provider.MeasurementForLocation["test"] = 11.3
	reporter := &testMeasurementReporter{ReportedLocations: make(map[string]float32)}
	location := &uv.Location{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"}
	err := uv.GetMeasureAndReportFunction(provider, reporter)(location)
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error: %w", err))
	}
	if provider.MeasuredLocations[0] != location.DisplayName {
		t.Errorf("Expected measured location %s but got %s", location.DisplayName, provider.MeasuredLocations[0])
	}
	if reporter.ReportedLocations[location.DisplayName] != 11.3 {
		t.Errorf("Expected reported location %s to be %.2f but got %.2f ", location.DisplayName, 11.3,
			reporter.ReportedLocations[location.DisplayName])
	}
}

func TestMeasureAndReportFunction_IndexSeverityChanged(t *testing.T) {
	provider := &testMeasurementProvider{MeasurementForLocation: make(map[string]float32)}
	reporter := &testMeasurementReporter{ReportedLocations: make(map[string]float32)}
	location := &uv.Location{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"}

	measurementsToCheck := []float32{2.4, 11.3}
	for i, measurementToCheck := range measurementsToCheck {
		provider.MeasurementForLocation["test"] = measurementToCheck

		err := uv.GetMeasureAndReportFunction(provider, reporter)(location)
		if err != nil {
			t.Error(fmt.Errorf("Unexpected error: %w", err))
		}

		if provider.MeasuredLocations[i] != location.DisplayName {
			t.Errorf("Expected measured location %s but got %s", location.DisplayName, provider.MeasuredLocations[i])
		}
		if reporter.ReportedLocations[location.DisplayName] != measurementToCheck {
			t.Errorf("Expected reported location %s to be %.2f but got %.2f ", location.DisplayName,
				measurementToCheck, reporter.ReportedLocations[location.DisplayName])
		}
	}
}

func TestMeasureAndReportFunction_IndexSeverityMaintained(t *testing.T) {
	provider := &testMeasurementProvider{MeasurementForLocation: make(map[string]float32)}
	reporter := &testMeasurementReporter{ReportedLocations: make(map[string]float32)}
	location := &uv.Location{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"}

	measurementsToCheck := []float32{3.1, 4.2}
	for i, measurementToCheck := range measurementsToCheck {
		provider.MeasurementForLocation["test"] = measurementToCheck

		err := uv.GetMeasureAndReportFunction(provider, reporter)(location)
		if err != nil {
			t.Error(fmt.Errorf("Unexpected error: %w", err))
		}

		if provider.MeasuredLocations[i] != location.DisplayName {
			t.Errorf("Expected measured location %s but got %s", location.DisplayName, provider.MeasuredLocations[i])
		}
		if reporter.ReportedLocations[location.DisplayName] != measurementsToCheck[0] {
			t.Errorf("Expected reported location %s to be %.2f but got %.2f ", location.DisplayName,
				measurementsToCheck, reporter.ReportedLocations[location.DisplayName])
		}
	}
}

func TestMeasureAndReportFunction_FailOnMeasure(t *testing.T) {
	provider := &testMeasurementProvider{FailOnLocation: map[string]bool{"test": true}}
	location := &uv.Location{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"}

	err := uv.GetMeasureAndReportFunction(provider, nil)(location)
	if err == nil {
		t.Error("Expected an error on measurement")
	}
}

func TestMeasureAndReportFunction_FailOnReport(t *testing.T) {
	provider := &testMeasurementProvider{MeasurementForLocation: map[string]float32{"test": 11.3}}
	reporter := &testMeasurementReporter{FailOnLocation: map[string]bool{"test": true}}
	location := &uv.Location{DisplayName: "test", IANA: "Continent/City", Latitude: "222.222", Longitude: "333.333"}

	err := uv.GetMeasureAndReportFunction(provider, reporter)(location)
	if err == nil {
		t.Error("Expected an error on report")
	}

	if provider.MeasuredLocations[0] != location.DisplayName {
		t.Errorf("Expected measured location %s but got %s", location.DisplayName, provider.MeasuredLocations[0])
	}
}

func TestOpenWeatherMap_FailOnRequestExec(t *testing.T) {
	openWeatherMap := &uv.OpenWeatherMap{Host: "!!!", AppID: "abcd"}
	_, measurementError := openWeatherMap.Measure(uv.Locations[0])
	if measurementError == nil {
		t.Error("Expected an error")
	}
}

func TestOpenWeatherMap_FailOnInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{{{"))
	}))
	defer server.Close()
	openWeatherMap := &uv.OpenWeatherMap{Host: server.URL, AppID: "abcd"}
	_, measurementError := openWeatherMap.Measure(uv.Locations[0])
	if measurementError == nil {
		t.Error("Expected an error")
	}
}

func TestOpenWeatherMap_Measure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/2.5/onecall" {
			t.Errorf("Expected URL path %s but got %s", "/data/2.5/onecall", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("lat") != uv.Locations[0].Latitude {
			t.Errorf("Expected lat query param %s but got %s", uv.Locations[0].Latitude, query.Get("lat"))
		}
		if query.Get("lon") != uv.Locations[0].Longitude {
			t.Errorf("Expected lon query param %s but got %s", uv.Locations[0].Longitude, query.Get("lon"))
		}
		if query.Get("appid") != "abcd" {
			t.Errorf("Expected appid query param %s but got %s", "abcd", query.Get("appid"))
		}
		if query.Get("exclude") != "minutely,hourly,alerts,daily" {
			t.Errorf("Expected exclude query param %s but got %s", "minutely,hourly,alerts,daily", query.Get("exclude"))
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(&uv.OneCallResponse{Current: &uv.OneCallCurrent{UVI: 5.32}})
	}))
	defer server.Close()
	openWeatherMap := &uv.OpenWeatherMap{Host: server.URL, AppID: "abcd"}
	uvIndex, measurementError := openWeatherMap.Measure(uv.Locations[0])
	if measurementError != nil {
		t.Error(fmt.Errorf("Unexpected error: %w", measurementError))
	}
	if uvIndex != 5.32 {
		t.Errorf("Expected index %f but got %f", 5.32, uvIndex)
	}
}

func TestUVIndexHasChanged(t *testing.T) {
	if !uv.IndexHasChanged(0, 5) {
		t.Error("UV index should have been classified as 'changed'")
	}

	if !uv.IndexHasChanged(5, 2) {
		t.Error("UV index should have been classified as 'changed'")
	}

	if !uv.IndexHasChanged(2, 5) {
		t.Error("UV index should have been classified as 'changed'")
	}

	if !uv.IndexHasChanged(5, 9) {
		t.Error("UV index should have been classified as 'changed'")
	}

	if uv.IndexHasChanged(5, 3) {
		t.Error("UV index should not have been classified as 'changed'")
	}
}

func TestSTDOutMeasurementReporter(t *testing.T) {
	origStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	stdoutReporter := &uv.STDOutMeasurementReporter{}
	stdoutReporter.Report(uv.Locations[0], 1.0)
	stdoutReporter.Report(uv.Locations[0], 4.0)
	stdoutReporter.Report(uv.Locations[0], 11.0)

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	expectedResult := `The UV index in Tel-Aviv is 1.0. It's safe to go outside! 😎
#uvindex #uvbot #telaviv
The UV Index in Tel-Aviv is 4.0. Seek shade and lather up on that sun screen! 🌞
#uvindex #uvbot #telaviv
Hot dang! The UV Index in Tel-Aviv is 11.0. Stay indoors! 🔥
#uvindex #uvbot #telaviv
`
	got := buf.String()
	if got != expectedResult {
		t.Errorf("Expected %s but got %s", expectedResult, got)
	}
}
