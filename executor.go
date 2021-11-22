package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Scaler struct {
	podPrefix      string
	kubeClient     *kubernetes.Clientset
	deploymentName string
	namespace      string
	lastScaleTime  time.Time
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

func (s *Scaler) Scale(results map[string]map[string]string) bool {
	if !ScalerOn {
		return false
	}
	if time.Since(s.lastScaleTime) <= ScaleInterval {
		return false
	}

	deploy, err := s.kubeClient.AppsV1().Deployments(s.namespace).Get(context.TODO(), s.deploymentName, v1.GetOptions{})
	if err != nil {
		log.Fatalf("Scale Error: %v", err)
		return false
	}

	for k := range results {
		result := results[k]
		_, avg := sum_and_avg(result)
		if strings.Contains(k, "per") && avg >= 50 {
			*(deploy.Spec.Replicas) += 1
                        if *(deploy.Spec.Replicas) > 5 {
                        	return false
                        }
			s.kubeClient.AppsV1().Deployments(s.namespace).Update(context.TODO(), deploy, v1.UpdateOptions{})
			s.lastScaleTime = time.Now()
			return true
		}
		if strings.Contains(k, "per") && strings.Contains(k, "cpu") && avg <= 5 {
			*(deploy.Spec.Replicas) -= 1
                        if *(deploy.Spec.Replicas) < 1 {
                        	return false
                        }
			s.kubeClient.AppsV1().Deployments(s.namespace).Update(context.TODO(), deploy, v1.UpdateOptions{})
			s.lastScaleTime = time.Now()
			return true
		}
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
			cpuPerResult, _ := p.Query(PodCpuUsagePercentage, s.podPrefix)
			sum, avg := sum_and_avg(cpuPerResult)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(cpuPerResult), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-cpu-per, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(cpuPerResult), sum, avg))

			memPerResult, _ := p.Query(PodMemoryUsagePercentage, s.podPrefix)
			sum, avg = sum_and_avg(memPerResult)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(memPerResult), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-mem-per, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(memPerResult), sum, avg))

			if ScalerOn {
				result_ := map[string]map[string]string{
					"cpu_per": cpuPerResult,
					"mem_per": memPerResult,
				}
				s.Scale(result_)
			}

			memUsageResult, _ := p.Query(PodMemoryUsage, s.podPrefix)
			sum, avg = sum_and_avg(memUsageResult)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(memUsageResult), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-mem, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(memUsageResult), sum, avg))
			cpuUsageResult, _ := p.Query(PodCpuUsage, s.podPrefix)
			sum, avg = sum_and_avg(cpuUsageResult)
			log.Printf("Scaler: %v, %v, %v, %v", s.podPrefix, len(cpuUsageResult), sum, avg)
			file.WriteString(fmt.Sprintf("%v, %v-cpu, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), s.podPrefix, len(cpuUsageResult), sum, avg))
		}
		time.Sleep(MonitorInterval)
	}
}
