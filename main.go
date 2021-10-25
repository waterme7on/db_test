package main

import (
	"fmt"
	"log"
	"sync"

	_ "github.com/prestodb/presto-go-client/presto"
)

var tm = &ThreadsManager{
	mu:   &sync.Mutex{},
	cnt:  0,
	size: 10,
}

func main() {
	queries := []string{
		`
		with r as (select node1id from match (:person)-[:hascreated]->(:post) as graph 
		where node0id = '4145') 
		SELECT COUNT(*) 
		from hive.unibench.person c, mongodb.unibench.orders o, hbase.default.feedback f, r 
		WHERE c.id='4145' and o.personid = '4145' and f.personid='4145'
		`,
		`
		select distinct orders.personid
		from match (:person)-[:knows]->(:person) as graph
		join hbase.default.feedback on graph.node0id = feedback.personid
		join mongodb.unibench.orders on orders.personid = feedback.personid
		where orderdate = '2018-07-07' and orderline[1].productId = '2675'
		`,
		`
		with graph as (select node1id from match (:person)-[:knows]->(:person) as graph where node1id is not null)
		select feedback.personid, feedback.feedback
		from graph
		join hbase.default.feedback on feedback.personid = node1id
		join mongodb.unibench.orders on orders.personid = node1id
		where feedback like '%1.0%' and orderdate = '2018-07-07' and orderline[1].productId = '1380'
		`,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		db, err := connect()
		defer db.Close()
		if err != nil {
			return
		}
		tables, err := Query(db, queries[1])
		if err != nil {
			log.Fatalf("Query error: %s", err)
			return
		}
		fmt.Println(tables)
		wg.Done()
	}()
	wg.Wait()
}
