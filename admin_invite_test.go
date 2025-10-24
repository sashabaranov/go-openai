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

func TestAdminInvite(t *testing.T) {
	adminInviteObject := "organization.invite"
	adminInviteID := "invite-abc-123"
	adminInviteEmail := "invite@openai.com"
	adminInviteRole := "owner"
	adminInviteStatus := "pending"

	adminInviteInvitedAt := int64(1711471533)
	adminInviteExpiresAt := int64(1711471533)
	adminInviteAcceptedAt := int64(1711471533)
	adminInviteProjects := []openai.AdminInviteProject{
		{
			ID:   "project-id",
			Role: "owner",
		},
	}

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/organization/invites",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminInviteList{
					Object: "list",
					AdminInvites: []openai.AdminInvite{
						{
							Object:     adminInviteObject,
							ID:         adminInviteID,
							Email:      adminInviteEmail,
							Role:       adminInviteRole,
							Status:     adminInviteStatus,
							InvitedAt:  adminInviteInvitedAt,
							ExpiresAt:  adminInviteExpiresAt,
							AcceptedAt: adminInviteAcceptedAt,
							Projects:   adminInviteProjects,
						},
					},
					FirstID: "first_id",
					LastID:  "last_id",
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminInvite{
					Object:     adminInviteObject,
					ID:         adminInviteID,
					Email:      adminInviteEmail,
					Role:       adminInviteRole,
					Status:     adminInviteStatus,
					InvitedAt:  adminInviteInvitedAt,
					ExpiresAt:  adminInviteExpiresAt,
					AcceptedAt: adminInviteAcceptedAt,
					Projects:   adminInviteProjects,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/invites/"+adminInviteID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodDelete:
				resBytes, _ := json.Marshal(openai.AdminInviteDeleteResponse{
					ID:      adminInviteID,
					Object:  adminInviteObject,
					Deleted: true,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminInvite{
					Object:     adminInviteObject,
					ID:         adminInviteID,
					Email:      adminInviteEmail,
					Role:       adminInviteRole,
					Status:     adminInviteStatus,
					InvitedAt:  adminInviteInvitedAt,
					ExpiresAt:  adminInviteExpiresAt,
					AcceptedAt: adminInviteAcceptedAt,
					Projects:   adminInviteProjects,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("ListAdminInvites", func(t *testing.T) {
		adminInvites, err := client.ListAdminInvites(ctx, nil, nil)
		checks.NoError(t, err, "ListAdminInvites error")

		if len(adminInvites.AdminInvites) != 1 {
			t.Fatalf("expected 1 admin invite, got %d", len(adminInvites.AdminInvites))
		}

		if adminInvites.AdminInvites[0].ID != adminInviteID {
			t.Errorf("expected admin invite ID %s, got %s", adminInviteID, adminInvites.AdminInvites[0].ID)
		}
	})

	t.Run("ListAdminInvitesFilter", func(t *testing.T) {
		limit := 10
		after := "after-id"

		adminInvites, err := client.ListAdminInvites(ctx, &limit, &after)
		checks.NoError(t, err, "ListAdminInvites error")

		if len(adminInvites.AdminInvites) != 1 {
			t.Fatalf("expected 1 admin invite, got %d", len(adminInvites.AdminInvites))
		}

		if adminInvites.AdminInvites[0].ID != adminInviteID {
			t.Errorf("expected admin invite ID %s, got %s", adminInviteID, adminInvites.AdminInvites[0].ID)
		}
	})

	t.Run("CreateAdminInvite", func(t *testing.T) {
		adminInvite, err := client.CreateAdminInvite(ctx, adminInviteEmail, adminInviteRole, &adminInviteProjects)
		checks.NoError(t, err, "CreateAdminInvite error")

		if adminInvite.ID != adminInviteID {
			t.Errorf("expected admin invite ID %s, got %s", adminInviteID, adminInvite.ID)
		}
	})

	t.Run("RetrieveAdminInvite", func(t *testing.T) {
		adminInvite, err := client.RetrieveAdminInvite(ctx, adminInviteID)
		checks.NoError(t, err, "RetrieveAdminInvite error")

		if adminInvite.ID != adminInviteID {
			t.Errorf("expected admin invite ID %s, got %s", adminInviteID, adminInvite.ID)
		}
	})

	t.Run("DeleteAdminInvite", func(t *testing.T) {
		adminInviteDeleteResponse, err := client.DeleteAdminInvite(ctx, adminInviteID)
		checks.NoError(t, err, "DeleteAdminInvite error")

		if !adminInviteDeleteResponse.Deleted {
			t.Errorf("expected admin invite to be deleted, got not deleted")
		}
	})
}
