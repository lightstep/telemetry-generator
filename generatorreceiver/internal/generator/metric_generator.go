package generator

import (
	"go.opentelemetry.io/collector/model/pdata"
	"time"
)

type MetricGenerator struct {

}

func (g *MetricGenerator) Generate(metricName string, serviceName string) *pdata.Metrics {
	metrics := pdata.NewMetrics()
	rms := metrics.ResourceMetrics().AppendEmpty()
	rms.Resource().Attributes().InsertString("service.name", serviceName)

	ilms := rms.InstrumentationLibraryMetrics().AppendEmpty()
	m := ilms.Metrics().AppendEmpty()
	m.SetName(metricName)
	m.SetDataType(pdata.MetricDataTypeGauge)

	dps := m.Gauge().DataPoints().AppendEmpty()
	dps.SetStartTimestamp(pdata.TimestampFromTime(time.Now()))
	dps.SetValue(100)
	return &metrics
}