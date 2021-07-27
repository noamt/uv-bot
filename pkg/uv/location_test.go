package uv_test

import (
	"fmt"
	"testing"

	"github.com/noamt/uv-bot/pkg/uv"
)

func TestLocations(t *testing.T) {
	firstLocation := uv.Locations[0]
	if firstLocation.DisplayName != "Tel-Aviv" {
		t.Errorf("Expected display name to be %s but got %s", "Tel-Aviv", firstLocation.DisplayName)
	}
	if firstLocation.IANA != "Asia/Jerusalem" {
		t.Errorf("Expected IANA to be %s but got %s", "Asia/Jerusalem", firstLocation.IANA)
	}
	if firstLocation.Latitude != "32.109333" {
		t.Errorf("Expected Latitude to be %s but got %s", "32.109333", firstLocation.Latitude)
	}
	if firstLocation.Longitude != "34.855499" {
		t.Errorf("Expected Latitude to be %s but got %s", "34.855499", firstLocation.Longitude)
	}
}

func TestGetLocation_Valid(t *testing.T) {
	location, err := uv.GetLocation("Asia/Jerusalem")
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error: %w", err))
	}
	if location.String() != "Asia/Jerusalem" {
		t.Error("Expected to find a location for Asia/Jerusalem")
	}
}

func TestGetLocationOrFail_Invalid(t *testing.T) {
	location, err := uv.GetLocation("haha")
	if err == nil {
		t.Errorf("Expected an error")
	}
	if location != nil {
		t.Errorf("Unexpected location")
	}
	if err.Error() != "failed to load location haha: unknown time zone haha" {
		t.Errorf("Expected error message %s but got %s", "failed to load location haha: unknown time zone haha", err.Error())
	}
}
