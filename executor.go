package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Scaler struct {
	podPrefix      string
	kubeClient     *kubernetes.Clientset
	deploymentName string
	namespace      string
}

func sum_and_avg(input map[string]string) (sum float64, avg float64) {
	sum = 0
	avg = 0
	if len(input) == 0 {
		return
	}
	for _, v := range input {
		fv, _ := strconv.ParseFloat(v, 64)
		sum += fv
	}
	avg = sum / float64(len(input))
	return
}

func (s *Scaler) QueryMetric(p *PromMonitor, Metric string) map[string]string {
	result, _ := p.Query(PodCpuUsagePercentage, s.podPrefix)
	return result
}

func (s *Scaler) Scale(result map[string]string, Metric string) bool {
	if !ScalerOn {
		return false
	}
	_, avg := sum_and_avg(result)
	if avg >= 70 {
		deploy, err := s.kubeClient.AppsV1().Deployments(s.namespace).Get(context.TODO(), s.deploymentName, v1.GetOptions{})
		if err != nil {
			log.Fatalf("Scale Error: %v", err)
		}
		*(deploy.Spec.Replicas) += 1
		s.kubeClient.AppsV1().Deployments(s.namespace).Update(context.TODO(), deploy, v1.UpdateOptions{})
		return true
	}
	return false
}

func (s *Scaler) Run(ctx context.Context) {
	p := &PromMonitor{}
	file, _ := os.Create(fmt.Sprintf("monitor-%v-%v-%v.csv", MaxQuerySize, DynamicWorkload, ScalerOn))
	defer file.Close()
	p.Connect(PrometheusServerAddr)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, _ := p.Query(PodCpuUsagePercentage, s.podPrefix)
			sum, avg := sum_and_avg(result)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(result), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-cpu-per, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(result), sum, avg))
			flag := s.Scale(result, PodCpuUsagePercentage)

			result, _ = p.Query(PodMemoryUsagePercentage, s.podPrefix)
			sum, avg = sum_and_avg(result)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(result), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-mem-per, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(result), sum, avg))
			if flag != true {
				s.Scale(result, PodMemoryUsagePercentage)
			}

			result, _ = p.Query(PodMemoryUsage, s.podPrefix)
			sum, avg = sum_and_avg(result)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(result), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-mem, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(result), sum, avg))
			result, _ = p.Query(PodCpuUsage, s.podPrefix)
			sum, avg = sum_and_avg(result)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(result), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-cpu, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(result), sum, avg))
		}
		time.Sleep(MonitorInterval)
	}
}
