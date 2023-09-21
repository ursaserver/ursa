package ursa

type seconds int

type rate struct {
	capacity int
	sec      seconds
}

const Minute = seconds(60)
const Hour = Minute * 60
const Day = Hour * 24

// Create a rate of some amount per given time
// for example, to create a rate of 500 request per hour,
// say Rate(500, usra.Hour)
func Rate(amount int, time seconds) rate {
	return rate{amount, time}
}

func (r rate) equal(s rate) bool {
	return r.capacity == s.capacity && r.sec == s.sec
}

// Header field to limit the rate by
type RateBy string

const rateByIP = RateBy("IP")
