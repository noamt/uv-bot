package uv

import (
	"fmt"
	"time"
)

var AltertsByLocation = map[string]Alerts{
	TelAviv.DisplayName: TelAvivAlerts{},
}

type Alerts interface {
	Low(uvIndex float32) string
	Moderate(uvIndex float32) string
	High(uvIndex float32) string
}

type TelAvivAlerts struct{}

func (TelAvivAlerts) Low(uvIndex float32) string {
	return fmt.Sprintf("The UV index in Tel-Aviv is %.1f. It's safe to go outside! 😎\n#uvindex #telaviv #uvbot_%d", uvIndex, time.Now().Unix())
}

func (TelAvivAlerts) Moderate(uvIndex float32) string {
	return fmt.Sprintf("The UV Index in Tel-Aviv is %.1f. Seek shade and lather up on that sun screen! 🌞\n#uvindex #telaviv #uvbot_%d", uvIndex, time.Now().Unix())
}

func (TelAvivAlerts) High(uvIndex float32) string {
	return fmt.Sprintf("Hot dang! The UV Index in Tel-Aviv is %.1f. Stay indoors! 🔥\n#uvindex #telaviv #uvbot_%d", uvIndex, time.Now().Unix())
}
