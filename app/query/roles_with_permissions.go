package query

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/rosstimothy/iam/ent"
	"github.com/rosstimothy/iam/ent/permission"
)

type RolesWithPermissions struct {
	Permissions []string
}

type RolesWithPermissionsHandler struct {
	client *ent.Client
}

func NewRolesWithPermissionsHandler(client *ent.Client) *RolesWithPermissionsHandler {
	if client == nil {
		panic("nil client")
	}

	return &RolesWithPermissionsHandler{client: client}
}

func (l *RolesWithPermissionsHandler) Handle(ctx context.Context, cmd RolesWithPermissions) (_ []Role, err error) {
	fmt.Printf("looking for roles with permissions %s\n", cmd.Permissions)

	defer func() {
		if err != nil {
			fmt.Printf("failed to find roles with permissions %s\n", cmd.Permissions)
			return
		}

		fmt.Printf("succesfully found roles with permissions %s\n", cmd.Permissions)
	}()

	roles, err := l.client.Role.
		Query().
		QueryPermissions().
		Where(permission.NameIn(cmd.Permissions...)).
		QueryRoles().
		WithPermissions().
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles for permissions %v", &cmd.Permissions)
	}

	r := make([]Role, len(roles))

	for i, rr := range roles {
		r[i] = Role{
			Name:        rr.Name,
			Title:       rr.Title,
			Description: rr.Description,
			Permissions: nil,
			Stage:       rr.Stage,
			Etag:        hex.EncodeToString(rr.Etag),
		}

		r[i].Permissions = make([]string, len(rr.Edges.Permissions))
		for j, p := range rr.Edges.Permissions {
			r[i].Permissions[j] = p.Name
		}
	}

	return r, nil
}
