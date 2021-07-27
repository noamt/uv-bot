package uv

import (
	"fmt"
	"time"
)

type Location struct {
	DisplayName string
	IANA        string
	Latitude    string
	Longitude   string
}

var TelAviv = &Location{DisplayName: "Tel-Aviv", IANA: "Asia/Jerusalem", Latitude: "32.109333", Longitude: "34.855499"}

var Locations = []*Location{
	TelAviv,
}

func GetLocation(iana string) (*time.Location, error) {
	location, locationError := time.LoadLocation(iana)
	if locationError != nil {
		return nil, fmt.Errorf("failed to load location %s: %w", iana, locationError)
	}
	return location, nil
}