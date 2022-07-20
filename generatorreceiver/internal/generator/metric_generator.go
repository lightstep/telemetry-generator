package generator

import (
	"math"
	"math/rand"
	"time"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"go.opentelemetry.io/collector/model/pdata"
)

type MetricGenerator struct {
	metricCount int
	random *rand.Rand
	flagManager *flags.FlagManager
}

func NewMetricGenerator(seed int64, fm *flags.FlagManager) *MetricGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)
	return &MetricGenerator{
		metricCount: 0,
		random: r,
		flagManager: fm,
	}
}

func (g *MetricGenerator) Generate(metricName string, metricType string, serviceName string, flagSet string, flagUnset string) pdata.Metrics {
	metrics := pdata.NewMetrics()

	// TODO: toggle metric gen based on flagSet/flagUnset
	// g.flagManager.GetFlag(flagSet)

	rms := metrics.ResourceMetrics().AppendEmpty()
	rms.Resource().Attributes().InsertString("service.name", serviceName)

	m := rms.InstrumentationLibraryMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName(metricName)
	if metricType == "Gauge" {
		m.SetDataType(pdata.MetricDataTypeGauge)
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetDoubleVal((math.Sin(.01 * float64(g.metricCount)) + 1) * 100)
	} else if metricType == "Sum" {
		m.SetDataType(pdata.MetricDataTypeSum)
		dp := m.Sum().DataPoints().AppendEmpty()
		dp.SetTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetDoubleVal((math.Sin(.01 * float64(g.metricCount)) + 1) * 100)
	}

	g.metricCount = g.metricCount + 1
	return metrics
}