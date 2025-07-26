package main

import (
	"go.uber.org/fx"
)

func main() {
	fx.New(
		// Provide dependencies
		fx.Provide(
			NewConfig,
			NewLogger,
			NewHealthSrvc,
			NewHealthHandler,
			NewServer,
		),
		// Invoke lifecycle hooks
		fx.Invoke((*Server).Start),
	).Run()
}
