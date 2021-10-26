package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/openpgp/errors"
)

type Worker struct {
	db      *sql.DB
	id      int
	queries *[]string
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %d quit", w.id)
			return
		default:
			k := tm.Alloc()
			if k != -1 {
				log.Printf("worker %d executing", w.id)
				_, err := w.QueryToId(rand.Int() % len(*w.queries))
				if err != nil {
					log.Printf("Query error: %s", err)
				}
				tm.Free(k)
			}
		}
		time.Sleep(QueryInterval)
	}
}

func (w *Worker) Init(dsn string, id int) (err error) {
	w.queries = &[]string{
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
		`
		select feedback.personid, feedback.feedback
		from (select node1id from match (:person)-[:knows]-(:person) as graph where node0id = '4145')
		join hbase.default.feedback on feedback.personid = node1id
		join mongodb.unibench.orders on orders.personid = node1id
		where feedback like '%5.0%' and orderline[1].productId = '6406'
		`,
		`
		with dfn as (select * from(select node1id from match (:person)-[:knows]->(:person)<-[:knows]-(:person) as graph where node0id='4145' and node2id='24189255845124') where node1id is not null)
		select o.personid, orderline[1].productId
		from mongodb.unibench.orders o, dfn
		where o.personid=dfn.node1id
		`,
	}
	w.id = id
	err = w.Connect(dsn)
	log.Printf("worker %d init", w.id)
	return
}

func (w *Worker) Connect(dsn string) (err error) {
	w.db, err = sql.Open("presto", dsn)
	if err != nil {
		log.Fatalf("worker %d failed to open dsn %s", w.id, dsn)
		return
	}
	err = w.db.Ping()
	if err != nil {
		log.Fatalf("worker %d failed to ping db %s", w.id, dsn)
		return
	}
	return
}

func (w *Worker) Close() (err error) {
	err = w.db.Close()
	if err != nil {
		return
	}
	w.db = nil
	return
}

// 查询
func (w *Worker) QueryToId(queryId int) ([]map[string]interface{}, error) {
	var tables []map[string]interface{}
	if queryId < 0 || queryId >= len(*w.queries) {
		return tables, errors.UnsupportedError("query id exceed")
	}
	startTime := time.Now()
	sqltxt := (*w.queries)[queryId]
	log.Printf("worker:%d, query:%v\n", w.id, sqltxt)
	rows, err := w.db.Query(sqltxt)
	log.Printf("worker:%d, query:%d, time:%v\n", w.id, queryId, time.Since(startTime).Milliseconds())
	if err != nil {
		return tables, err
	}
	defer rows.Close()
	tables, err = Sqlrows2Maps(rows)
	if err != nil {
		return tables, err
	}
	return tables, nil
}

// 查询
func (w *Worker) QueryString(sqltxt string) ([]map[string]interface{}, error) {
	var tables []map[string]interface{}
	startTime := time.Now()
	log.Printf("worker:%d, query:%v\n", w.id, sqltxt)
	rows, err := w.db.Query(sqltxt)
	log.Printf("worker:%d, query:%v, time:%v\n", w.id, sqltxt, time.Since(startTime).Milliseconds())
	if err != nil {
		return tables, err
	}
	defer rows.Close()
	tables, err = Sqlrows2Maps(rows)
	if err != nil {
		return tables, err
	}
	return tables, nil
}

// Sqlrows2Maps sql查询结果rows转为maps
func Sqlrows2Maps(rws *sql.Rows) ([]map[string]interface{}, error) {
	rowMaps := make([]map[string]interface{}, 0)
	var columns []string
	columns, err := rws.Columns()
	if err != nil {
		return rowMaps, err
	}
	values := make([]sql.RawBytes, len(columns))
	scans := make([]interface{}, len(columns))
	for i := range values {
		scans[i] = &values[i]
	}
	for rws.Next() {
		_ = rws.Scan(scans...)
		each := map[string]interface{}{}
		for i, col := range values {
			each[columns[i]] = string(col)
		}

		rowMaps = append(rowMaps, each)
	}
	return rowMaps, nil
}
