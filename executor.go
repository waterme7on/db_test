package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
)

type Scaler struct {
	podPrefix string
}

func (s *Scaler) Run(ctx context.Context) {
	p := &PromMonitor{}
	p.Connect(PrometheusServerAddr)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, _ := p.Query(prometheusUtil.PodCpuUsage, s.podPrefix)
			if len(result) != 0 {
				var avg float64
				var sum float64
				for _, v := range result {
					fv, _ := strconv.Atoi(v)
					sum += float64(fv)
				}
				avg = sum / float64(len(result))
				log.Printf("Scaler: %v, %v", s.podPrefix, avg)
			}
		}
		time.Sleep(MonitorInterval)
	}
}
