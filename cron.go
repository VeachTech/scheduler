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
	"strconv"
	"strings"
	"time"
)

type CronTime struct {
	Sec, Min, Hour, Day, Month, Weekday string
}

// Returns a CronTime object that matches every time
func NewCron() *CronTime {
	return &CronTime{"*", "*", "*", "*", "*", "*"}
}

// Calculates the next time that matches the crontab after the start time
func (CFormat *CronTime) NextRunTime(start time.Time) (time.Time, error) {
	next := start.Add(time.Second)
	next = next.Add(time.Duration(-next.Nanosecond()) * time.Nanosecond) // round down to the nearest second
	// Calculate next Cron Date
	for {
		valid, err := validMatch(CFormat.Month, int(next.Month())-1) // Subtract 1 to make months zero based
		if err != nil {
			return next, err
		}
		if !valid {
			next = next.AddDate(0, 1, -next.Day()+1)                     // Add one month and set the day to the first
			next = next.Add(time.Duration(-next.Hour()) * time.Hour)     // zero the hours
			next = next.Add(time.Duration(-next.Minute()) * time.Minute) // zero the minutes
			next = next.Add(time.Duration(-next.Second()) * time.Second) // zero the seconds
			continue
		}

		dayValid, err := validDay(CFormat.Day, int(next.Day())) // Days have more cron format options
		if err != nil {
			return next, err
		}

		weekdayValid2, err := validWeekDay(CFormat.Weekday, int(next.Weekday()))
		if err != nil {
			return next, err
		}
		if !dayValid || !weekdayValid2 {
			next = next.AddDate(0, 0, 1)
			next = next.Add(time.Duration(-next.Hour()) * time.Hour)     // zero the hours
			next = next.Add(time.Duration(-next.Minute()) * time.Minute) // zero the minutes
			next = next.Add(time.Duration(-next.Second()) * time.Second) // zero the seconds
			continue
		}

		valid, err = validMatch(CFormat.Hour, int(next.Hour()))
		if err != nil {
			return next, err
		}
		if !valid {
			next = next.Add(time.Hour)
			next = next.Add(time.Duration(-next.Minute()) * time.Minute) // zero the minutes
			next = next.Add(time.Duration(-next.Second()) * time.Second) // zero the seconds
			continue
		}

		valid, err = validMatch(CFormat.Min, int(next.Minute()))
		if err != nil {
			return next, err
		}
		if !valid {
			next = next.Add(time.Minute)
			next = next.Add(time.Duration(-next.Second()) * time.Second) // zero the seconds
			continue
		}

		valid, err = validMatch(CFormat.Sec, int(next.Second()))
		if err != nil {
			return next, err
		}
		if !valid {
			next = next.Add(time.Second)
			continue
		}
		break
	}
	return next, nil
}

// returns true if the given "time" is valid for the cron string
// The range of time is zero based so subtract 1 from the month
// or day to make the first time be zero based
// returns an error if given an invalid cron string
func validMatch(cron string, time int) (bool, error) {
	for _, part := range strings.Split(cron, ",") {
		switch {
		case part == "*":
			return true, nil // matches all times so no need to check

		case strings.HasPrefix(part, "*/"):
			num, err := strconv.Atoi(strings.TrimLeft(part, "*/"))
			if err != nil {
				return false, errors.New("Cron syntax error */ must be followed by a number")
			}
			if time%num == 0 {
				return true, nil
			}

		case strings.Contains(part, "-"):
			AB := strings.SplitN(part, "-", 2)
			a, err := strconv.Atoi(AB[0])
			if err != nil {
				return false, errors.New("Expected number to the left of \"-\"")
			}
			b, err := strconv.Atoi(AB[1])
			if err != nil {
				return false, errors.New("Expected number to the right of \"-\"")
			}
			if a >= b {
				return false, errors.New("When using A-B, A must be less than B")
			}
			if time >= a && time <= b {
				return true, nil
			}
		}
	}
	return false, nil
}

func validDay(cron string, time int) (bool, error) {
	return true, nil
}

func validWeekDay(cron string, time int) (bool, error) {
	return true, nil
}

// daysIn returns the number of days in a month for a given year. 
func daysIn(m time.Month, year int) int {
	// This is equivalent to time.daysIn(m, year). 
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
