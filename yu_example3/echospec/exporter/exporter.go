package exporter

import "fmt"

type metricExporter struct {
	mappers map[string]*MetricMapper
}

var defaultExporter = &metricExporter{mappers: map[string]*MetricMapper{}}

// AddMetricMapper registers provided MetricMapper, and the key to retrieve
// this MetricMapper is the subsystem value associated with it.
func AddMetricMapper(mapper *MetricMapper) {
	subsystem := mapper.getSubsystem()
	if _, ok := defaultExporter.mappers[subsystem]; ok {
		panic("metricMapper has already been registered")
	}
	defaultExporter.mappers[subsystem] = mapper
}

// ScrapeMetrics recursively parses struct object s to get metrics,
// and export them with common label values lvs and also some specific
// label values that will be extracted from struct s.
// The subsystem must be provided as it's the key mapping to a MetricMapper.
func ScrapeMetrics(subsystem string, s interface{}, lvs []string) error {
	mapper, ok := defaultExporter.mappers[subsystem]
	if !ok {
		return fmt.Errorf("metricMapper %s has not already been registered", subsystem)
	}

	return mapper.scrapeMetrics(s, lvs)
}
