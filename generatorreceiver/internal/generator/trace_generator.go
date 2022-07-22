package generator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gonum.org/v1/gonum/stat/distuv"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"go.opentelemetry.io/collector/model/pdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TraceGenerator struct {
	topology       *topology.Topology
	service        string
	route          string
	sequenceNumber int
	random         *rand.Rand
	sync.Mutex
	tagNameGenerator topology.Generator
	flagManager      *flags.FlagManager
}

func NewTraceGenerator(t *topology.Topology, seed int64, service string, route string, fm *flags.FlagManager) *TraceGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)

	tg := &TraceGenerator{
		topology:    t,
		random:      r,
		service:     service,
		route:       route,
		flagManager: fm,
	}
	return tg
}

func (g *TraceGenerator) genTraceId() pdata.TraceID {
	g.Lock()
	defer g.Unlock()
	traceIdBytes := make([]byte, 16)
	g.random.Read(traceIdBytes)
	var traceId [16]byte
	copy(traceId[:], traceIdBytes)
	return pdata.NewTraceID(traceId)
}

func (g *TraceGenerator) genSpanId() pdata.SpanID {
	g.Lock()
	defer g.Unlock()
	traceIdBytes := make([]byte, 16)
	g.random.Read(traceIdBytes)
	var traceId [8]byte
	copy(traceId[:], traceIdBytes)
	return pdata.NewSpanID(traceId)
}

// starts from now
func (g *TraceGenerator) Generate(startTimeMicros int64) *pdata.Traces {
	rootService := g.topology.GetServiceTier(g.service)
	traces := pdata.NewTraces()
	if g.shouldCreateSpanForRoute(rootService, g.route) {
		g.createSpanForServiceRouteCall(&traces, rootService, g.route, startTimeMicros, g.genTraceId(), pdata.NewSpanID([8]byte{0x0}))
	}
	return &traces
}

func (g *TraceGenerator) shouldCreateTagSet(ts topology.TagSet) bool {
	// TODO: I'm changing this to not panic if the flag doesn't exist,
	// and act as though it's unset, but we might want some kind of
	// validation instead.
	if len(ts.FlagSet) > 0 {
		f := g.flagManager.GetFlag(ts.FlagSet)
		return f != nil && f.Enabled()
	} else if len(ts.FlagUnset) > 0 {
		f := g.flagManager.GetFlag(ts.FlagUnset)
		return !(f != nil && f.Enabled())
	}
	return true
}

func (g *TraceGenerator) shouldCreateSpanForRoute(serviceTier *topology.ServiceTier, r string) bool {
	// TODO: I'm changing this to not panic if the flag doesn't exist,
	// and act as though it's unset, but we might want some kind of
	// validation instead.

	// TODO: multiple routes with the same name not supported
	route := serviceTier.GetRoute(r)

	if len(route.FlagSet) > 0 {
		f := g.flagManager.GetFlag(route.FlagSet)
		return f != nil && f.Enabled()
	} else if len(route.FlagUnset) > 0 {
		f := g.flagManager.GetFlag(route.FlagUnset)
		return !(f != nil && f.Enabled())
	}
	return true
}

func (g *TraceGenerator) createSpanForServiceRouteCall(traces *pdata.Traces, serviceTier *topology.ServiceTier, routeName string, startTimeMicros int64, traceId pdata.TraceID, parentSpanId pdata.SpanID) *pdata.Span {
	// logger := log.New(os.Stdout, "trace_generator: ", log.LstdFlags)
	serviceTier.Random = g.random
	route := serviceTier.GetRoute(routeName)

	rspanSlice := traces.ResourceSpans()
	rspan := rspanSlice.AppendEmpty()

	resource := rspan.Resource()

	resource.Attributes().InsertString(string(semconv.ServiceNameKey), serviceTier.ServiceName)

	resourceAttributeSet := serviceTier.GetResourceAttributeSet()
	if resourceAttributeSet != nil {
		for k, v := range resourceAttributeSet.ResourceAttributes {
			resource.Attributes().InsertString(k, fmt.Sprintf("%v", v))
		}
	}

	ils := rspan.InstrumentationLibrarySpans().AppendEmpty()
	spans := ils.Spans()

	span := spans.AppendEmpty()
	newSpanId := g.genSpanId()
	span.SetName(routeName)
	span.SetTraceID(traceId)
	span.SetParentSpanID(parentSpanId)
	span.SetSpanID(newSpanId)
	span.SetKind(pdata.SpanKindServer)
	span.Attributes().InsertString("load_generator.seq_num", fmt.Sprintf("%v", g.sequenceNumber))

	tagSet := serviceTier.GetTagSet(routeName)
	for _, ts := range tagSet {
		if !g.shouldCreateTagSet(ts) {
			continue
		}
		for k, v := range ts.Tags {
			span.Attributes().InsertString(k, v)
		}
		for _, tg := range ts.TagGenerators {
			tg.Random = g.random
			for k, v := range tg.GenerateTags() {
				span.Attributes().InsertString(k, v)
			}
		}
	}

	maxEndTime := startTimeMicros
	var percentileBasedLatency int64
	if route.LatencyPercentiles != nil {
		percentileBasedLatency = int64(calculateLatencyBasedOnPercentiles(route.LatencyPercentiles)) * 1000000
	}
	for s, r := range route.DownstreamCalls {
		var childStartTimeMicros int64
		if route.LatencyPercentiles != nil {
			childStartTimeMicros = startTimeMicros + percentileBasedLatency
		} else {
			childStartTimeMicros = startTimeMicros + (g.random.Int63n(route.MaxLatencyMillis * 1000000))
		}
		childSvc := g.topology.GetServiceTier(s)
		if g.shouldCreateSpanForRoute(childSvc, r) {
			g.createSpanForServiceRouteCall(traces, childSvc, r, childStartTimeMicros, traceId, newSpanId)
		}
		maxEndTime = Max(maxEndTime, childStartTimeMicros)
	}

	var ownDuration int64
	if route.LatencyPercentiles != nil {
		ownDuration = percentileBasedLatency
	} else {
		ownDuration = g.random.Int63n(route.MaxLatencyMillis * 1000000)
	}

	span.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(0, startTimeMicros)))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(time.Unix(0, maxEndTime+ownDuration)))
	g.sequenceNumber = g.sequenceNumber + 1
	return &span
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func calculateLatencyBasedOnPercentiles(routePercentiles *topology.LatencyPercentiles) float64 {
	_, p50, p95, p99, p999, _, err := routePercentiles.ParseDurations()
	if err != nil {
		return 0
	}

	dist := distuv.Normal{
		Mu:    float64(p50.Microseconds()), // mean is p50 because we're using a normal distribution, todo - support other distribution type?
		Sigma: float64(p50.Microseconds()) / 2,
		// standard deviation based on p50 could probably do a better normalized
		// guess on this based on the percentiles given
	}
	// +(p95.Microseconds()/3)
	// normalize variance
	// variance := (rand.Float64() - float64(p0.Microseconds())) / float64(p100.Microseconds()-p0.Microseconds()) * float64(p50.Microseconds())

	genNumber := rand.Float64()
	variance := genNumber * float64(p50.Microseconds())
	switch {
	case genNumber <= 0.05:
		// 5% of requests are p95
		return float64(p95.Microseconds()) + variance
	case genNumber <= 0.01:
		// 1% of requests are p99
		return float64(p99.Microseconds()) + variance
	case genNumber <= 0.001:
		// 0.1% of requests are p999
		return float64(p999.Microseconds()) + variance
	default:
		return dist.Rand()
	}
}
