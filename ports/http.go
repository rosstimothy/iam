package ports

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rosstimothy/iam/app"
	"github.com/rosstimothy/iam/app/query"
)

type HttpServer struct {
	app *app.Application
}

func NewHttpServer(app *app.Application) *HttpServer {
	if app == nil {
		panic("nil app")
	}

	return &HttpServer{app: app}
}

func (h *HttpServer) RolesWithPermissions() http.HandlerFunc {
	type request struct {
		Permissions []string `json:"permissions"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		cmd := query.RolesWithPermissions{Permissions: req.Permissions}
		roles, err := h.app.Queries.RolesWithPermissions.Handle(r.Context(), cmd)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(roles)
	}
}

func (h *HttpServer) RoleByName() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		cmd := query.RoleByName{Role: req.Name}
		role, err := h.app.Queries.RoleByName.Handle(r.Context(), cmd)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(role)
	}
}
