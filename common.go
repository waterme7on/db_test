package main

import "time"

const ResizeInterval = 180 * time.Second
const MonitorInterval = 5 * time.Second
const QueryInterval = 5 * time.Second
const DSN = "http://root@10.77.50.201:31314"
const PrometheusServerAddr = "http://10.77.50.201:31111"
const TestInterval = 600 * time.Second
