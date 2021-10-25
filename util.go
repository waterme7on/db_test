package main

import (
	"database/sql"
	"log"
)

const dsn = "http://root@10.77.50.201:31314"

func connect() (db *sql.DB, err error) {
	db, err = sql.Open("presto", dsn)
	if err != nil {
		return db, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}
	return db, nil
}

// 查询
func Query(db *sql.DB, sqltxt string) ([]map[string]interface{}, error) {
	var tables []map[string]interface{}
	rows, err := db.Query(sqltxt)
	log.Printf("%v %v\n", rows, err)
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
