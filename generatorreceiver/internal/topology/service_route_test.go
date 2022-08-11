package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

var routeTestTopology = Topology{
	Services: map[string]*ServiceTier{
		"frontend":        &routeTestFrontend,
		"cartservice":     &routeTestCartService,
		"checkoutservice": &routeTestCheckoutService,
	},
}

var routeTestFrontend = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/cart": {
			DownstreamCalls:  map[string]string{"cartservice": "/GetCart"}, // valid service and route
			MaxLatencyMillis: 500,
			EmbeddedFlags: flags.EmbeddedFlags{
				FlagSet:   "someFlag",
				FlagUnset: "someOtherFlag"},
		},
		"/checkout": {
			DownstreamCalls:  map[string]string{"checkoutservice": "/FakeRoute"}, // valid service, invalid route
			MaxLatencyMillis: 500,
		},
		"/badroute": {
			DownstreamCalls:  map[string]string{"nonexistentservice": "/NonexistentRoute"}, // invalid service & route
			MaxLatencyMillis: 500,
		},
	},
}

var routeTestCartService = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/GetCart": {
			MaxLatencyMillis: 500,
			EmbeddedFlags: flags.EmbeddedFlags{
				FlagSet: "fakeFlag"},
		},
	},
}

var routeTestCheckoutService = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/PlaceOrder": {},
	},
}

func TestServiceRoute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		service string
		route   string
		flags   []string
		error   bool
	}{
		{
			name:    "Downstream service exists and has specified route with valid flags",
			service: "frontend",
			route:   "/cart",
			flags:   []string{"someFlag", "someOtherFlag"},
			error:   false,
		},
		{
			name:    "Downstream service exists but does not have specified route",
			service: "frontend",
			route:   "/checkout",
			error:   true,
		},
		{
			name:    "Downstream service does not exist",
			service: "frontend",
			route:   "/badroute",
			error:   true,
		},
		{
			name:    "Flag was specified but it does not exist",
			service: "cartservice",
			route:   "/GetCart",
			flags:   []string{"someFlag", "someOtherFlag"},
			error:   true,
		},
		{
			name:    "Missing maxLatencyMillis",
			service: "checkoutservice",
			route:   "/PlaceOrder",
			error:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := routeTestTopology.GetServiceTier(tt.service).GetRoute(tt.route)

			flags.Manager.Clear()
			theFlags := make([]flags.FlagConfig, 0, len(tt.flags))
			for _, name := range tt.flags {
				theFlags = append(theFlags, flags.FlagConfig{Name: name})
			}
			flags.Manager.LoadFlags(theFlags, zap.NewNop())

			err := r.validate(routeTestTopology)
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
		})
	}
}
