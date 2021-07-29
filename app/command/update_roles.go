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

type UpdateRoles struct{}

type UpdateRolesHandler struct {
	client *ent.Client
}

func NewUpdateRolesHandler(client *ent.Client) *UpdateRolesHandler {
	if client == nil {
		panic("nil client")
	}

	return &UpdateRolesHandler{client: client}
}

var (
	notFound *ent.NotFoundError
)

func (l *UpdateRolesHandler) Handle(ctx context.Context, cmd UpdateRoles) (err error) {
	defer func() {
		if err != nil {
			fmt.Println("failed to update roles")
			return
		}

		fmt.Println("completed updating roles")
	}()

	fmt.Println("fetching roles")
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
		if iamRole.Deleted {
			_, err := tx.Role.Delete().Where(role.Name(iamRole.Name)).Exec(ctx)
			if err != nil {
				fmt.Println(err.Error())
			}
			continue
		}

		r, err := tx.Role.Query().Where(role.Name(iamRole.Name)).WithPermissions().Only(ctx)
		if err != nil {
			if !errors.As(err, &notFound) {
				return err
			}

			fmt.Printf("creating role %s\n", iamRole.Name)
			if err := createRole(ctx, tx, iamRole); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("updating role %s\n", iamRole.Name)
		if err := updateRole(ctx, tx, r, iamRole); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func newPermissions(ctx context.Context, tx *ent.Tx, iamRole *adminpb.Role) ([]*ent.Permission, error) {
	permissions, err := tx.Permission.
		Query().
		Where(permission.NameIn(iamRole.IncludedPermissions...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var newPermissions []*ent.PermissionCreate

	for i, includedPermission := range iamRole.IncludedPermissions {
		idx := -1
		for _, existingPermission := range permissions {
			if existingPermission.Name == includedPermission {
				idx = i
				break
			}
		}

		if idx < 0 {
			fmt.Printf("creating permission %s\n", includedPermission)
			newPermissions = append(newPermissions, tx.Permission.Create().SetName(includedPermission))
		}
	}

	if len(newPermissions) > 0 {
		p, err := tx.Permission.CreateBulk(newPermissions...).Save(ctx)
		if err != nil {
			return nil, err
		}

		return append(permissions, p...), nil
	}

	return permissions, nil
}

func createRole(ctx context.Context, tx *ent.Tx, iamRole *adminpb.Role) error {
	perms, err := newPermissions(ctx, tx, iamRole)
	if err != nil {
		return err
	}

	_, err = tx.Role.Create().
		SetName(iamRole.Name).
		SetTitle(iamRole.Title).
		SetDescription(iamRole.Description).
		SetEtag(iamRole.Etag).
		SetStage(int(iamRole.Stage)).
		AddPermissions(perms...).
		Save(ctx)

	return err
}

func removedPermissions(ctx context.Context, tx *ent.Tx, r *ent.Role, iamRole *adminpb.Role) ([]*ent.Permission, error) {

	var removedPermissions []*ent.Permission

	for i, existingPermission := range r.Edges.Permissions {
		idx := -1
		for _, includedPermission := range iamRole.IncludedPermissions {
			if existingPermission.Name == includedPermission {
				idx = i
				break
			}
		}

		if idx < 0 {
			fmt.Printf("removing permission %s\n", existingPermission.Name)
			removedPermissions = append(removedPermissions, existingPermission)
		}
	}

	return removedPermissions, nil
}

func updateRole(ctx context.Context, tx *ent.Tx, r *ent.Role, iamRole *adminpb.Role) error {

	newPerms, err := newPermissions(ctx, tx, iamRole)
	if err != nil {
		return err
	}

	removedPerms, err := removedPermissions(ctx, tx, r, iamRole)
	if err != nil {
		return err
	}

	_, err = r.Update().
		SetTitle(iamRole.Title).
		SetDescription(iamRole.Description).
		SetEtag(iamRole.Etag).
		SetStage(int(iamRole.Stage)).
		RemovePermissions(removedPerms...).
		AddPermissions(newPerms...).
		Save(ctx)

	return err
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
			ShowDeleted: true,
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
