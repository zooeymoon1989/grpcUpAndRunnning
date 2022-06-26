package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// prometheus 初始化变量
var (
	Reg                 = prometheus.NewRegistry()
	GrpcMetrics         = grpc_prometheus.NewServerMetrics()
	CustomMetricCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "product_mgt_server_handle_count",
		Help: "Total number of RPCs handled on the server",
	}, []string{"name"})
)
