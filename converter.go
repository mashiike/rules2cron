package rules2cron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Converter struct {
	ReferenceDate time.Time
	TimeZone      *time.Location
}

func (c *Converter) Convert(scheduleExpression string) (string, error) {
	if c.TimeZone == nil {
		c.TimeZone = time.Local
	}
	switch {
	case strings.HasPrefix(scheduleExpression, "rate("):
		return c.convertRate(scheduleExpression)
	case strings.HasPrefix(scheduleExpression, "cron("):
		return c.convertCron(scheduleExpression)
	default:
		return "", errors.New("invalid format")
	}
}

func (c *Converter) convertRate(scheduleExpression string) (string, error) {
	parts := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(scheduleExpression, "rate("), ")"))
	if len(parts) != 2 {
		return "", errors.New("invalid format: require rate(Value Unit) ")
	}
	value, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid format: parse Value: %w", err)
	}
	if value == 0 {
		return "", errors.New("invalid format: require Value over 0")
	}
	unit := parts[1]
	if value == 1 {
		switch unit {
		case "minute":
			s := &Schedule{
				Minute:     "*",
				Hour:       "*",
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "*",
			}
			return s.String(), nil
		case "hour":
			s := &Schedule{
				Minute:     "0",
				Hour:       "*",
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "*",
			}
			return s.String(), nil
		case "day":
			s := &Schedule{
				Minute:     "0",
				Hour:       fmt.Sprintf("%d", convertTimeZone(0, c.ReferenceDate.Location(), c.TimeZone)),
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "*",
			}
			return s.String(), nil
		case "minutes":
			return "", errors.New("invalid format: can not use pluralistic")
		case "hours":
			return "", errors.New("invalid format: can not use pluralistic")
		case "days":
			return "", errors.New("invalid format: can not use pluralistic")
		default:
			return "", fmt.Errorf("invalid format: unknown unit: %s", unit)
		}
	}
	switch unit {
	case "minute":
		return "", errors.New("invalid format: can not use singular form")
	case "hour":
		return "", errors.New("invalid format: can not use singular form")
	case "day":
		return "", errors.New("invalid format: can not use singular form")
	case "minutes":
		s := &Schedule{
			Minute:     fmt.Sprintf("*/%d", value),
			Hour:       "*",
			DayOfMonth: "*",
			Month:      "*",
			DayOfWeek:  "*",
		}
		return s.String(), nil
	case "hours":
		s := &Schedule{
			Minute:     "0",
			Hour:       fmt.Sprintf("*/%d", value),
			DayOfMonth: "*",
			Month:      "*",
			DayOfWeek:  "*",
		}
		return s.String(), nil
	case "days":
		s := &Schedule{
			Minute:     "0",
			Hour:       fmt.Sprintf("%d", convertTimeZone(0, c.ReferenceDate.Location(), c.TimeZone)),
			DayOfMonth: fmt.Sprintf("*/%d", value),
			Month:      "*",
			DayOfWeek:  "*",
		}
		return s.String(), nil
	default:
		return "", fmt.Errorf("invalid format: unknown unit: %s", unit)
	}
}

func (c *Converter) convertCron(scheduleExpression string) (string, error) {
	parts := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(scheduleExpression, "cron("), ")"))
	if len(parts) != 6 {
		return "", errors.New("invalid format: require cron(Minutes Hours Day-of-month Month Day-of-week Year) ")
	}
	minute := parts[0]
	if minute == "?" {
		minute = "*"
	}
	hour, err := convertCronHourPartTimeZone(parts[1], c.ReferenceDate.Location(), c.TimeZone)
	if err != nil {
		return "", err
	}
	dayOfMonth := parts[2]
	if dayOfMonth == "?" {
		dayOfMonth = "*"
	}
	if strings.ContainsRune(dayOfMonth, 'W') {
		value, err := strconv.ParseUint(strings.TrimRight(dayOfMonth, "W"), 10, 32)
		if err != nil {
			return "", fmt.Errorf("invalid format: parse day of month: %w", err)
		}
		date := time.Date(c.ReferenceDate.Year(), c.ReferenceDate.Month(), int(value), 0, 0, 0, 0, c.ReferenceDate.Location())
		for date.Weekday() == time.Sunday || date.Weekday() == time.Saturday {
			date = date.AddDate(0, 0, -1)
		}
		dayOfMonth = fmt.Sprintf("%d", date.Day())
	}
	if dayOfMonth == "L" {
		dayOfMonth = fmt.Sprintf("%d", c.ReferenceDate.AddDate(0, 1, 0).AddDate(0, 0, -1).Day())
	}
	month := parts[3]
	if month == "?" {
		month = "*"
	}
	month = convertCronMonthPart(month)
	dayOfWeek := parts[4]
	if dayOfWeek == "?" {
		dayOfWeek = "*"
	}
	if strings.ContainsRune(dayOfWeek, '#') {
		p := strings.SplitAfterN(dayOfWeek, "#", 2)
		if len(p) != 2 {
			return "", fmt.Errorf("invalid format: parse day of week syntax: %s", dayOfWeek)
		}
		value, err := strconv.ParseUint(strings.TrimRight(p[0], "#"), 10, 32)
		if err != nil {
			return "", fmt.Errorf("invalid format: parse day of week value: %w", err)
		}
		weekday := time.Weekday(value - 1)
		count, err := strconv.ParseUint(p[1], 10, 32)
		if err != nil {
			return "", fmt.Errorf("invalid format: parse day of week count: %w", err)
		}
		date := time.Date(c.ReferenceDate.Year(), c.ReferenceDate.Month(), 1, 0, 0, 0, 0, c.ReferenceDate.Location())
		i := uint64(0)
		for date.Month() == c.ReferenceDate.Month() {
			if date.Weekday() != weekday {
				date = date.AddDate(0, 0, 1)
				continue
			}
			i++
			if i == count {
				break
			}
			date = date.AddDate(0, 0, 7)
		}
		dayOfMonth = fmt.Sprintf("%d", date.Day())
		dayOfWeek = "*"
	}
	if strings.ContainsRune(dayOfWeek, 'L') {
		value, err := strconv.ParseUint(strings.TrimRight(dayOfWeek, "L"), 10, 32)
		if err != nil {
			return "", fmt.Errorf("invalid format: parse day of week: %w", err)
		}
		weekday := time.Weekday(value - 1)
		date := c.ReferenceDate.AddDate(0, 1, 0).AddDate(0, 0, -1)
		for date.Weekday() != weekday {
			date = date.AddDate(0, 0, -1)
		}
		dayOfMonth = fmt.Sprintf("%d", date.Day())
		dayOfWeek = "*"
	}
	dayOfWeek = convertCronDayOfWeekPart(dayOfWeek)

	year := parts[5]
	isTarget, err := referenceDateIsTargetYear(c.ReferenceDate, year)
	if err != nil {
		return "", err
	}
	if !isTarget {
		return "", fmt.Errorf("cannot be converted because the reference date is not the target year: %s", year)
	}
	s := &Schedule{
		Minute:     minute,
		Hour:       hour,
		DayOfMonth: dayOfMonth,
		Month:      month,
		DayOfWeek:  dayOfWeek,
	}
	return s.String(), nil
}

func convertTimeZone(value uint64, base *time.Location, to *time.Location) uint64 {
	return uint64(time.Date(2022, 06, 01, int(value), 0, 0, 0, base).In(to).Hour())
}

func convertCronDayOfWeekPart(dayOfWeek string) string {
	dayOfWeek = strings.ToUpper(dayOfWeek)
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "1", "0")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "2", "1")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "3", "2")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "4", "3")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "5", "4")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "6", "5")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "7", "6")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "MON", "1")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "TUE", "2")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "WED", "3")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "THU", "4")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "FRI", "5")
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "SAT", "6")
	if strings.ContainsRune(dayOfWeek, '-') {
		parts := strings.SplitN(dayOfWeek, "-", 2)
		if len(parts) == 2 {
			parts[1] = strings.ReplaceAll(parts[1], "SUN", "7")
		}
		dayOfWeek = strings.Join(parts, "-")
	}
	dayOfWeek = strings.ReplaceAll(dayOfWeek, "SUN", "0")
	return dayOfWeek
}

func convertCronMonthPart(month string) string {
	month = strings.ToUpper(month)
	month = strings.ReplaceAll(month, "JAN", "1")
	month = strings.ReplaceAll(month, "FEB", "2")
	month = strings.ReplaceAll(month, "MAR", "3")
	month = strings.ReplaceAll(month, "APR", "4")
	month = strings.ReplaceAll(month, "MAY", "5")
	month = strings.ReplaceAll(month, "JUN", "6")
	month = strings.ReplaceAll(month, "JUL", "7")
	month = strings.ReplaceAll(month, "AUG", "8")
	month = strings.ReplaceAll(month, "SEP", "9")
	month = strings.ReplaceAll(month, "OCT", "10")
	month = strings.ReplaceAll(month, "NOV", "11")
	month = strings.ReplaceAll(month, "DEC", "12")
	return month
}

func convertCronHourPartTimeZone(hour string, base *time.Location, to *time.Location) (string, error) {
	var hourRate string
	if strings.ContainsRune(hour, '/') {
		p := strings.SplitAfterN(hour, "/", 2)
		if len(p) != 2 {
			return "", fmt.Errorf("invalid format: parse hour syntax: %s", hour)
		}
		hour = strings.TrimRight(p[0], "/")
		hourRate = "/" + p[1]
	}
	hours := strings.Split(hour, ",")
	afterHours := make([]string, 0, len(hours))
	for _, hour := range hours {
		switch {
		case hour == "*" || hour == "?":
			afterHours = append(afterHours, "*")
		case strings.ContainsRune(hour, '-'):
			p := strings.SplitAfterN(hour, "-", 2)
			if len(p) != 2 {
				return "", fmt.Errorf("invalid format: parse hour syntax: %s", hour)
			}
			start, err := strconv.ParseUint(strings.TrimRight(p[0], "-"), 10, 32)
			if err != nil {
				return "", fmt.Errorf("invalid format: parse hour start: %w", err)
			}
			start = convertTimeZone(start, base, to)
			end, err := strconv.ParseUint(p[1], 10, 32)
			if err != nil {
				return "", fmt.Errorf("invalid format: parse hour start: %w", err)
			}
			end = convertTimeZone(end, base, to)
			afterHours = append(afterHours, fmt.Sprintf("%d-%d", start, end))
		default:
			value, err := strconv.ParseUint(hour, 10, 32)
			if err != nil {
				return "", fmt.Errorf("invalid format: parse hour: %w", err)
			}
			afterHours = append(afterHours, fmt.Sprintf("%d", convertTimeZone(value, base, to)))
		}
	}
	return strings.Join(afterHours, ",") + hourRate, nil
}

func referenceDateIsTargetYear(referenceDate time.Time, year string) (bool, error) {
	var yearRate *uint64
	if strings.ContainsRune(year, '/') {
		p := strings.SplitAfterN(year, "/", 2)
		if len(p) != 2 {
			return false, fmt.Errorf("invalid format: parse year syntax: %s", year)
		}
		year = strings.TrimRight(p[0], "/")
		value, err := strconv.ParseUint(p[1], 10, 32)
		if err != nil {
			return false, fmt.Errorf("invalid format: parse year rate: %w", err)
		}
		yearRate = &value
	}
	years := strings.Split(year, ",")
	refYear := uint64(referenceDate.Year())
	for _, y := range years {
		switch {
		case y == "*":
			if yearRate == nil {
				return true, nil
			}
			if refYear%*yearRate == 0 {
				return true, nil
			}
		case strings.ContainsRune(y, '-'):
			p := strings.SplitAfterN(y, "-", 2)
			if len(p) != 2 {
				return false, fmt.Errorf("invalid format: parse year syntax: %s", year)
			}
			start, err := strconv.ParseUint(strings.TrimRight(p[0], "-"), 10, 32)
			if err != nil {
				return false, fmt.Errorf("invalid format: parse year start: %w", err)
			}
			end, err := strconv.ParseUint(p[1], 10, 32)
			if err != nil {
				return false, fmt.Errorf("invalid format: parse year start: %w", err)
			}
			if refYear < start {
				break
			}
			if refYear > end {
				break
			}
			if yearRate == nil {
				return true, nil
			}
			if refYear%*yearRate == 0 {
				return true, nil
			}
		default:
			value, err := strconv.ParseUint(y, 10, 32)
			if err != nil {
				return false, fmt.Errorf("invalid format: parse year: %w", err)
			}
			if value == refYear {
				return true, nil
			}
		}
	}
	return false, nil
}
