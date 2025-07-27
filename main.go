package main

import (
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			NewConfig,
			NewLogger,
			NewPgxPool,
			NewHealthSrvc,
			NewHealthHandler,
			NewServer,
		),
		fx.Invoke((*Server).Start),
	).Run()
}
