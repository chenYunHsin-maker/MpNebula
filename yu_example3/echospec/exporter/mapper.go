package exporter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricMapper is a container that bundles a set of metrics that all from structs,
// and each field of the structs is mapped to a metric.
type MetricMapper struct {
	subsystem  string
	labelNames []string
	metrics    map[string]*prometheus.MetricVec
}

// NewMetricMapper return a new *MetricMapper with metric name subsystem.
func NewMetricMapper(subsystem string, labelNames []string) *MetricMapper {
	m := &MetricMapper{
		subsystem:  subsystem,
		labelNames: labelNames,
		metrics:    map[string]*prometheus.MetricVec{},
	}
	return m
}

// RegisterMetrics register all metrics from given structs.
func (m *MetricMapper) RegisterMetrics(structs ...interface{}) {
	for _, s := range structs {
		m.registerMetrics(s)
	}
}

func (m *MetricMapper) registerMetrics(s interface{}) {
	v := reflect.ValueOf(s).Elem()
	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	t := v.Type()
	if !strings.HasSuffix(t.Name(), "Metrics") {
		panic("invalid type name")
	}

	var labelNames []string
	tag := t.Field(0).Tag.Get("labelnames")
	if tag != "" {
		labelNames = strings.Split(tag, ",")
		labelNames = append(m.labelNames, labelNames...)
	} else {
		labelNames = m.labelNames
	}

	for i := 0; i < t.NumField(); i++ {
		tag = t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		switch v.Field(i).Interface().(type) {
		case int32, int64, uint32, uint64, float32, float64:
			vec := m.newMetricVec(name, labelNames, t.Field(i).Tag.Get("help"))
			prometheus.MustRegister(vec)
		default:
			panic("field type not supported")
		}

	}
}

func (m *MetricMapper) newMetricVec(name string, lns []string, help string) *prometheus.MetricVec {
	metric := prometheus.NewUntypedVec(
		prometheus.UntypedOpts{
			Help:      help,
			Subsystem: m.subsystem,
			Name:      name,
		}, lns)
	if _, ok := m.metrics[name]; ok {
		panic("metric has already been registered")
	}
	m.metrics[name] = metric.MetricVec
	return metric.MetricVec
}

// GetMetricVec returns prometheus MetricVec object by name.
func (m *MetricMapper) getMetricVec(name string) *prometheus.MetricVec {
	if metric, ok := m.metrics[name]; ok {
		return metric
	}
	return nil
}

func (m *MetricMapper) getSubsystem() string {
	return m.subsystem
}

func (m *MetricMapper) scrapeMetrics(s interface{}, lvs []string) error {
	v := reflect.ValueOf(s).Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%s not struct", v.Kind().String())
	}

	t := v.Type()
	if strings.HasSuffix(t.Name(), "Metrics") {
		m.exportMetrics(t, v, lvs)
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.FieldByName(f.Name)

		var newLvs []string
		tag := f.Tag.Get("label")
		if tag != "" {
			lv := strings.Split(tag, "=")[1]
			if fv.Kind() == reflect.String && lv == "?" {
				lvs = append(lvs, fv.String())
				continue
			} else {
				newLvs = append(lvs, lv)
			}
		} else {
			newLvs = lvs
		}

		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			if err := m.scrapeMetrics(fv.Interface(), newLvs); err != nil {
				return err
			}
		} else if fv.Kind() == reflect.Slice {
			if fv.IsNil() {
				continue
			}
			for j := 0; j < fv.Len(); j++ {
				if err := m.scrapeMetrics(fv.Index(j).Interface(), newLvs); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *MetricMapper) exportMetrics(t reflect.Type, v reflect.Value, lvs []string) {
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		tag := t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		var val float64
		switch f.Interface().(type) {
		case int32, int64:
			val = float64(f.Int())
		case uint32, uint64:
			val = float64(f.Uint())
		case float32, float64:
			val = f.Float()
		}
		m.getMetricVec(name).WithLabelValues(lvs...).(prometheus.Untyped).Set(val)
	}
}
