package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRex *regexp.Regexp

func init() {
	localRex, err := regexp.Compile("^[0-9]+(m|s|ms)$")
	if err != nil {
		panic(err)
	}
	durationRex = localRex
}

func parseDuration(durationToParse string) (time.Duration, error) {

	if !durationRex.Match([]byte(durationToParse)) {
		return 0, fmt.Errorf("value %s is not a duration", durationToParse)
	}

	var unit time.Duration = time.Minute
	var justDigits string
	if strings.HasSuffix(durationToParse, "ms") {
		unit = time.Millisecond
		justDigits = strings.TrimRight(durationToParse, "ms")
	} else if strings.HasSuffix(durationToParse, "s") {
		unit = time.Second
		justDigits = strings.TrimRight(durationToParse, "s")
	} else {
		unit = time.Minute
		justDigits = strings.TrimRight(durationToParse, "m")
	}

	digits, err := strconv.Atoi(justDigits)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %s as integer", justDigits)
	}

	return time.Duration(digits) * unit, nil
}
