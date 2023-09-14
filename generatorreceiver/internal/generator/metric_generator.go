package generator

import (
	"time"
	"math/rand"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
)

type MetricGenerator struct {
	metricCount int
	random         *rand.Rand
}

func NewMetricGenerator(seed int64) *MetricGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)
	return &MetricGenerator{
		metricCount: 0,
		random: r,
	}
}

func (g *MetricGenerator) Generate(metric *topology.Metric, serviceName string) (pmetric.Metrics, bool) {
	metric.Random = g.random
	metrics := pmetric.NewMetrics()

	if !metric.ShouldGenerate() {
		return metrics, false
	}

	rms := metrics.ResourceMetrics().AppendEmpty()
	rms.Resource().Attributes().PutStr("service.name", serviceName)

	m := rms.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName(metric.Name)
	if metric.Type == "Gauge" {
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		dp.SetDoubleValue(metric.GetValue())
		for k, v := range metric.GetTags() {
			dp.Attributes().PutStr(k, v)
		}
	} else if metric.Type == "Sum" {
		// TODO: support int-type values
		// TODO: support cumulative?
		m.SetEmptySum()
		m.Sum().SetIsMonotonic(true)
		m.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
		dp := m.Sum().DataPoints().AppendEmpty()
		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		dp.SetDoubleValue(metric.GetValue())
		for k, v := range metric.GetTags() {
			dp.Attributes().PutStr(k, v)
		}
	}
	// TODO: support histograms!

	g.metricCount = g.metricCount + 1
	return metrics, true
}
