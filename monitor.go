package main

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
	for k, v := range result {
		fmt.Println(k, v)
	}
	return
}

// func main() {
// 	// skeleton code
// 	// 连接prometheus Client
// 	address := PrometheusServerAddr
// 	p := PromMonitor{}
// 	p.Connect(address)
// 	p.Query(prometheusUtil.PodCpuUsage, "gourdstore-slave")
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
