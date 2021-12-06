package sql_test

import (
	"github.com/Reisender/pipe/extras/sql"
	"github.com/Reisender/pipe/line"
	"github.com/Reisender/pipe/message"
)

func ExampleExec() {
	line.New().SetP(func(out chan<- interface{}, errs chan<- error) {

		// using a string directly as the query
		out <- "INSERT INTO foo (name) VALUES ('bar')"

		// using message.Query
		out <- message.Query{SQL: "INSERT INTO foo (name) VALUES ('bar')"}

		// using message.Query with args (using sqlx under the hood for arg matching)
		query := message.Query{SQL: "INSERT INTO foo (name) VALUES (?)"}
		query.Args = append(query.Args, "bar")
		out <- query

		// using delta messages
		record := message.NewRecordFromMSI(map[string]interface{}{"name": "bar"})
		delta := message.InsertDelta{Table: "foo", Record: record}
		out <- delta

	}).Add(
		sql.Exec{Driver: "postgres", DSN: "postgres://user:pass@localhost/dbname?sslmode=verify-full"}.T,
	).Run()
}
