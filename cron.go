/**
Copyright (c) 2013, Ryan Veach
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of the <organization> nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
**/

package scheduler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// regular expressions for parsing cron syntax
var reNumber = regexp.MustCompile(`^(\d+)$`) // matches numbers starting at the beginning of string
var reNumberRange = regexp.MustCompile(`^(\d+)-(\d+)$`)
var reInterval = regexp.MustCompile(`^\*/(\d+)$`)
var reIntervalInRange = regexp.MustCompile(`^(\d+)-(\d+)/(\d+)$`)

// Structure to hold the crontab matching strings.
// Each string should NOT contain spaces, if you want multiple matches
// use a comma (,). A space will cause all checks to fail.
type CronTime struct {
	Sec, Min, Hour, Day, Month, Weekday string
}

// Returns a CronTime object that matches every time
func NewCron(cronString string) (CronTime, error) {
	fields := strings.Fields(cronString)
	if len(fields) != 6 {
		return CronTime{}, errors.New("Cron String must have 6 fields")
	}
	return CronTime{fields[0], fields[1], fields[2], cronDayHelper(fields[3]), cronMonthHelper(fields[4]), cronDOWHelper(fields[5])}, nil
}

func cronDayHelper(cronString string) string {
	var output = make([]string, 0, 5)
	for _, field := range strings.Split(cronString, ",") {
		if reInterval.MatchString(field) {
			regularExpressionMatches := reInterval.FindStringSubmatch(field)
			modulos, _ := strconv.Atoi(regularExpressionMatches[1])
			field = fmt.Sprintf("1-31/%d", modulos)
		}
		output = append(output, field)
	}
	return strings.Join(output, ",")
}

func cronMonthHelper(cronString string) string {
	var replacer = strings.NewReplacer("JAN", "1", "FEB", "2", "MAR", "3", "APR", "4", "MAY", "5", "JUN", "6",
		"JUL", "7", "AUG", "8", "SEP", "9", "OCT", "10", "NOV", "11", "DEC", "12")
	var output = make([]string, 0, 5)
	for _, field := range strings.Split(cronString, ",") {
		if reInterval.MatchString(field) {
			regularExpressionMatches := reInterval.FindStringSubmatch(field)
			modulos, _ := strconv.Atoi(regularExpressionMatches[1])
			field = fmt.Sprintf("1-12/%d", modulos)
		}
		field = replacer.Replace(field)
		output = append(output, field)
	}
	return strings.Join(output, ",")
}

func cronDOWHelper(cronString string) string {
	var replacer = strings.NewReplacer("SUN", "0", "MON", "1", "TUE", "2", "WED", "3", "THU", "4", "FRI", "5", "SAT", "6")
	var output = make([]string, 0, 5)
	for _, field := range strings.Split(cronString, ",") {
		field = replacer.Replace(field)
		output = append(output, field)
	}
	return strings.Join(output, ",")
}

// Calculates the next time that matches the crontab after the start time
func (CFormat CronTime) NextRunTime(start time.Time) time.Time {
	nextTime := start
	// Calculate next Cron Date
	for {
		if !timeMatchesCron(CFormat.Month, int(nextTime.Month())) {
			nextTime = startOfNextMonth(nextTime)
			continue
		}
		if !timeMatchesCron(CFormat.Day, int(nextTime.Day())) || !timeMatchesCron(CFormat.Weekday, int(nextTime.Weekday())) {
			nextTime = startOfNextDay(nextTime)
			continue
		}
		if !timeMatchesCron(CFormat.Hour, int(nextTime.Hour())) {
			nextTime = startOfNextHour(nextTime)
			continue
		}
		if !timeMatchesCron(CFormat.Min, int(nextTime.Minute())) {
			nextTime = startOfNextMinute(nextTime)
			continue
		}
		if !timeMatchesCron(CFormat.Sec, int(nextTime.Second())) {
			nextTime = nextTime.Add(time.Second)
			continue
		}
		break
	}
	return nextTime
}

func startOfNextMonth(original time.Time) time.Time {
	next := original.AddDate(0, 1, -original.Day()+1) // the first of the next month
	clockTimeToRemove := time.Duration(original.Hour()) * time.Hour
	clockTimeToRemove += time.Duration(original.Minute()) * time.Minute
	clockTimeToRemove += time.Duration(original.Second()) * time.Second
	return next.Add(-clockTimeToRemove)
}

func startOfNextDay(original time.Time) time.Time {
	next := original.AddDate(0, 0, 1)
	clockTimeToRemove := time.Duration(original.Hour()) * time.Hour
	clockTimeToRemove += time.Duration(original.Minute()) * time.Minute
	clockTimeToRemove += time.Duration(original.Second()) * time.Second
	return next.Add(-clockTimeToRemove)
}

func startOfNextHour(original time.Time) time.Time {
	next := original.Add(time.Hour)
	MinuteSecondsToRemove := time.Duration(original.Minute()) * time.Minute
	MinuteSecondsToRemove += time.Duration(original.Second()) * time.Second
	return next.Add(-MinuteSecondsToRemove)
}

func startOfNextMinute(original time.Time) time.Time {
	next := original.Add(time.Minute)
	secondsToRemove := time.Duration(original.Second()) * time.Second
	return next.Add(-secondsToRemove)
}

// returns true if the given "time" is valid for the cron string
// The range of time is zero based so this works for Hours, Minutes
// or Seconds
func timeMatchesCron(cron string, time int) bool {
	for _, matcher := range strings.Split(cron, ",") {
		switch {
		case matcher == "*":
			fallthrough
		case reNumberRange.MatchString(matcher) && numberRangeIsMatch(matcher, time):
			fallthrough
		case reNumber.MatchString(matcher) && numberIsMatch(matcher, time):
			fallthrough
		case reInterval.MatchString(matcher) && intervalIsMatch(matcher, time):
			fallthrough
		case reIntervalInRange.MatchString(matcher) && intervalIsMatch(matcher, time):
			return true
		}
	}
	return false
}

func numberRangeIsMatch(matcher string, number int) bool {
	regularExpressionMatches := reNumberRange.FindStringSubmatch(matcher)
	beginRange, _ := strconv.Atoi(regularExpressionMatches[1])
	endRange, _ := strconv.Atoi(regularExpressionMatches[2])
	if number >= beginRange && number <= endRange {
		return true
	}
	return false
}

func numberIsMatch(matcher string, number int) bool {
	regularExpressionMatches := reNumber.FindStringSubmatch(matcher)
	matcherNumber, _ := strconv.Atoi(regularExpressionMatches[1])
	if matcherNumber == number {
		return true
	}
	return false
}

func intervalIsMatch(matcher string, number int) bool {
	regularExpressionMatches := reInterval.FindStringSubmatch(matcher)
	modulos, _ := strconv.Atoi(regularExpressionMatches[1])
	if number%modulos == 0 {
		return true
	}
	return false
}

func rangeIntervalIsMatch(matcher string, number int) bool {
	regularExpressionMatches := reIntervalInRange.FindStringSubmatch(matcher)
	beginRange, _ := strconv.Atoi(regularExpressionMatches[1])
	endRange, _ := strconv.Atoi(regularExpressionMatches[2])
	modulos, _ := strconv.Atoi(regularExpressionMatches[3])
	if number >= beginRange && number <= endRange && (number-beginRange)%modulos == 0 {
		return true
	}
	return false
}
