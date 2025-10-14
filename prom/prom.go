package prom

import "github.com/prometheus/client_golang/prometheus"

type Prom struct {
	registry *prometheus.Registry
	metrics  []prometheus.Collector
}

func NewProm(collectors []prometheus.Collector) (*Prom, error) {
	c := &Prom{
		registry: prometheus.NewRegistry(),
		metrics:  collectors,
	}
	for _, metric := range c.metrics {
		if err := c.registry.Register(metric); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (p *Prom) Registry() *prometheus.Registry {
	return p.registry
}
