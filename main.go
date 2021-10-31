package main

/*
	创建n个查询work，对数据库进行查询
	查询的数量由ThreadsPool进行管理
*/

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/prestodb/presto-go-client/presto"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

var tm = ThreadsPool(10)
var scaler = Scaler{
	podPrefix: "gourdstore-slave",
}

func main() {

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
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
			w := Worker{}
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

	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("Main routine Exit...", s)
			cancel()
			quit = true
		default:
			log.Println("other signal", s)
		}
		if quit {
			break
		}
	}
	log.Println("Main routine Start exit...")
	log.Println("Execute clean and wait for subroutines to quit...")
	wg.Wait()
	log.Println("Main routine end exit...")
}
