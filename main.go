package main

/*
	创建n个查询work，对数据库进行查询
	查询的数量由ThreadsPool进行管理
*/

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/prestodb/presto-go-client/presto"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	MasterUrl       string
	Kubeconfig      string
	ScalerOn        bool
	DynamicWorkload bool
	MaxQuerySize    int
	WorkerSize      int
)

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "Path to a Kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&MasterUrl, "master", "", "The address of the Kubernetes API server. Overrides any value in Kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&ScalerOn, "scale", false, "Trun on or off the scaler")
	flag.BoolVar(&DynamicWorkload, "dynamic", false, "Turn on or off dynamic workload")
	flag.IntVar(&MaxQuerySize, "qsize", 20, "Max query size")
	flag.IntVar(&WorkerSize, "wsize", 20, "Worker size")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	initWorkerSize := MaxQuerySize
	if DynamicWorkload {
		initWorkerSize = rand.Int() % MaxQuerySize
	}
	var tm = ThreadsPool(initWorkerSize)
	var scaler = Scaler{
		podPrefix:      "gourdstore-slave",
		deploymentName: "gourdstore-slave",
		namespace:      "citybrain",
	}

	cfg, err := clientcmd.BuildConfigFromFlags(MasterUrl, Kubeconfig)
	if err != nil {
		log.Fatalf("Error building Kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	scaler.kubeClient = kubeClient

	//创建监听退出chan
	c := make(chan os.Signal)
	quit := false
	//监听指定信号 ctrl+c kill
	ctx, cancel := context.WithTimeout(context.TODO(), TestInterval)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	var wg sync.WaitGroup
	resultCh := make(chan string, WorkerSize)
	wg.Add(3 + WorkerSize)
	for i := 0; i < WorkerSize; i++ {
		go func(i int) {
			w := Worker{
				tm: tm,
			}
			err := w.Init(DSN, i)
			defer w.Close()
			defer wg.Done()
			if err != nil {
				return
			}
			w.Run(ctx, resultCh)
		}(i)
	}
	go func() {
		tm.CollectResult(ctx, resultCh)
		defer wg.Done()
	}()
	go func() {
		scaler.Run(ctx)
		defer wg.Done()
	}()
	go func() {
		tm.Run(ctx)
		defer wg.Done()
	}()

	for {
		select {
		case s := <-c:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("Main routine Exit...", s)
				quit = true
			default:
				log.Println("other signal", s)
			}
		case <-ctx.Done():
			quit = true
		}
		if quit {
			break
		}
	}
	cancel()
	log.Println("Main routine Start exit...")
	log.Println("Execute clean and wait for subroutines to quit...")
	wg.Wait()
	log.Println("Main routine end exit...")
}
