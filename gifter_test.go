package ursa

import (
	"testing"
	"time"
)

func TestTickOnceEvery(t *testing.T) {
	type test struct {
		r Rate
		t time.Duration
	}
	tests := []test{
		{r: Rate{60, Minute}, t: time.Minute},
		{r: Rate{30, Minute}, t: time.Minute},
		{r: Rate{60, Hour}, t: time.Hour},
		{r: Rate{30, Hour}, t: time.Hour},
		{r: Rate{30, Day}, t: time.Hour * 24},
	}
	for _, test := range tests {
		expected := test.t
		got := tickOnceEvery(test.r)
		if expected != got {
			t.Errorf("expected tick interval %v got %v for rate %vreqs/%vsecs", expected, got, test.r.Capacity, test.r.RefillDurationInSec)
		}
	}
}

func TestSecondsBeforeSuccess(t *testing.T) {
	type test struct {
		r                         Rate
		currentTime               time.Time
		lastGiftedTime            time.Time
		tokens                    int
		expectedSecondsForSuccess int
	}
	timeInPast := time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
	tests := []test{
		{
			r:                         Rate{60, Minute},
			lastGiftedTime:            timeInPast,
			currentTime:               timeInPast.Add(time.Second * 5),
			tokens:                    1,
			expectedSecondsForSuccess: 0,
		},
		{
			r:                         Rate{60, Minute},
			lastGiftedTime:            timeInPast,
			currentTime:               timeInPast.Add(time.Second * 5),
			tokens:                    0,
			expectedSecondsForSuccess: 55,
		},
		{
			r:                         Rate{60, Minute},
			lastGiftedTime:            timeInPast,
			currentTime:               timeInPast.Add(time.Second * 5),
			tokens:                    -10,
			expectedSecondsForSuccess: 55,
		},
		{
			r:                         Rate{60, Minute},
			lastGiftedTime:            timeInPast,
			currentTime:               timeInPast.Add(time.Second * 5),
			tokens:                    -110,
			expectedSecondsForSuccess: 55 + 60,
		},
		{
			r:                         Rate{60, Minute},
			lastGiftedTime:            timeInPast,
			currentTime:               timeInPast.Add(time.Second * 5),
			tokens:                    -200,
			expectedSecondsForSuccess: 55 + 60*3,
		},
	}
	for _, test := range tests {
		expected := test.expectedSecondsForSuccess
		got := secondsBeforeSuccess(test.currentTime, test.lastGiftedTime, &test.r, test.tokens)
		if expected != got {
			t.Errorf("expected  waiting time %v got %v. current tokens: %v", expected, got, test.tokens)
		}
	}
}
