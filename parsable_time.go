package main

import (
	"errors"
	"log"
	"strings"
	"time"
)

// https://groups.google.com/forum/#!topic/golang-dev/I1dGXiwhJaw
// http://golang.org/pkg/encoding/json/#Unmarshaler

const timeFormat = "2006-01-02 15:04"

type ParsableTime struct {
	time.Time
}

func (new_time *ParsableTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	strs := strings.Split(str, "\"")
	if len(strs) != 3 {
		return errors.New("Date must be in double quotes")
	}
	parsed, err := time.Parse(timeFormat, strs[1])
	if nil != err {
		log.Print("ParsableTime failed to parse string: " + strs[1])
		return err
	}
	parsed_time := ParsableTime{parsed}
	new_time = &parsed_time

	return nil
}

func (new_time ParsableTime) MarshalJSON() ([]byte, error) {
	bytes := []byte(`"` + new_time.Format(timeFormat) + `"`)
	return bytes, nil
}
