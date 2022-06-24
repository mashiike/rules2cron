package rules2cron

import (
	"fmt"
)

type Schedule struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
}

func (s *Schedule) String() string {
	return fmt.Sprintf("%s %s %s %s %s", s.Minute, s.Hour, s.DayOfMonth, s.Month, s.DayOfWeek)
}
