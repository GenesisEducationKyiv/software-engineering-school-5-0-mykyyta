package subscription

import "time"

type Frequency string

const (
	FreqHourly Frequency = "hourly"
	FreqDaily  Frequency = "daily"
)

func (f Frequency) Valid() bool { return f == FreqHourly || f == FreqDaily }

type Subscription struct {
	ID             string
	Email          string
	City           string
	Frequency      Frequency
	IsConfirmed    bool
	IsUnsubscribed bool
	Token          string
	CreatedAt      time.Time
}
