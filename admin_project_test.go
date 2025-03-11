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

func TestAdminProject(t *testing.T) {
	adminProjectObject := "organization.project"
	adminProjectID := "project-abc-123"
	adminProjectName := "Project Name"

	adminProjectCreatedAt := int64(1711471533)
	adminProjectArchivedAt := int64(1711471533)
	adminProjectStatus := "active"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/organization/projects",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminProjectList{
					Object: "list",
					Data: []openai.AdminProject{
						{
							ID:         adminProjectID,
							Object:     adminProjectObject,
							Name:       adminProjectName,
							CreatedAt:  adminProjectCreatedAt,
							ArchivedAt: &adminProjectArchivedAt,
							Status:     adminProjectStatus,
						},
					},
					FirstID: "first_id",
					LastID:  "last_id",
					HasMore: false,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminProject{
					ID:        adminProjectID,
					Object:    adminProjectObject,
					Name:      adminProjectName,
					CreatedAt: adminProjectCreatedAt,
					Status:    adminProjectStatus,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/projects/"+adminProjectID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				resBytes, _ := json.Marshal(openai.AdminProject{
					ID:         adminProjectID,
					Object:     adminProjectObject,
					Name:       adminProjectName,
					CreatedAt:  adminProjectCreatedAt,
					ArchivedAt: &adminProjectArchivedAt,
					Status:     adminProjectStatus,
				})
				fmt.Fprintln(w, string(resBytes))

			case http.MethodPost:
				resBytes, _ := json.Marshal(openai.AdminProject{
					ID:        adminProjectID,
					Object:    adminProjectObject,
					Name:      adminProjectName,
					CreatedAt: adminProjectCreatedAt,
					Status:    adminProjectStatus,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/projects/"+adminProjectID+"/archive",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST request, got %s", r.Method)
			}
			resBytes, _ := json.Marshal(openai.AdminProject{
				ID:        adminProjectID,
				Object:    adminProjectObject,
				Name:      adminProjectName,
				CreatedAt: adminProjectCreatedAt,
				Status:    "archived",
			})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	ctx := context.Background()

	t.Run("ListAdminProjects", func(t *testing.T) {
		adminProjects, err := client.ListAdminProjects(ctx, nil, nil, nil)
		checks.NoError(t, err, "ListAdminProjects error")

		if len(adminProjects.Data) != 1 {
			t.Errorf("expected 1 project, got %d", len(adminProjects.Data))
		}

		if adminProjects.Data[0].ID != adminProjectID {
			t.Errorf("expected project ID %s, got %s", adminProjectID, adminProjects.Data[0].ID)
		}
	})

	t.Run("CreateAdminProject", func(t *testing.T) {
		adminProject, err := client.CreateAdminProject(ctx, adminProjectName)
		checks.NoError(t, err, "CreateAdminProject error")

		if adminProject.ID != adminProjectID {
			t.Errorf("expected project ID %s, got %s", adminProjectID, adminProject.ID)
		}
	})

	t.Run("GetAdminProject", func(t *testing.T) {
		adminProject, err := client.RetrieveAdminProject(ctx, adminProjectID)
		checks.NoError(t, err, "GetAdminProject error")

		if adminProject.ID != adminProjectID {
			t.Errorf("expected project ID %s, got %s", adminProjectID, adminProject.ID)
		}
	})

	t.Run("ModifyAdminProject", func(t *testing.T) {
		adminProject, err := client.ModifyAdminProject(ctx, adminProjectID, adminProjectName)
		checks.NoError(t, err, "ModifyAdminProject error")

		if adminProject.ID != adminProjectID {
			t.Errorf("expected project ID %s, got %s", adminProjectID, adminProject.ID)
		}
	})

	t.Run("ArchiveAdminProject", func(t *testing.T) {
		adminProject, err := client.ArchiveAdminProject(ctx, adminProjectID)
		checks.NoError(t, err, "ArchiveAdminProject error")

		if adminProject.Status != "archived" {
			t.Errorf("expected project status archived, got %s", adminProject.Status)
		}
	})
}
