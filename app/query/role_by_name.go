package query

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/rosstimothy/iam/ent"
	"github.com/rosstimothy/iam/ent/role"
)

type RoleByName struct {
	Role string
}

type RoleByNameHandler struct {
	client *ent.Client
}

func NewRoleByNameHandler(client *ent.Client) *RoleByNameHandler {
	if client == nil {
		panic("nil client")
	}

	return &RoleByNameHandler{client: client}
}

func (l *RoleByNameHandler) Handle(ctx context.Context, cmd RoleByName) (_ *Role, err error) {
	fmt.Printf("looking for roles named %s\n", cmd.Role)
	defer func() {
		if err != nil {
			fmt.Printf("failed to find roles named %s\n", cmd.Role)
			return
		}

		fmt.Printf("succesfully found roles named %s\n", cmd.Role)
	}()

	entRole, err := l.client.Role.
		Query().
		Where(role.Name(cmd.Role)).
		WithPermissions().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	r := &Role{
		Name:        entRole.Name,
		Title:       entRole.Title,
		Description: entRole.Description,
		Permissions: nil,
		Stage:       entRole.Stage,
		Etag:        hex.EncodeToString(entRole.Etag),
	}

	r.Permissions = make([]string, len(entRole.Edges.Permissions))
	for i, p := range entRole.Edges.Permissions {
		r.Permissions[i] = p.Name
	}

	return r, nil
}
