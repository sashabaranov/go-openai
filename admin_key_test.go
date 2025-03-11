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

func TestAdminKey(t *testing.T) {
	adminKeyObject := "oranization.admin_api_key"
	adminKeyID := "test_key_id"
	adminKeyName := "test_key_name"
	adminKeyRedactedValue := "test_key_redacted_value"
	adminKeyCreatedAt := int64(1711471533)

	adminKeyOwnerType := "service_account"
	adminKeyOwnerObject := "organization.service_account"
	adminKeyOwnerID := "test_owner_id"
	adminKeyOwnerName := "test_owner_name"
	adminKeyOwnerRole := "member"
	adminKeyOwnerCreatedAt := int64(1711471533)

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/organization/admin_api_keys",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminKeyList{
					Object: "list",
					AdminKeys: []openai.AdminKey{
						{
							Object:        adminKeyObject,
							ID:            adminKeyID,
							Name:          adminKeyName,
							RedactedValue: adminKeyRedactedValue,
							CreatedAt:     adminKeyCreatedAt,
							Owner: openai.AdminKeyOwner{
								Type:      adminKeyOwnerType,
								Object:    adminKeyOwnerObject,
								ID:        adminKeyOwnerID,
								Name:      adminKeyOwnerName,
								CreatedAt: adminKeyOwnerCreatedAt,
								Role:      adminKeyOwnerRole,
							},
						},
					},
					FirstID: "first_id",
					LastID:  "last_id",
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminKey{
					Object:        adminKeyObject,
					ID:            adminKeyID,
					Name:          adminKeyName,
					RedactedValue: adminKeyRedactedValue,
					CreatedAt:     adminKeyCreatedAt,
					Owner: openai.AdminKeyOwner{
						Type:      adminKeyOwnerType,
						Object:    adminKeyOwnerObject,
						ID:        adminKeyOwnerID,
						Name:      adminKeyOwnerName,
						CreatedAt: adminKeyOwnerCreatedAt,
						Role:      adminKeyOwnerRole,
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/admin_api_keys/"+adminKeyID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminKey{
					Object:        adminKeyObject,
					ID:            adminKeyID,
					Name:          adminKeyName,
					RedactedValue: adminKeyRedactedValue,
					CreatedAt:     adminKeyCreatedAt,
					Owner: openai.AdminKeyOwner{
						Type:      adminKeyOwnerType,
						Object:    adminKeyOwnerObject,
						ID:        adminKeyOwnerID,
						Name:      adminKeyOwnerName,
						CreatedAt: adminKeyOwnerCreatedAt,
						Role:      adminKeyOwnerRole,
					},
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodDelete:
				resBytes, _ := json.Marshal(openai.AdminKeyDeleteResponse{
					ID:      adminKeyID,
					Object:  adminKeyObject,
					Deleted: true,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	t.Run("ListAdminKeys", func(t *testing.T) {
		adminKeys, err := client.ListAdminKeys(ctx, nil, nil, nil)
		checks.NoError(t, err, "ListAdminKeys error")

		if len(adminKeys.AdminKeys) != 1 {
			t.Fatalf("ListAdminKeys: expected 1 key, got %d", len(adminKeys.AdminKeys))
		}

		if adminKeys.AdminKeys[0].ID != adminKeyID {
			t.Fatalf("ListAdminKeys: expected key ID %s, got %s", adminKeyID, adminKeys.AdminKeys[0].ID)
		}
	})

	t.Run("CreateAdminKey", func(t *testing.T) {
		adminKey, err := client.CreateAdminKey(ctx, adminKeyName)
		checks.NoError(t, err, "CreateAdminKey error")

		if adminKey.ID != adminKeyID {
			t.Fatalf("CreateAdminKey: expected key ID %s, got %s", adminKeyID, adminKey.ID)
		}
	})

	t.Run("RetrieveAdminKey", func(t *testing.T) {
		adminKey, err := client.RetrieveAdminKey(ctx, adminKeyID)
		checks.NoError(t, err, "RetrieveAdminKey error")

		if adminKey.ID != adminKeyID {
			t.Fatalf("RetrieveAdminKey: expected key ID %s, got %s", adminKeyID, adminKey.ID)
		}
	})

	t.Run("DeleteAdminKey", func(t *testing.T) {
		adminKeyDeleteResponse, err := client.DeleteAdminKey(ctx, adminKeyID)
		checks.NoError(t, err, "DeleteAdminKey error")

		if !adminKeyDeleteResponse.Deleted {
			t.Fatalf("DeleteAdminKey: expected key to be deleted, got not deleted")
		}
	})
}
