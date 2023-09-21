package ursa

type Conf struct {
	routes       []Route
	defaultRates map[RateBy]rate
	nonRateLimit []NonRateLimitRoute
}

type Route struct {
	pattern   string // regex describing HTTP path to match
	rate      map[RateBy]rate
	forwardTo string // the address of the server to forward requests to
}

type NonRateLimitRoute struct {
	pattern   string // regex describing HTTP path to match
	forwardTo string // the address of the server to forward requests to
}
