package main

/*
	promMonotor 通过prometheus提供的go客户端连接prometheus，进行查询并最终处理得到的结果
*/

import (
	"context"
	"fmt"
	"log"
	"time"

	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
	"k8s.io/client-go/kubernetes"
)

type PromMonitor struct {
	client     *prometheus.API
	kubeClient *kubernetes.Clientset
}

func (p *PromMonitor) Connect(address string) (err error) {
	_, queryClient, err := prometheusUtil.GetPrometheusClient(address)
	p.client = queryClient
	return
}

func (p *PromMonitor) Close() {
	return
}

func (p *PromMonitor) Query(promQL string, podPrefix string) (result map[string]string, err error) {
	queryResult, _, err := (*p.client).Query(context.TODO(), fmt.Sprintf(promQL, podPrefix), time.Now())
	if err != nil {
		log.Printf("PromMonitor query failed: %v, %v", promQL)
		return
	}
	result = extractResult(&queryResult)
	// for k, v := range result {
	// 	log.Println(k, v)
	// }
	return
}

// const PrometheusServerAddr = "http://10.77.50.201:31111"

// const (
// 	PodMemoryUsagePercentage = "100 * (sum(container_memory_rss{pod=~\"%s.*\"}) by(pod)/1024/1024/1024) / (sum(container_spec_memory_limit_bytes) by(pod)/1024/1024/1024-8)"
// 	PodCpuUsagePercentage    = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod) / (((sum(container_spec_cpu_quota) by (pod))/ (sum(container_spec_cpu_period) by (pod))) - 2) * 100"
// )

// func main() {
// 	// skeleton code
// 	// 连接prometheus Client
// 	address := PrometheusServerAddr
// 	p := PromMonitor{}
// 	p.Connect(address)
// 	res, _ := p.Query(PodMemoryUsagePercentage, "gourdstore-slave")
// 	fmt.Println(res)
// }

func extractResult(v *model.Value) (m map[string]string) {
	switch (*v).(type) {
	case model.Vector:
		vec, _ := (*v).(model.Vector)
		m = vectorToMap(&vec)
	default:
		break
	}
	return
}

func vectorToMap(v *model.Vector) (m map[string]string) {
	m = make(map[string]string)
	for i := range *v {
		m[(*v)[i].Metric.String()] = (*v)[i].Value.String()
	}
	return
}
