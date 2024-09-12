package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRex *regexp.Regexp
var unixTimestampRex *regexp.Regexp
var relativeTimestampRex *regexp.Regexp

func init() {
	localRex, err := regexp.Compile("^[0-9]+(d|h|m|s|ms)$")
	if err != nil {
		panic(err)
	}
	durationRex = localRex

	localUnixTimestampRex, err := regexp.Compile("^[0-9]+$")
	if err != nil {
		panic(err)
	}
	unixTimestampRex = localUnixTimestampRex

	localRelativeTimestampRex, err := regexp.Compile("^now(-[0-9]+d|-[0-9]+h|-[0-9]+m|-[0-9]+s|-[0-9]+ms)?$")
	if err != nil {
		panic(err)
	}
	relativeTimestampRex = localRelativeTimestampRex
}

func ParseDuration(durationToParse string) (time.Duration, error) {

	durationToParse = strings.Trim(durationToParse, " ")

	if !durationRex.MatchString(durationToParse) {
		return 0, fmt.Errorf("value %s is not a duration", durationToParse)
	}

	var unit time.Duration
	var justDigits string
	if strings.HasSuffix(durationToParse, "ms") {
		unit = time.Millisecond
		justDigits = strings.TrimRight(durationToParse, "ms")
	} else if strings.HasSuffix(durationToParse, "s") {
		unit = time.Second
		justDigits = strings.TrimRight(durationToParse, "s")
	} else if strings.HasSuffix(durationToParse, "m") {
		unit = time.Minute
		justDigits = strings.TrimRight(durationToParse, "m")
	} else if strings.HasSuffix(durationToParse, "h") {
		unit = time.Hour
		justDigits = strings.TrimRight(durationToParse, "h")
	} else {
		unit = time.Hour * 24
		justDigits = strings.TrimRight(durationToParse, "d")
	}

	digits, err := strconv.Atoi(justDigits)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %s as integer", justDigits)
	}

	return time.Duration(digits) * unit, nil
}

func ParseDate(dateToParse string, now time.Time) (time.Time, error) {

	dateToParse = strings.Trim(dateToParse, " ")

	dt, err := time.Parse(time.RFC3339, dateToParse)
	success := err == nil
	if success {
		return dt, nil
	}

	dt, err = tryParseRelativeTimestamp(dateToParse, now)
	success = err == nil
	if success {
		return dt, nil
	}

	dt, err = tryParseUnixTimestamp(dateToParse, now)
	success = err == nil
	if success {
		return dt, nil
	}

	return now, fmt.Errorf("unable to parse time: %w", err)
}

func tryParseUnixTimestamp(dateToParse string, now time.Time) (time.Time, error) {
	if !unixTimestampRex.MatchString(dateToParse) {
		return now, fmt.Errorf("value %s is not a unix timestamp", dateToParse)
	}

	unixTime, err := strconv.ParseInt(dateToParse, 10, 64)
	if err != nil {
		return now, err
	}

	dt := time.Unix(unixTime, 0)
	return dt, nil
}

func tryParseRelativeTimestamp(dateToParse string, now time.Time) (time.Time, error) {
	if !relativeTimestampRex.MatchString(dateToParse) {
		return now, fmt.Errorf("value %s is not a relative date", dateToParse)
	}

	if dateToParse == "now" { //TODO: remover magic string
		return now, nil
	}

	parts := strings.Split(dateToParse, "-") //TODO: extract magic string
	if len(parts) <= 1 {
		return now, fmt.Errorf("value %s is not a relative date", dateToParse)
	}

	duration := parts[1]
	parsedDuration, err := ParseDuration(duration)
	if err != nil {
		return now, fmt.Errorf("error parsing period on date: %w", err)
	}

	result := now.Add(-parsedDuration)

	return result, nil
}
