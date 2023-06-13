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

var topologyTestFrontend = topology.ServiceTier{

	ServiceName: "frontend",
	Routes: map[string]*topology.ServiceRoute{
		"/product": {
			DownstreamCalls: []topology.Call{
				{Service: "productcatalogservice", Route: "/GetProducts"},
				{Service: "recommendationservice", Route: "/GetRecommendations"},
			},
			MaxLatencyMillis: 20,
		},

		"/checkout": {MaxLatencyMillis: 20},
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

var topologyTestCatalogService = topology.ServiceTier{
	ServiceName: "productcatalogservice",
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

var topologyTestRecommendationService = topology.ServiceTier{
	ServiceName: "recommendationservice",
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

var testTopology = &topology.Topology{
	Services: map[string]*topology.ServiceTier{
		"frontend":              &topologyTestFrontend,
		"productcatalogservice": &topologyTestCatalogService,
		"recommendationservice": &topologyTestRecommendationService,
	},
}

func TestTraceGenerator_createSpanForServiceRouteCall2(t *testing.T) {

	type args struct {
		serviceName    string
		routeName      string
		startTimeNanos int64
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "span with downstream calls",
			args: args{
				serviceName:    "frontend",
				routeName:      "/product",
				startTimeNanos: 000000123,
			},
		},

		{
			name: "span without downstream calls",
			args: args{
				serviceName:    "frontend",
				routeName:      "/checkout",
				startTimeNanos: 000000123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traces := ptrace.NewTraces()
			g := NewTraceGenerator(testTopology, rand.New(rand.NewSource(rand.Int63())), tt.args.serviceName, tt.args.routeName)
			genTraceID := g.genTraceId()
			rootSpan := g.createSpanForServiceRouteCall(&traces, g.service, g.route, tt.args.startTimeNanos, genTraceID, pcommon.NewSpanIDEmpty())
			convertedSpanStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))

			require.Equal(t, rootSpan.StartTimestamp(), convertedSpanStartTime)
			require.Equal(t, rootSpan.Name(), tt.args.routeName)
			require.Equal(t, rootSpan.TraceID(), genTraceID)
			require.Equal(t, rootSpan.ParentSpanID(), pcommon.NewSpanIDEmpty()) //root span will have parent span id of 0

			resourceSpans := traces.ResourceSpans()
			for i := 1; i < resourceSpans.Len(); i++ {

				resourceSpan := resourceSpans.At(i)
				scopeSpan := resourceSpan.ScopeSpans().At(0).Spans().At(0)
				require.Equal(t, scopeSpan.ParentSpanID(), rootSpan.SpanID())
				childSpanStartTime := scopeSpan.StartTimestamp()
				require.LessOrEqual(t, rootSpan.StartTimestamp(), childSpanStartTime)
				childSpanEndTime := scopeSpan.EndTimestamp()
				require.GreaterOrEqual(t, rootSpan.EndTimestamp(), childSpanEndTime)
				require.Equal(t, scopeSpan.TraceID(), genTraceID)

				serviceName, ok := resourceSpan.Resource().Attributes().Get("service.name")
				require.True(t, ok)
				service, ok := testTopology.Services[serviceName.AsString()]
				require.True(t, ok)
				_, ok = service.Routes[scopeSpan.Name()]
				require.True(t, ok)

				if len(service.TagSets) > 0 {
					for tagKey, tagValue := range service.TagSets[0].Tags {
						retrievedTagValue, exists := scopeSpan.Attributes().Get(tagKey)
						require.Equal(t, tagValue, retrievedTagValue.AsString())
						require.True(t, exists)
					}
				}

			}
		})
	}
}
