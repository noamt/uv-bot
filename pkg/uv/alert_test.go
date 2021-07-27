package uv_test

import (
	"reflect"
	"testing"

	"github.com/noamt/uv-bot/pkg/uv"
)

func TestAltertsByLocation(t *testing.T) {
	alerts := uv.AltertsByLocation[uv.TelAviv.DisplayName]
	if reflect.TypeOf(alerts).Kind() != reflect.TypeOf(uv.TelAvivAlerts{}).Kind() {
		t.Error("Alerts object is of an unexpected type")
	}
}

func TestTelAvivAlerts(t *testing.T) {
	telAvivAlerts := &uv.TelAvivAlerts{}

	expectedLow := "The UV index in Tel-Aviv is 1.1. It's safe to go outside! 😎\n#uvindex #uvbot #telaviv"
	if telAvivAlerts.Low(1.1) != expectedLow {
		t.Errorf("Expected %s but got %s", expectedLow, telAvivAlerts.Low(1.1))
	}

	expectedModerate := "The UV Index in Tel-Aviv is 2.1. Seek shade and lather up on that sun screen! 🌞\n#uvindex #uvbot #telaviv"
	if telAvivAlerts.Moderate(2.1) != expectedModerate {
		t.Errorf("Expected %s but got %s", expectedModerate, telAvivAlerts.Moderate(2.1))
	}

	expectedHigh := "Hot dang! The UV Index in Tel-Aviv is 3.2. Stay indoors! 🔥\n#uvindex #uvbot #telaviv"
	if telAvivAlerts.High(3.21) != expectedHigh {
		t.Errorf("Expected %s but got %s", expectedHigh, telAvivAlerts.High(3.21))
	}
}
