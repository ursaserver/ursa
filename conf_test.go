package ursa

import (
	"net/url"
	"regexp"
	"testing"
)

func NilConf() Conf {
	var conf Conf
	return conf
}

func NilUpstream() Conf {
	conf := Conf{
		Routes: []Route{},
	}
	return conf
}

func NilRoutes() Conf {
	conf := Conf{
		Upstream: upstream(),
	}
	return conf
}

func ZeroRoutes() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes:   []Route{},
	}
	return conf
}

func NilRouteRate() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{{
			Methods: []string{"GET"},
			Pattern: regexp.MustCompile("/about"),
		}},
	}
	return conf
}

func NilRoutePattern() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Methods: []string{"GET"},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
		},
	}
	return conf
}

func NilMethodsRoute() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{{
			Pattern: regexp.MustCompile("/about"),
			Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
		}},
	}
	return conf
}

func ZeroMethodsRoute() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{{
			Methods: []string{},
			Pattern: regexp.MustCompile("/about"),
			Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
		}},
	}
	return conf
}

func ZeroRouteRates() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{{
			Methods: []string{"GET"},
			Pattern: regexp.MustCompile("/about"),
			Rates:   RouteRates{},
		}},
	}
	return conf
}

func ValidConfSingleRoute() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{{
			Methods: []string{"GET"},
			Pattern: regexp.MustCompile("/about"),
			Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
		}},
	}
	return conf
}

func InvalidConfBySecondRoutePattern() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Methods: []string{"GET"},
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
			{
				Methods: []string{"GET"},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
		},
	}
	return conf
}

func InvalidConfBySecondRouteRates() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Pattern: regexp.MustCompile("/about"),
				Methods: []string{"GET"},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
			{
				Pattern: regexp.MustCompile("/about"),
				Methods: []string{"GET"},
			},
		},
	}
	return conf
}

func InvalidConfBySecondRouteMethods() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Pattern: regexp.MustCompile("/about"),
				Methods: []string{"GET"},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
			{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
		},
	}
	return conf
}

func InvalidConfBySecondRouteEmptyMethods() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Pattern: regexp.MustCompile("/about"),
				Methods: []string{"GET"},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
			{
				Pattern: regexp.MustCompile("/about"),
				Methods: []string{},
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
			},
		},
	}
	return conf
}

func ValidConfMultipleRoutes() Conf {
	conf := Conf{
		Upstream: upstream(),
		Routes: []Route{
			{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
				Methods: []string{"GET"},
			},
			{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByIP: NewRate(60, Hour)},
				Methods: []string{"POST"},
			},
		},
	}
	return conf
}

func upstream() *url.URL {
	u, _ := url.Parse("https://example.com")
	return u
}

func TestConfInvalid(t *testing.T) {
	type test struct {
		c           func() Conf
		valid       bool
		description string
	}
	tests := []test{
		{c: NilConf, valid: false, description: "NilConf"},
		{c: NilUpstream, valid: false, description: "NilUpstream"},
		{c: NilRoutes, valid: false, description: "NilRoutes"},
		{c: ZeroRoutes, valid: false, description: "ZeroRoutes"},
		{c: NilRouteRate, valid: false, description: "NilRouteRate"},
		{c: NilRoutePattern, valid: false, description: "NilRoutePattern"},
		{c: NilMethodsRoute, valid: false, description: "NilMethodsRoute"},
		{c: ZeroMethodsRoute, valid: false, description: "ZeroMethodsRoute"},
		{c: ZeroRouteRates, valid: false, description: "ZeroRouteRates"},
		{c: ValidConfSingleRoute, valid: true, description: "ValidConfSingleRoute"},
		{
			c:           InvalidConfBySecondRoutePattern,
			valid:       false,
			description: "InvalidConfBySecondRoutePattern",
		},
		{
			c:           InvalidConfBySecondRouteRates,
			valid:       false,
			description: "InvalidConfBySecondRouteRates",
		},
		{
			c:           InvalidConfBySecondRouteMethods,
			valid:       false,
			description: "InvalidConfBySecondRouteMethods",
		},
		{
			c:           InvalidConfBySecondRouteEmptyMethods,
			valid:       false,
			description: "InvalidConfBySecondRouteEmptyMethods",
		},
		{
			c:           ValidConfMultipleRoutes,
			valid:       true,
			description: "ValidConfMultipleRoutes",
		},
	}
	for _, test := range tests {
		hasError := ValidateConf(test.c(), false)
		gotIsValid := !hasError
		if gotIsValid != test.valid {
			t.Errorf("case %v expected %v got %v", test.description, test.valid, gotIsValid)
		}
	}
}
