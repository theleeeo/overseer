package main

import (
	"context"
	"overseer/runner"
)

func main() {
	runner := runner.New(&runner.Config{
		DbConnString: "postgres://user:pass@localhost:5432/overseer",
	})
	if err := runner.Run(context.Background()); err != nil {
		panic(err)
	}
}
