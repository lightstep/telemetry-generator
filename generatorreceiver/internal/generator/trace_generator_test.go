package generator

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var topoTestFrontend = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/product": {
			DownstreamCalls: []topology.Call{
				{"productcatalogservice", "/GetProducts"},
				{"recommendationservice", "/GetRecommendations"},
			},
			MaxLatencyMillis: 100,
		},
		"/cart": {
			DownstreamCalls:  []topology.Call{{"recommendationservice", "/GetRecommendations"}},
			MaxLatencyMillis: 200,
		},
	},
}

var topoTestCatalogService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{"/GetProducts": {
		MaxLatencyMillis: 300,
	}},
}

var topoTestCyclicalCatalogService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/GetProducts": {
			DownstreamCalls:  []topology.Call{{"frontend", "/cart"}},
			MaxLatencyMillis: 400,
		},
	},
}

var topoTestRecommendationService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/GetRecommendations": {
			DownstreamCalls:  []topology.Call{{"productcatalogservice", "/GetProducts"}},
			MaxLatencyMillis: 500,
		},
	},
}

func TestTraceGenerator_createSpanForServiceRouteCall(t *testing.T) {

	type args struct {
		traces         *ptrace.Traces
		serviceName    string
		routeName      string
		startTimeNanos int64
	}

	traces := ptrace.NewTraces()
	tests := []struct {
		name string
		topo *topology.Topology
		args args
	}{
		{
			name: "valid args",
			topo: &topology.Topology{
				Services: map[string]*topology.ServiceTier{
					"frontend":              &topoTestFrontend,
					"productcatalogservice": &topoTestCatalogService,
					"recommendationservice": &topoTestRecommendationService,
				},
			},
			args: args{
				traces:         &traces,
				serviceName:    "frontend",
				routeName:      "/product",
				startTimeNanos: 000000123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewTraceGenerator(tt.topo, rand.New(rand.NewSource(rand.Int63())), tt.args.serviceName, tt.args.routeName)
			genTraceID := g.genTraceId()
			span := g.createSpanForServiceRouteCall(tt.args.traces, g.service, g.route, tt.args.startTimeNanos, genTraceID, pcommon.NewSpanIDEmpty())
			require.Equal(t, span.Name(), tt.args.routeName)
			require.Equal(t, span.Name(), g.route)
			require.Equal(t, span.TraceID(), genTraceID)
			require.Equal(t, span.ParentSpanID(), pcommon.NewSpanIDEmpty())

			convertedStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))
			require.Equal(t, span.StartTimestamp(), convertedStartTime)

			endTime := tt.args.startTimeNanos + g.topology.Services[tt.args.serviceName].Routes[tt.args.routeName].SampleLatency(genTraceID)
			convertedEndTime := pcommon.NewTimestampFromTime(time.Unix(0, endTime))
			require.Equal(t, span.EndTimestamp(), convertedEndTime)

		})
	}
}
