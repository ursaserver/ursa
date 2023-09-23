package ursa

import (
	"errors"
)

type duration int

type rate struct {
	capacity int
	sec      duration
}

const (
	second duration = 1 // Note second is intentionally unexported
	Minute          = second * 60
	Hour            = Minute * 60
	Day             = Hour * 24
)

const MaxRatePerSec = 1 // per second

var ErrMaxRateExceed = errors.New("rate exceed maximum capacity")

// Create a rate of some amount per given time for example, to create a rate of
// 500 request per hour, say Rate(500, ursa.Hour)
//
// The error returned is non-nil when the rate exceeds the maximum supported
// rate. The rate value when error is not nil must be discarded.
func Rate(amount int, time duration) (rate, error) {
	if amount/int(time) > MaxRatePerSec {
		return rate{}, ErrMaxRateExceed
	}
	return rate{amount, time}, nil
}

// Header field to limit the rate by
type RateBy string

const rateByIP = RateBy("IP")

// Return the rate based on configuration that should be used for the a given reqPath.
func rateForPath(r reqPath, conf Conf) *rate {
	// Search linearly through the routes in the configuration to find a
	// pattern that matches reqPath. Note that speed won't be an issue here
	// since this function is supposed to be memoized when using.
	// Memoization should be possible since the configuration is not changed once loaded.
	return new(rate)
}
