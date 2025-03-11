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

func TestAdminUser(t *testing.T) {
	adminUserObject := "organization.user"
	adminUserID := "user-id"
	adminUserName := "user-name"
	adminUserEmail := "user-email"
	adminUserRole := "member"
	adminUserAddedAt := int64(1711471533)

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/organization/users",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.AdminUserList{
					Object: "list",
					User: []openai.AdminUser{
						{
							Object:  adminUserObject,
							ID:      adminUserID,
							Name:    adminUserName,
							Email:   adminUserEmail,
							Role:    adminUserRole,
							AddedAt: adminUserAddedAt,
						},
					},
					FirstID: "first_id",
					LastID:  "last_id",
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/users/"+adminUserID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminUser{
					Object:  adminUserObject,
					ID:      adminUserID,
					Name:    adminUserName,
					Email:   adminUserEmail,
					Role:    adminUserRole,
					AddedAt: adminUserAddedAt,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminUser{
					Object:  adminUserObject,
					ID:      adminUserID,
					Name:    adminUserName,
					Email:   adminUserEmail,
					Role:    adminUserRole,
					AddedAt: adminUserAddedAt,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodDelete:
				resBytes, _ := json.Marshal(openai.AdminUserDeleteResponse{
					ID:      adminUserID,
					Object:  adminUserObject,
					Deleted: true,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("ListAdminUsers", func(t *testing.T) {
		response, err := client.ListAdminUsers(ctx, nil, nil, nil)
		checks.NoError(t, err, "AdminListUsers error")

		if len(response.User) != 1 {
			t.Errorf("AdminListUsers returned %d users, want 1", len(response.User))
		}

		if response.User[0].ID != adminUserID {
			t.Errorf("AdminListUsers returned user ID %s, want %s", response.User[0].ID, adminUserID)
		}
	})

	t.Run("ModifyAdminUser", func(t *testing.T) {
		response, err := client.ModifyAdminUser(ctx, adminUserID, adminUserRole)
		checks.NoError(t, err, "ModifyAdminUser error")

		if response.ID != adminUserID {
			t.Errorf("ModifyAdminUser returned user ID %s, want %s", response.ID, adminUserID)
		}
	})

	t.Run("RetrieveAdminUser", func(t *testing.T) {
		response, err := client.RetrieveAdminUser(ctx, adminUserID)
		checks.NoError(t, err, "GetAdminUser error")

		if response.ID != adminUserID {
			t.Errorf("GetAdminUser returned user ID %s, want %s", response.ID, adminUserID)
		}
	})

	t.Run("DeleteAdminUser", func(t *testing.T) {
		response, err := client.DeleteAdminUser(ctx, adminUserID)
		checks.NoError(t, err, "DeleteAdminUser error")

		if !response.Deleted {
			t.Errorf("DeleteAdminUser returned user not deleted, want deleted")
		}
	})
}
