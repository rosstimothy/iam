package app

import (
	"github.com/rosstimothy/iam/app/command"
	"github.com/rosstimothy/iam/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	UpdateRoles *command.UpdateRolesHandler
}

type Queries struct {
	RolesWithPermissions *query.RolesWithPermissionsHandler
	RoleByName           *query.RoleByNameHandler
}
