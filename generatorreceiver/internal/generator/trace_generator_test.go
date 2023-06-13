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

// func TestTraceGenerator_createSpanForServiceRouteCall(t *testing.T) {

// 	type args struct {
// 		serviceName    string
// 		routeName      string
// 		startTimeNanos int64
// 	}

// 	tests := []struct {
// 		name string
// 		topo *topology.Topology
// 		args args
// 	}{
// 		{
// 			name: "span with downstream calls",
// 			topo: &topology.Topology{
// 				Services: map[string]*topology.ServiceTier{
// 					"frontend":              &topologyTestFrontend,
// 					"productcatalogservice": &topologyTestCatalogService,
// 					"recommendationservice": &topologyTestRecommendationService,
// 				},
// 			},
// 			args: args{
// 				serviceName:    "frontend",
// 				routeName:      "/product",
// 				startTimeNanos: 000000123,
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			traces := ptrace.NewTraces()
// 			g := NewTraceGenerator(tt.topo, rand.New(rand.NewSource(rand.Int63())), tt.args.serviceName, tt.args.routeName)
// 			genTraceID := g.genTraceId()
// 			span := g.createSpanForServiceRouteCall(&traces, g.service, g.route, tt.args.startTimeNanos, genTraceID, pcommon.NewSpanIDEmpty())
// 			convertedSpanStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))

// 			require.Equal(t, span.StartTimestamp(), convertedSpanStartTime)
// 			require.Equal(t, span.Name(), tt.args.routeName)
// 			require.Equal(t, span.TraceID(), genTraceID)
// 			require.Equal(t, span.ParentSpanID(), pcommon.NewSpanIDEmpty()) //root span will have parent span id of 0

// 			resourceSpans := traces.ResourceSpans()
// 			serviceTier := tt.topo.Services
// 			indexResourceSpans := 0
// 			childSpan := resourceSpans.At(indexResourceSpans).ScopeSpans().At(0).Spans().At(0)

// 			if childSpan.ParentSpanID() != pcommon.NewSpanIDEmpty() {
// 				require.Equal(t, childSpan.ParentSpanID(), span.SpanID())
// 			}

// 			for _, service := range serviceTier {

// 				if len(service.TagSets) > 0 {
// 					for tagKey, tagValue := range service.TagSets[0].Tags {
// 						retrievedTagValue, exists := childSpan.Attributes().Get(tagKey)
// 						require.Equal(t, tagValue, retrievedTagValue.AsString())
// 						require.True(t, exists)
// 					}
// 				}

// 				_, exists := service.Routes[childSpan.Name()]
// 				require.True(t, exists)

// 				childSpanStartTime := childSpan.StartTimestamp()
// 				require.LessOrEqual(t, span.StartTimestamp(), childSpanStartTime)
// 				childSpanEndTime := childSpan.EndTimestamp() //there are no children of children so index 0
// 				require.GreaterOrEqual(t, span.EndTimestamp(), childSpanEndTime)
// 				indexResourceSpans++
// 			}

// 		})
// 	}
// }

func TestTraceGenerator_createSpanForServiceRouteCall2(t *testing.T) {

	type args struct {
		serviceName    string
		routeName      string
		startTimeNanos int64
	}

	tests := []struct {
		name string
		topo *topology.Topology
		args args
	}{
		{
			name: "span with downstream calls",
			topo: &topology.Topology{
				Services: map[string]*topology.ServiceTier{
					"frontend":              &topologyTestFrontend,
					"productcatalogservice": &topologyTestCatalogService,
					"recommendationservice": &topologyTestRecommendationService,
				},
			},
			args: args{
				serviceName:    "frontend",
				routeName:      "/product",
				startTimeNanos: 000000123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traces := ptrace.NewTraces()
			g := NewTraceGenerator(tt.topo, rand.New(rand.NewSource(rand.Int63())), tt.args.serviceName, tt.args.routeName)
			genTraceID := g.genTraceId()
			rootSpan := g.createSpanForServiceRouteCall(&traces, g.service, g.route, tt.args.startTimeNanos, genTraceID, pcommon.NewSpanIDEmpty())
			convertedSpanStartTime := pcommon.NewTimestampFromTime(time.Unix(0, tt.args.startTimeNanos))

			require.Equal(t, rootSpan.StartTimestamp(), convertedSpanStartTime)
			require.Equal(t, rootSpan.Name(), tt.args.routeName)
			require.Equal(t, rootSpan.TraceID(), genTraceID)
			require.Equal(t, rootSpan.ParentSpanID(), pcommon.NewSpanIDEmpty()) //root span will have parent span id of 0

			resourceSpans := traces.ResourceSpans()

			for i := 1; i < resourceSpans.Len(); i++ {

				childSpan := resourceSpans.At(i).ScopeSpans().At(0).Spans().At(0)
				require.Equal(t, childSpan.ParentSpanID(), rootSpan.SpanID())
				childSpanStartTime := childSpan.StartTimestamp()
				require.LessOrEqual(t, rootSpan.StartTimestamp(), childSpanStartTime)
				childSpanEndTime := childSpan.EndTimestamp()
				require.GreaterOrEqual(t, rootSpan.EndTimestamp(), childSpanEndTime)
				require.Equal(t, childSpan.TraceID(), genTraceID)

				service := getServiceFromRoute(childSpan.Name(), *tt.topo)
				fmt.Println(service)
				fmt.Printf("%+v\n", childSpan)

			}
		})
	}
}

func getServiceFromRoute(route string, topo topology.Topology) topology.ServiceTier {

	for _, service := range topo.Services {
		for _, serviceRoute := range service.Routes {
			if route == serviceRoute.Route {
				return *service
			}
		}
	}
	return topology.ServiceTier{}
}
