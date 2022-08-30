package topology

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var topoTestFrontend = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/product": {
			DownstreamCalls: []Call{
				{"productcatalogservice", "/GetProducts"},
				{"recommendationservice", "/GetRecommendations"},
			},
		},
		"/cart": {
			DownstreamCalls: []Call{{"recommendationservice", "/GetRecommendations"}},
		},
	},
}

var topoTestCatalogService = ServiceTier{
	Routes: map[string]*ServiceRoute{"/GetProducts": {}},
}

var topoTestCyclicalCatalogService = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/GetProducts": {DownstreamCalls: []Call{{"frontend", "/cart"}}},
	},
}

var topoTestRecommendationService = ServiceTier{
	Routes: map[string]*ServiceRoute{
		"/GetRecommendations": {DownstreamCalls: []Call{{"productcatalogservice", "/GetProducts"}}},
	},
}

func TestTopology_ValidateServiceGraph(t *testing.T) {
	tests := []struct {
		name       string
		topo       Topology
		rootRoutes []RootRoute
		error      bool
	}{
		{
			name: "Non-cyclical service/route graph",
			topo: Topology{
				Services: map[string]*ServiceTier{
					"frontend":              &topoTestFrontend,
					"productcatalogservice": &topoTestCatalogService,
					"recommendationservice": &topoTestRecommendationService,
				},
			},
			rootRoutes: []RootRoute{
				{
					Service:       "frontend",
					Route:         "/product",
					TracesPerHour: 100,
				},
				{
					Service:       "frontend",
					Route:         "/cart",
					TracesPerHour: 100,
				},
			},
			error: false,
		},
		{
			name: "Cyclical service/route graph",
			topo: Topology{
				Services: map[string]*ServiceTier{
					"frontend":              &topoTestFrontend,
					"productcatalogservice": &topoTestCyclicalCatalogService,
					"recommendationservice": &topoTestRecommendationService,
				},
			},
			rootRoutes: []RootRoute{
				{
					Service:       "frontend",
					Route:         "/product",
					TracesPerHour: 100,
				},
				{
					Service:       "frontend",
					Route:         "/cart",
					TracesPerHour: 100,
				},
			},
			error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.topo.Load() // needed for populating ServiceTier.ServiceName and ServiceRoute.Route
			err := tt.topo.ValidateServiceGraph(tt.rootRoutes)
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
		})
	}
}
