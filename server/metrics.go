package server

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// StartMetrics start pushing Prometheus metrics
func StartMetrics(gameserverStore GameserverStore) {
	go func() {
		metrics := make(map[string]prometheus.Gauge, 0)
		for {
			servers, _ := gameserverStore.ListGameservers()
			for _, server := range servers {
				if _, ok := metrics[server.Definition.UUID]; !ok {
					metric := prometheus.NewGauge(prometheus.GaugeOpts{
						Name: "gameserver_running",
						ConstLabels: prometheus.Labels{
							"name":  server.Definition.Name,
							"owner": server.Definition.Owner,
							"game":  server.Definition.Game,
						},
					})
					prometheus.MustRegister(metric)
					metrics[server.Definition.UUID] = metric
				}
				metrics[server.Definition.UUID].Set(1)
			}

			for UUID, metric := range metrics {
				toRemove := true
				for _, server := range servers {
					if UUID == server.Definition.UUID {
						toRemove = false
						break
					}
				}

				if toRemove == true {
					prometheus.Unregister(metric)
					delete(metrics, UUID)
				}
			}

			time.Sleep(time.Duration(1 * time.Second))
		}
	}()
}
