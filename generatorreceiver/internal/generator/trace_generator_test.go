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

// scafolding for tests
var topoTestFrontend = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/product": {
			DownstreamCalls: []topology.Call{
				{Service: "productcatalogservice", Route: "/GetProducts"},
				{Service: "recommendationservice", Route: "/GetRecommendations"},
			},
			MaxLatencyMillis: 20,
		},
	},
	TagSets: []topology.TagSet{
		{
			Tags: topology.TagMap{
				"version": "v01",
				"color":   "blue",
			},
		},
	},
}

var topoTestCatalogService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/GetProducts": {
			MaxLatencyMillis: 5,
		},
	},
	TagSets: []topology.TagSet{
		{
			Tags: topology.TagMap{
				"version": "v02",
				"color":   "red",
			},
		},
	},
}

var topoTestRecommendationService = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/GetRecommendations": {MaxLatencyMillis: 10},
	},
	TagSets: []topology.TagSet{
		{
			Tags: topology.TagMap{
				"version": "v03",
				"color":   "green",
			},
		},
	},
}

var topoTestFrontendWithoutDownStreamCalls = topology.ServiceTier{
	Routes: map[string]*topology.ServiceRoute{
		"/product": {MaxLatencyMillis: 10},
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
			name: "span with downstream calls",
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

		{
			name: "span without downstream calls",
			topo: &topology.Topology{
				Services: map[string]*topology.ServiceTier{
					"frontend": &topoTestFrontendWithoutDownStreamCalls,
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
			require.Equal(t, span.TraceID(), genTraceID)
			require.Equal(t, span.ParentSpanID(), pcommon.NewSpanIDEmpty()) //root span will have parent span id of 0

			convertedSpanStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))
			require.Equal(t, span.StartTimestamp(), convertedSpanStartTime)

			resourceSpans := traces.ResourceSpans()
			serviceTier := tt.topo.Services //get the service tiers from the topology

			i := 0
			for _, value := range serviceTier {

				if len(value.TagSets) > 0 {
					for tagKey, tagValue := range value.TagSets[0].Tags {
						retrievedTagValue, exists := resourceSpans.At(i).ScopeSpans().At(0).Spans().At(0).Attributes().Get(tagKey)
						require.Equal(t, tagValue, retrievedTagValue.AsString())
						require.True(t, exists)
						spanEndTime := resourceSpans.At(i).ScopeSpans().At(0).Spans().At(0).EndTimestamp() //there are no children of children so index 0
						require.GreaterOrEqual(t, span.EndTimestamp(), spanEndTime)
					}
					i++
				}

			}

		})
	}
}
