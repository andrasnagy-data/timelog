// Timelog
// @title Timelog
// @version 0.1.0
// @description A REST API for managing activity log
// @termsOfService http://swagger.io/terms/

// @contact.name András
// @contact.email andrasna@proton.me

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

package main

import (
	"github.com/andrasnagy-data/timelog/internal/components/activity"
	"github.com/andrasnagy-data/timelog/internal/components/auth"
	"github.com/andrasnagy-data/timelog/internal/server"
	"github.com/andrasnagy-data/timelog/internal/shared/config"
	"github.com/andrasnagy-data/timelog/internal/shared/database"
	"github.com/andrasnagy-data/timelog/internal/shared/logging"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			logging.NewLogger,
			database.NewPgxPool,
			server.NewServer,
			server.NewHealthSrvc,
			server.NewHealthHandler,
			activity.NewRepo,
			activity.NewService,
			fx.Annotate(activity.NewRouter, fx.ResultTags(`name:"activityRouter"`)),
			auth.NewAuthService,
			fx.Annotate(auth.NewRouter, fx.ResultTags(`name:"authRouter"`)),
		),
		fx.Invoke(server.Register),
	).Run()
}
