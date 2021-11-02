# db_test

## 说明

threadpool - 管理并发数量，并汇集结果（查询耗时、并发数量）
util - 用于查询的worker
executor - 监控集群并进行伸缩等操作

## Usage

```
go build -o test .
./test -kubeconfig=$HOME/.kube/config  -scale=false -dynamic=false -qsize=3 -wsize=10
```