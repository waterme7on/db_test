package main

import "time"

const ResizeInterval = 120 * time.Second
const MonitorInterval = 5 * time.Second
const QueryInterval = 5 * time.Second
const DSN = "http://root@10.77.50.201:31314"
const PrometheusServerAddr = "http://10.77.50.201:31111"
const TestInterval = 60 * 80 * time.Second
const ScaleInterval = 60 * time.Second

const (
	PodMemoryUsagePercentage = "100 * (sum(container_memory_rss{pod=~\"%s.*\"}) by(pod)/1024/1024/1024) / (sum(container_spec_memory_limit_bytes) by(pod)/1024/1024/1024-8)"
	PodCpuUsagePercentage    = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod) / (((sum(container_spec_cpu_quota) by (pod))/ (sum(container_spec_cpu_period) by (pod))) - 2) * 100"
	PodCpuUsage              = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod)" // core
	PodMemoryUsage           = "sum(container_memory_rss{pod=~\"%s.*\"}) by(pod)/1024/1024/1024"          // GB
)
