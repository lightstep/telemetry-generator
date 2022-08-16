package generator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

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
}

func NewTraceGenerator(t *topology.Topology, seed int64, service string, route string) *TraceGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)

	tg := &TraceGenerator{
		topology: t,
		random:   r,
		service:  service,
		route:    route,
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

func (g *TraceGenerator) Generate(startTimeNanos int64) *pdata.Traces {
	traces := pdata.NewTraces()

	g.createSpanForServiceRouteCall(&traces, g.service, g.route, startTimeNanos, g.genTraceId(), pdata.NewSpanID([8]byte{0x0}))

	return &traces
}

func pickBasedOnWeight(tagSets []topology.TagSet) topology.TagSet {
	totalWeight := 0.0
	for _, ts := range tagSets {
		totalWeight += ts.Weight
	}

	choice := rand.Float64() * totalWeight
	current := 0.0
	for _, ts := range tagSets {
		current += ts.Weight
		if choice < current {
			return ts
		}
	}
	return topology.TagSet{}
}

func (g *TraceGenerator) createSpanForServiceRouteCall(traces *pdata.Traces, serviceName string, routeName string, startTimeNanos int64, traceId pdata.TraceID, parentSpanId pdata.SpanID) *pdata.Span {
	serviceTier := g.topology.GetServiceTier(serviceName)
	serviceTier.Random = g.random
	route := serviceTier.GetRoute(routeName)

	if !route.ShouldGenerate() {
		return nil
	}

	rspanSlice := traces.ResourceSpans()
	rspan := rspanSlice.AppendEmpty()

	resource := rspan.Resource()

	resource.Attributes().InsertString(string(semconv.ServiceNameKey), serviceTier.ServiceName)

	resourceAttributeSet := serviceTier.GetResourceAttributeSet()
	if resourceAttributeSet != nil {
		attrs := resource.Attributes()
		resourceAttributeSet.ResourceAttributes.InsertTags(&attrs)
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
	filteredTagSets := []topology.TagSet{} //empty slice

	for _, ts := range tagSet {
		if !ts.ShouldGenerate() {
			continue
		}
		filteredTagSets = append(filteredTagSets, ts)
	}

	//once ShouldGenerate is true aka flag is set or unset, then worry about weights
	chosen := pickBasedOnWeight(filteredTagSets)

	//adds attributes to the tagSets that have specified weights
	attr := span.Attributes()
	chosen.Tags.InsertTags(&attr)
	for _, tg := range chosen.TagGenerators {
		tg.Random = g.random
		for k, v := range tg.GenerateTags() {
			span.Attributes().InsertString(k, v)
		}
	}

	// TODO: this is still a bit weird - we're calling each downstream route
	// after a sample of the current route's latency, which doesn't really
	// make sense - but maybe it's realistic enough?
	endTime := startTimeNanos + route.SampleLatency()
	for s, r := range route.DownstreamCalls {
		var childStartTimeNanos = startTimeNanos + route.SampleLatency()

		childSpan := g.createSpanForServiceRouteCall(traces, s, r, childStartTimeNanos, traceId, newSpanId)
		endTime = Max(endTime, int64(childSpan.EndTimestamp()))
	}

	span.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(0, startTimeNanos)))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(time.Unix(0, endTime)))
	g.sequenceNumber += 1
	return &span
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
