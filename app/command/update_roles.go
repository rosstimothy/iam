package command

import (
	"context"
	"errors"
	"fmt"

	admin "cloud.google.com/go/iam/admin/apiv1"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"

	"github.com/rosstimothy/iam/ent"
	"github.com/rosstimothy/iam/ent/permission"
	"github.com/rosstimothy/iam/ent/role"
)

type UpdateRoles struct {
}

type UpdateRolesHandler struct {
	client *ent.Client
}

func NewUpdateRolesHandler(client *ent.Client) *UpdateRolesHandler {
	if client == nil {
		panic("nil client")
	}

	return &UpdateRolesHandler{client: client}
}

func (l *UpdateRolesHandler) Handle(ctx context.Context, cmd UpdateRoles) (err error) {
	fmt.Println("fetching roles")
	defer func() {
		if err != nil {
			fmt.Println("failed to update roles")
			return
		}

		fmt.Println("completed updating roles")
	}()

	roles, err := fetchRoles(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("found %d roles\n", len(roles))

	tx, err := l.client.Tx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, iamRole := range roles {
		permissions := make([]*ent.Permission, 0, len(iamRole.IncludedPermissions))
		perms, err := tx.Permission.
			Query().
			Where(permission.NameIn(iamRole.IncludedPermissions...)).
			All(ctx)
		if err != nil {
			return err
		}

		for _, p := range iamRole.IncludedPermissions {
			idx := -1
			for j, pp := range perms {
				if p == pp.Name {
					idx = j
					break
				}
			}

			if idx >= 0 {
				permissions = append(permissions, perms[idx])
				continue
			}

			fmt.Printf("creating permission %s\n", p)
			pp, err := tx.Permission.
				Create().
				SetName(p).
				Save(ctx)
			if err != nil {
				return err
			}

			permissions = append(permissions, pp)
		}

		r, err := tx.Role.
			Query().
			Where(role.Name(iamRole.Name)).
			WithPermissions().
			Only(ctx)

		var notFound *ent.NotFoundError
		if err != nil {
			if !errors.As(err, &notFound) {
				return err
			}

			fmt.Printf("creating role %s\n", iamRole.Name)
			_, err = tx.Role.Create().
				SetName(iamRole.Name).
				SetTitle(iamRole.Title).
				SetDescription(iamRole.Description).
				SetEtag(iamRole.Etag).
				SetStage(int(iamRole.Stage)).
				AddPermissions(permissions...).
				Save(ctx)
			if err != nil {
				return err
			}

			continue
		}

		for _, rp := range r.Edges.Permissions {
			idx := -1
			for j, pp := range permissions {
				if rp.Name == pp.Name {
					idx = j
					break
				}
			}

			if idx < 0 {
				continue
			}

			permissions = append(permissions[:idx], permissions[idx+1:]...)
		}

		fmt.Printf("updating role %s %v\n", iamRole.Name, len(permissions))
		_, err = r.Update().
			SetTitle(iamRole.Title).
			SetDescription(iamRole.Description).
			SetEtag(iamRole.Etag).
			SetStage(int(iamRole.Stage)).
			AddPermissions(permissions...).
			Save(ctx)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}

func fetchRoles(ctx context.Context) ([]*adminpb.Role, error) {
	c, err := admin.NewIamClient(ctx)
	if err != nil {
		return nil, err
	}

	token := ""
	var roles []*adminpb.Role
	for {
		resp, err := c.ListRoles(ctx, &adminpb.ListRolesRequest{
			Parent:      "",
			PageToken:   token,
			View:        adminpb.RoleView_FULL,
			ShowDeleted: false,
		})
		if err != nil {
			return nil, err
		}

		token = resp.NextPageToken
		roles = append(roles, resp.Roles...)

		if token == "" {
			break
		}
	}

	return roles, nil
}
