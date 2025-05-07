package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestAdminProjectUser(t *testing.T) {
	adminProjectID := "project-abc-123"
	adminProjectUserObject := "organization.project.user"
	adminProjectUserID := "user-abc-123"
	adminProjectUserName := "User Name"
	adminProjectUserEmail := "test@here.com"
	adminProjectUserRole := "owner"
	adminProjectUserAddedAt := int64(1711471533)

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		fmt.Sprintf("/v1/organization/projects/%s/users", adminProjectID),
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminProjectUserList{
					Object: "list",
					Data: []openai.AdminProjectUser{
						{
							ID:      adminProjectUserID,
							Object:  adminProjectUserObject,
							Name:    adminProjectUserName,
							Email:   adminProjectUserEmail,
							Role:    adminProjectUserRole,
							AddedAt: adminProjectUserAddedAt,
						},
					},
					FirstID: "first_id",
					LastID:  "last_id",
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminProjectUser{
					ID:      adminProjectUserID,
					Object:  adminProjectUserObject,
					Email:   adminProjectUserEmail,
					Role:    adminProjectUserRole,
					AddedAt: adminProjectUserAddedAt,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		fmt.Sprintf("/v1/organization/projects/%s/users/%s", adminProjectID, adminProjectUserID),
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminProjectUser{
					ID:      adminProjectUserID,
					Object:  adminProjectUserObject,
					Name:    adminProjectUserName,
					Email:   adminProjectUserEmail,
					Role:    adminProjectUserRole,
					AddedAt: adminProjectUserAddedAt,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminProjectUser{
					ID:      adminProjectUserID,
					Object:  adminProjectUserObject,
					Email:   adminProjectUserEmail,
					Role:    adminProjectUserRole,
					AddedAt: adminProjectUserAddedAt,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodDelete:
				resBytes, _ := json.Marshal(openai.AdminProjectDeleteResponse{
					Object:  "delete",
					ID:      adminProjectUserID,
					Deleted: true,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("ListAdminProjectUsers", func(t *testing.T) {
		adminProjectUsers, err := client.ListAdminProjectUsers(ctx, adminProjectID, nil, nil)
		checks.NoError(t, err, "ListAdminProjectUsers error")

		if len(adminProjectUsers.Data) != 1 {
			t.Errorf("expected 1 project user, got %d", len(adminProjectUsers.Data))
		}

		adminProjectUser := adminProjectUsers.Data[0]
		if adminProjectUser.ID != adminProjectUserID {
			t.Errorf("expected user ID %s, got %s", adminProjectUserID, adminProjectUser.ID)
		}
	})

	t.Run("ListAdminProjectUsersFilter", func(t *testing.T) {
		limit := 5
		after := "after_id"

		adminProjectUsers, err := client.ListAdminProjectUsers(ctx, adminProjectID, &limit, &after)
		checks.NoError(t, err, "ListAdminProjectUsers error")

		if len(adminProjectUsers.Data) != 1 {
			t.Errorf("expected 1 project user, got %d", len(adminProjectUsers.Data))
		}

		adminProjectUser := adminProjectUsers.Data[0]
		if adminProjectUser.ID != adminProjectUserID {
			t.Errorf("expected user ID %s, got %s", adminProjectUserID, adminProjectUser.ID)
		}
	})

	t.Run("CreateAdminProjectUser", func(t *testing.T) {
		adminProjectUser, err := client.CreateAdminProjectUser(
			ctx,
			adminProjectID,
			adminProjectUserEmail,
			adminProjectUserRole,
		)
		checks.NoError(t, err, "CreateAdminProjectUser error")

		if adminProjectUser.ID != adminProjectUserID {
			t.Errorf("expected user ID %s, got %s", adminProjectUserID, adminProjectUser.ID)
		}
	})

	t.Run("RetrieveAdminProjectUser", func(t *testing.T) {
		adminProjectUser, err := client.RetrieveAdminProjectUser(ctx, adminProjectID, adminProjectUserID)
		checks.NoError(t, err, "RetrieveAdminProjectUser error")

		if adminProjectUser.ID != adminProjectUserID {
			t.Errorf("expected user ID %s, got %s", adminProjectUserID, adminProjectUser.ID)
		}
	})

	t.Run("ModifyAdminProjectUser", func(t *testing.T) {
		adminProjectUser, err := client.ModifyAdminProjectUser(ctx, adminProjectID, adminProjectUserID, adminProjectUserRole)
		checks.NoError(t, err, "ModifyAdminProjectUser error")

		if adminProjectUser.ID != adminProjectUserID {
			t.Errorf("expected user ID %s, got %s", adminProjectUserID, adminProjectUser.ID)
		}
	})

	t.Run("DeleteAdminProjectUser", func(t *testing.T) {
		adminProjectUser, err := client.DeleteAdminProjectUser(ctx, adminProjectID, adminProjectUserID)
		checks.NoError(t, err, "DeleteAdminProjectUser error")

		if !adminProjectUser.Deleted {
			t.Errorf("expected user to be deleted, got %t", adminProjectUser.Deleted)
		}
	})
}
