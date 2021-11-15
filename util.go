package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/openpgp/errors"
)

type Worker struct {
	db      *sql.DB
	id      int
	queries *[]string
	tm      *threadsPool
}

func (w *Worker) Run(ctx context.Context, resultCh chan string) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %d quit", w.id)
			return
		default:
			k := w.tm.Alloc()
			if k != -1 {
				log.Printf("worker %d executing", w.id)
				queryId := rand.Int() % len(*w.queries)
				_, executeTime, err := w.QueryToId(queryId)
				if err != nil {
					log.Printf("Query error: %s", err)
					resultCh <- fmt.Sprintf("%v, worker-%v, %v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), w.id, queryId, executeTime.Milliseconds(), err.Error())
				} else {
					resultCh <- fmt.Sprintf("%v, worker-%v, %v, %v\n", time.Now().Format("2006-01-02 15:04:05"), w.id, queryId, executeTime.Milliseconds())
				}
				w.tm.Free(k)
			}
		}
		time.Sleep(QueryInterval)
	}
}

func (w *Worker) Init(dsn string, id int) (err error) {
	w.queries = &[]string{
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
		// `
		// with graph as (select node1id from match (:person)-[:knows]->(:person) as graph where node1id is not null)
		// select feedback.personid, feedback.feedback
		// from graph
		// join hbase.default.feedback on feedback.personid = node1id
		// join mongodb.unibench.orders on orders.personid = node1id
		// where feedback like '%1.0%' and orderdate = '2018-07-07' and orderline[1].productId = '1380'
		// `,

		// `
		// with pids as (select PersonId as pid, SUM(TotalPrice) as sum
		//     from mongodb.unibench.orders
		//     Group by PersonId
		//     order by sum desc
		//     limit 2)
		// select node1id
		// from (select * from (select node0id, node1id, node2id from match (:person)-[:knows]->(:person)<-[:knows]-(:person) as graph) where node0id is not null)
		// where node0id in (select pid from pids) and node2id in (select pid from pids)
		// `,
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
		// `
		// WITH salesone as (SELECT feedback.asin, count(totalprice) as count
		// 		FROM mongodb.unibench.orders
		// 		join hive.unibench.person on person.id = orders.personid
		// 		join hbase.default.feedback on feedback.personid = orders.personid
		// 		join mongodb.unibench.product on product.asin = feedback.asin
		// 		where mongodb.unibench.product.brand = 1 and orders.orderdate like '%2019%'
		// 		Group by feedback.asin
		// 		Order by count DESC limit 10),
		// salestwo as (SELECT feedback.asin, count(totalprice) as count
		// 		FROM mongodb.unibench.orders
		// 		join hive.unibench.person on person.id = orders.personid
		// 		join hbase.default.feedback on feedback.personid = orders.personid
		// 		join mongodb.unibench.product on product.asin = feedback.asin
		// 		where mongodb.unibench.product.brand = 1 and orders.orderdate like '%2018%'
		// 		Group by feedback.asin
		// 		Order by count DESC limit 10)
		// SELECT distinct salesone.asin, (salesone.count-salestwo.count) as residual, feedback.feedback
		// FROM salesone
		// INNER JOIN hbase.default.feedback on salesone.asin = hbase.default.feedback.asin
		// INNER JOIN salestwo on salesone.asin=salestwo.asin
		// WHERE salesone.count>salestwo.count
		// `,
		// `
		// WITH graph as (select node1asin FROM MATCH (:post)-[:hastag]->(:tag) as graph where node1asin is not null),
		// 	Totalsales as (SELECT asin,count(orderline) as ol
		// 					FROM mongodb.unibench.orders o
		// 					join hbase.default.feedback f on f.personid = o.personid
		// 					GROUP BY asin 
		// 					ORDER BY ol DESC)
		// SELECT Totalsales.asin,count(Totalsales.asin) as ts
		// FROM graph
		// INNER JOIN Totalsales on Totalsales.asin = graph.node1asin
		// GROUP BY Totalsales.asin 
		// ORDER BY ts DESC
		// `,
		// `
		// WITH brandList as (select id from hive.unibench.vendor where country='China'),
		// 	TopCompany as (SELECT c.gender, count(c.id), p.brand FROM mongodb.unibench.product p
		// 				JOIN brandList bl on bl.id=cast(p.brand as varchar)
		// 				join hbase.default.feedback f on f.asin= p.asin
		// 				JOIN hive.unibench.person c on c.id=f.personid
		// 				GROUP BY c.gender, p.brand
		// 				Order by count(c.id) DESC, p.brand DESC
		// 				limit 3)
		// 				SELECT distinct tc.*, node0content,node0creationDate
		// 				FROM match (:post)-[:hastag]->(:tag) as graph
		// 				CROSS JOIN TopCompany tc
		// 				order by node0creationDate DESC
		// `,
		// `
		// with dfn as (select node0id as pid, count(node1creationDate) as pc
		// from match (:person)-[:hascreated]->(:post) as graph
		// where pc > '2012-10'
		// group by pid
		// order by pc desc
		// limit 10)
		// SELECT o.personid as Active_person, max(o.orderdate) as Recency, count(o.orderdate) as Frequency,sum(o.totalprice) as Monetary
		// from mongodb.unibench.orders o
		// inner join dfn on o.person = node0id
		// group by o.personid
		// `,
	}
	w.id = id
	err = w.Connect(dsn)
	log.Printf("worker %d init", w.id)
	return
}

func (w *Worker) Connect(dsn string) (err error) {
	w.db, err = sql.Open("presto", dsn)
	if err != nil {
		log.Printf("worker %d failed to open dsn %s", w.id, dsn)
		return
	}
	err = w.db.Ping()
	if err != nil {
		log.Printf("worker %d failed to ping db %s", w.id, dsn)
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
func (w *Worker) QueryToId(queryId int) (tables []map[string]interface{}, executeTime time.Duration, err error) {
	if queryId < 0 || queryId >= len(*w.queries) {
		err = errors.UnsupportedError("query id exceed")
		return
	}
	startTime := time.Now()
	sqltxt := (*w.queries)[queryId]
	log.Printf("worker:%d, query:%v\n", w.id, sqltxt)
	rows, err := w.db.Query(sqltxt)
	executeTime = time.Since(startTime)
	log.Printf("worker:%d, query:%d, time:%v\n", w.id, queryId, executeTime.Milliseconds())
	if err != nil {
		return
	}
	defer rows.Close()
	tables, err = Sqlrows2Maps(rows)
	if err != nil {
		return
	}
	return
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
