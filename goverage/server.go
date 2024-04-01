package main

import (
	"context"
	"database/sql"
	"embed"
	"goverage/internal/config"
	apiv1 "goverage/routers/api/v1"
	"goverage/routers/public"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed sql/versions/*.sql
var migrations embed.FS

func runMigrations() {
	db, err := sql.Open("pgx", config.Config.DBConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		db.Close()
		log.Fatal().Err(err).Msg("Failed to set goose dialect")
	}

	if err := goose.Up(db, "sql/versions"); err != nil {
		db.Close()
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	db.Close()
}

func main() {
	ctx := context.Background()

	config.LoadConfig()

	runMigrations()

	conn, err := pgx.Connect(ctx, config.Config.DBConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer conn.Close(ctx)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	apiV1Router := apiv1.NewAPIV1Router(e, conn)
	apiV1Router.Register()

	publicRouter := public.NewPublicRouter(e, conn)
	publicRouter.Register()

	e.GET("/_live", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.GET("/_ready", func(c echo.Context) error {
		if _, err := conn.Exec(c.Request().Context(), "SELECT 1"); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return c.NoContent(http.StatusNoContent)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
