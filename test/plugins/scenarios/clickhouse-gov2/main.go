// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"strconv"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	_ "github.com/apache/skywalking-go"
)

type testFunc func(ctx context.Context) error

var (
	url  = "clickhouse-server:8123"
	conn driver.Conn
)

func main() {
	var err error
	conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{url},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		Protocol: clickhouse.HTTP,
	})
	if err != nil {
		log.Fatalf("connect to clickhouse error: %v \n", err)
		return
	}
	http.HandleFunc("/execute", executeHandler)
	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	_ = http.ListenAndServe(":8080", nil)
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	testCases := []struct {
		name string
		fn   testFunc
	}{
		{"createOp", createOp},
		{"insertOp", insertOp},
		{"selectOp", selectOp},
		{"dropOp", dropOp},
	}
	ctx := context.Background()
	for _, test := range testCases {
		log.Printf("excute test case %s", test.name)
		if err := test.fn(ctx); err != nil {
			log.Fatalf("test case %s failed: %v", test.name, err)
		}
	}
}

func createOp(ctx context.Context) error {
	err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS sw_ck (
			  Col1 Array(String)
			, Col2 Array(Array(Int64))
		) Engine = Memory
	`)
	if err != nil {
		return err
	}
	return nil
}

func insertOp(ctx context.Context) error {
	batch, err := conn.PrepareBatch(ctx, "INSERT INTO sw_ck")
	if err != nil {
		return err
	}
	var i int64
	for i = 0; i < 10; i++ {
		err := batch.Append(
			[]string{strconv.Itoa(int(i)), strconv.Itoa(int(i + 1)), strconv.Itoa(int(i + 2)), strconv.Itoa(int(i + 3))},
			[][]int64{{i, i + 1}, {i + 2, i + 3}, {i + 4, i + 5}},
		)
		if err != nil {
			return err
		}
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}

func selectOp(ctx context.Context) error {
	var (
		col1 []string
		col2 [][]int64
	)
	rows, err := conn.Query(ctx, "SELECT * FROM sw_ck")
	if err != nil {
		return err
	}
	for rows.Next() {
		if err := rows.Scan(&col1, &col2); err != nil {
			return err
		}
		fmt.Printf("row: col1=%v, col2=%v\n", col1, col2)
	}
	rows.Close()
	return rows.Err()
}

func dropOp(ctx context.Context) error {
	return conn.Exec(ctx, "DROP TABLE sw_ck")
}
