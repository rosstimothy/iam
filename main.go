package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/run"

	"github.com/rosstimothy/iam/app"
	"github.com/rosstimothy/iam/app/command"
	"github.com/rosstimothy/iam/app/query"
	"github.com/rosstimothy/iam/ent"
	"github.com/rosstimothy/iam/ports"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		fmt.Printf("failed opening connection to sqlite: %v\n", err)
		return
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		fmt.Printf("failed creating schema resources: %v\n", err)
		return
	}

	application := &app.Application{
		Commands: app.Commands{
			UpdateRoles: command.NewUpdateRolesHandler(client),
		},
		Queries: app.Queries{
			RolesWithPermissions: query.NewRolesWithPermissionsHandler(client),
			RoleByName:           query.NewRoleByNameHandler(client),
		},
	}

	apiRouter := chi.NewRouter()

	apiRouter.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
		middleware.SetHeader("Content-Type", "application/json; charset=utf-8"),
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.NoCache,
	)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/v1", ports.NewHandlerForMux(ports.NewHttpServer(application), apiRouter))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: rootRouter,
	}

	var g run.Group
	{
		g.Add(func() error {
			fmt.Printf("Server Started on %s\n", ":8080")

			return srv.ListenAndServe()
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				fmt.Printf("Failed to shutdown server: %v\n", err)
			}
		})
	}
	{
		g.Add(func() error {
			update := func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				return application.Commands.UpdateRoles.Handle(ctx, command.UpdateRoles{})
			}

			for {
				if err := update(); err != nil {
					return err
				}
				time.Sleep(time.Minute * 5)
			}

		}, func(err error) {
			fmt.Printf("failed to update roles: %v\n", err)
		})
	}

	if err := g.Run(); err != nil {
		fmt.Printf("The run group was terminated: %v\n", err)
	}
}
