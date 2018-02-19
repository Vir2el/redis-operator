package metrics_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spotahome/redis-operator/metrics"
)

func TestPrometheusMetrics(t *testing.T) {

	tests := []struct {
		name       string
		addMetrics func(pm *metrics.PromMetrics)
		expMetrics []string
		expCode    int
	}{
		{
			name: "Setting OK should give an OK",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterOK("testns", "test")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting Error should give an Error",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterError("testns", "test")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns"} 0`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting Error after ok should give an Error",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterOK("testns", "test")
				pm.SetClusterError("testns", "test")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns"} 0`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting OK after Error should give an OK",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterError("testns", "test")
				pm.SetClusterOK("testns", "test")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Multiple clusters should appear",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterOK("testns", "test")
				pm.SetClusterOK("testns", "test2")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns"} 1`,
				`redis_operator_controller_cluster_ok{name="test2",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Same name on different namespaces should appear",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterOK("testns1", "test")
				pm.SetClusterOK("testns2", "test")
			},
			expMetrics: []string{
				`redis_operator_controller_cluster_ok{name="test",namespace="testns1"} 1`,
				`redis_operator_controller_cluster_ok{name="test",namespace="testns2"} 1`,
			},
			expCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			path := "/awesome-metrics"

			// Create the muxer for testing.
			mx := http.NewServeMux()
			pm := metrics.NewPrometheusMetrics(path, mx)

			// Add metrics to prometheus.
			test.addMetrics(pm)

			// Make the request to the metrics.
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			mx.ServeHTTP(w, req)

			resp := w.Result()
			if assert.Equal(test.expCode, resp.StatusCode) {
				body, _ := ioutil.ReadAll(resp.Body)
				// Check all the metrics are present.
				for _, expMetric := range test.expMetrics {
					assert.Contains(string(body), expMetric)
				}
			}
		})
	}
}
