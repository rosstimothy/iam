package ports

import (
	"net/http"

	"github.com/go-chi/chi"
)

func NewHandlerForMux(server *HttpServer, r chi.Router) http.Handler {
	r.Get("/role/named", server.RoleByName())
	r.Get("/role/permissions", server.RolesWithPermissions())

	return r
}
