package generator

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// scafolding for tests
var topoTestFrontend = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/product": {
			DownstreamCalls: []topology.Call{
				{Service: "productcatalogservice", Route: "/GetProducts"},
				{Service: "recommendationservice", Route: "/GetRecommendations"},
			},
			MaxLatencyMillis: 100,
		},
	},
}

var topoTestCatalogService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{"/GetProducts": {
		MaxLatencyMillis: 300,
	}},
}

var topoTestRecommendationService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/GetRecommendations": {
			DownstreamCalls: []topology.Call{
				{Service: "productcatalogservice", Route: "/GetProducts"}},
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
			require.Equal(t, span.ParentSpanID(), pcommon.NewSpanIDEmpty()) //root route will have parent span id of 0

			convertedStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))
			require.Equal(t, span.StartTimestamp(), convertedStartTime)

			//create child spans for downstream calls to ultimately calculate the end time
			serviceTier := g.topology.GetServiceTier(tt.args.serviceName)
			route := serviceTier.GetRoute(g.route)
			endTime := tt.args.startTimeNanos + g.topology.Services[tt.args.serviceName].Routes[tt.args.routeName].SampleLatency(genTraceID)
			for _, c := range route.DownstreamCalls {
				var childStartTimeNanos = tt.args.startTimeNanos + route.SampleLatency(genTraceID)

				childSpan := g.createSpanForServiceRouteCall(tt.args.traces, c.Service, c.Route, childStartTimeNanos, genTraceID, g.genSpanId())
				endTime = Max(endTime, int64(childSpan.EndTimestamp()))
			}
			convertedEndTime := pcommon.NewTimestampFromTime(time.Unix(0, endTime))
			fmt.Println(convertedEndTime, span.EndTimestamp())
			require.Equal(t, span.EndTimestamp(), convertedEndTime)

		})
	}
}
