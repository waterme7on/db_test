

# echo "->>未开启动态伸缩<<-" >> time.log
# echo start `date` >> time.log
# echo "-scale=false -dynamic=true -qsize=3 -wsize=3 -workloadf=3-workload.csv" >> time.log
# ./test -kubeconfig=$HOME/.kube/config  -scale=false -dynamic=true -qsize=3 -wsize=3 -workloadf=3-workload.csv >> run.log
# echo end `date` >> time.log
# echo "------------" >> time.log

# sleep 60
# echo start `date` >> time.log
# echo "-scale=false -dynamic=true -qsize=4 -wsize=4 -workloadf=4-workload.csv" >> time.log
# ./test -kubeconfig=$HOME/.kube/config  -scale=false -dynamic=true -qsize=4 -wsize=4 -workloadf=4-workload.csv >> run.log
# echo end `date` >> time.log
# echo "------------" >> time.log

# sleep 60
# echo start `date` >> time.log
# echo "-scale=false -dynamic=true -qsize=5 -wsize=5 -workloadf=5-workload.csv" >> time.log
# ./test -kubeconfig=$HOME/.kube/config  -scale=false -dynamic=true -qsize=5 -wsize=5 -workloadf=5-workload.csv >> run.log
# echo end `date` >> time.log
# echo "------------" >> time.log




echo "->>开启动态伸缩<<-" >> time.log


kubectl scale --replicas=1 deployment/gourdstore-slave -n citybrain
sleep 60
echo start `date` >> time.log
echo "-scale=true -dynamic=true -qsize=3 -wsize=3 -workloadf=3-workload.csv" >> time.log
./test -kubeconfig=$HOME/.kube/config  -scale=true -dynamic=true -qsize=3 -wsize=3 -workloadf=3-workload.csv >> run.log
echo end `date` >> time.log
echo "------------" >> time.log

kubectl scale --replicas=1 deployment/gourdstore-slave -n citybrain
sleep 60
echo start `date` >> time.log
echo "-scale=true -dynamic=true -qsize=4 -wsize=4 -workloadf=4-workload.csv" >> time.log
./test -kubeconfig=$HOME/.kube/config  -scale=true -dynamic=true -qsize=4 -wsize=4 -workloadf=4-workload.csv >> run.log
echo end `date` >> time.log
echo "------------" >> time.log

kubectl scale --replicas=1 deployment/gourdstore-slave -n citybrain
sleep 60
echo start `date` >> time.log
echo "-scale=true -dynamic=true -qsize=5 -wsize=5 -workloadf=5-workload.csv" >> time.log
./test -kubeconfig=$HOME/.kube/config  -scale=true -dynamic=true -qsize=5 -wsize=5 -workloadf=5-workload.csv >> run.log
echo end `date` >> time.log
echo "------------" >> time.log
