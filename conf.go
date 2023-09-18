package ursa

type Conf struct {
	routes       []Route
	defaultRates map[RateBy]rate
	nonRateLimit []string // slice of regex describing HTTP paths to avoid rate limiting
}

type Route struct {
	pattern   string // regex describing HTTP path to match
	rate      map[RateBy]rate
	forwardTo string // the address of the server to forward requests to
}
