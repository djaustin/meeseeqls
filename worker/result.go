package worker

import "time"

// Result represents the result of a query to be output by the application
type Result struct {
	Start       int64         `json:"queryStart"`
	End         int64         `json:"queryEnd"`
	Duration    time.Duration `json:"duration"`
	Source      string        `json:"fromSite"`
	Destination string        `json:"toSite"`
}
