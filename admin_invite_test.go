package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

type adminInviteFixture struct {
	object     string
	inviteID   string
	email      string
	role       string
	memberRole string
	status     string
	invitedAt  int64
	expiresAt  int64
	acceptedAt int64
	projects   []openai.AdminInviteProject
}

func newAdminInviteFixture() *adminInviteFixture {
	return &adminInviteFixture{
		object:     "organization.invite",
		inviteID:   "invite-abc-123",
		email:      "invite@openai.com",
		role:       "owner",
		memberRole: "member",
		status:     "pending",
		invitedAt:  1711471533,
		expiresAt:  1711471533,
		acceptedAt: 1711471533,
		projects: []openai.AdminInviteProject{
			{
				ID:   "project-id",
				Role: "owner",
			},
		},
	}
}

func (f *adminInviteFixture) newInvite() openai.AdminInvite {
	return openai.AdminInvite{
		Object:     f.object,
		ID:         f.inviteID,
		Email:      f.email,
		Role:       f.role,
		Status:     f.status,
		InvitedAt:  f.invitedAt,
		ExpiresAt:  f.expiresAt,
		AcceptedAt: f.acceptedAt,
		Projects:   f.projects,
	}
}

func (f *adminInviteFixture) listResponse() openai.AdminInviteList {
	return openai.AdminInviteList{
		Object:       "list",
		AdminInvites: []openai.AdminInvite{f.newInvite()},
		FirstID:      "first_id",
		LastID:       "last_id",
		HasMore:      false,
	}
}

func (f *adminInviteFixture) deleteResponse() openai.AdminInviteDeleteResponse {
	return openai.AdminInviteDeleteResponse{
		ID:      f.inviteID,
		Object:  f.object,
		Deleted: true,
	}
}

func (f *adminInviteFixture) projectsPointer() *[]openai.AdminInviteProject {
	return &f.projects
}

type adminInviteTestEnv struct {
	ctx     context.Context
	client  *openai.Client
	fixture *adminInviteFixture
}

func newAdminInviteTestEnv(t *testing.T) *adminInviteTestEnv {
	t.Helper()

	client, server, teardown := setupOpenAITestServer()
	t.Cleanup(teardown)

	fixture := newAdminInviteFixture()
	registerAdminInviteHandlers(server, fixture)

	return &adminInviteTestEnv{
		ctx:     context.Background(),
		client:  client,
		fixture: fixture,
	}
}

func registerAdminInviteHandlers(server *test.ServerTest, fixture *adminInviteFixture) {
	server.RegisterHandler(
		"/v1/organization/invites",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				respondWithJSON(w, fixture.listResponse())
			case http.MethodPost:
				respondWithJSON(w, fixture.newInvite())
			}
		},
	)

	server.RegisterHandler(
		"/v1/organization/invites/"+fixture.inviteID,
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodDelete:
				respondWithJSON(w, fixture.deleteResponse())
			case http.MethodGet:
				respondWithJSON(w, fixture.newInvite())
			}
		},
	)
}

func respondWithJSON(w http.ResponseWriter, payload interface{}) {
	resBytes, _ := json.Marshal(payload)
	fmt.Fprintln(w, string(resBytes))
}

func (e *adminInviteTestEnv) assertInviteID(t *testing.T, got string) {
	t.Helper()

	if got != e.fixture.inviteID {
		t.Errorf("expected admin invite ID %s, got %s", e.fixture.inviteID, got)
	}
}

func (e *adminInviteTestEnv) assertSingleInviteResponse(t *testing.T, resp openai.AdminInviteList) {
	t.Helper()

	if len(resp.AdminInvites) != 1 {
		t.Fatalf("expected 1 admin invite, got %d", len(resp.AdminInvites))
	}

	e.assertInviteID(t, resp.AdminInvites[0].ID)
}

func (e *adminInviteTestEnv) runListAdminInvites(t *testing.T, limit *int, after *string) {
	t.Helper()

	adminInvites, err := e.client.ListAdminInvites(e.ctx, limit, after)
	checks.NoError(t, err, "ListAdminInvites error")
	e.assertSingleInviteResponse(t, adminInvites)
}

func TestAdminInvite_List(t *testing.T) {
	env := newAdminInviteTestEnv(t)

	t.Run("WithoutFilters", func(t *testing.T) {
		env.runListAdminInvites(t, nil, nil)
	})

	t.Run("WithOnlyLimit", func(t *testing.T) {
		limit := 5
		env.runListAdminInvites(t, &limit, nil)
	})

	t.Run("WithOnlyAfter", func(t *testing.T) {
		after := "after-token"
		env.runListAdminInvites(t, nil, &after)
	})

	t.Run("WithLimitAndAfter", func(t *testing.T) {
		limit := 10
		after := "after-id"
		env.runListAdminInvites(t, &limit, &after)
	})
}

func TestAdminInvite_Create(t *testing.T) {
	env := newAdminInviteTestEnv(t)

	t.Run("WithProjects", func(t *testing.T) {
		adminInvite, err := env.client.CreateAdminInvite(env.ctx,
			env.fixture.email,
			env.fixture.role,
			env.fixture.projectsPointer(),
		)
		checks.NoError(t, err, "CreateAdminInvite error")
		env.assertInviteID(t, adminInvite.ID)
	})

	t.Run("WithoutProjects", func(t *testing.T) {
		adminInvite, err := env.client.CreateAdminInvite(env.ctx,
			env.fixture.email,
			env.fixture.role,
			nil,
		)
		checks.NoError(t, err, "CreateAdminInvite error")
		env.assertInviteID(t, adminInvite.ID)
	})

	t.Run("WithMemberRole", func(t *testing.T) {
		adminInvite, err := env.client.CreateAdminInvite(env.ctx,
			env.fixture.email,
			env.fixture.memberRole,
			env.fixture.projectsPointer(),
		)
		checks.NoError(t, err, "CreateAdminInvite error")
		env.assertInviteID(t, adminInvite.ID)
	})

	t.Run("InvalidRole", func(t *testing.T) {
		invalidRole := "invalid-role"
		_, err := env.client.CreateAdminInvite(env.ctx,
			env.fixture.email,
			invalidRole,
			env.fixture.projectsPointer(),
		)
		if err == nil {
			t.Fatal("expected error for invalid role, got nil")
		}
	})
}

func TestAdminInvite_Retrieve(t *testing.T) {
	env := newAdminInviteTestEnv(t)

	adminInvite, err := env.client.RetrieveAdminInvite(env.ctx, env.fixture.inviteID)
	checks.NoError(t, err, "RetrieveAdminInvite error")
	env.assertInviteID(t, adminInvite.ID)
}

func TestAdminInvite_Delete(t *testing.T) {
	env := newAdminInviteTestEnv(t)

	adminInviteDeleteResponse, err := env.client.DeleteAdminInvite(env.ctx, env.fixture.inviteID)
	checks.NoError(t, err, "DeleteAdminInvite error")

	if !adminInviteDeleteResponse.Deleted {
		t.Errorf("expected admin invite to be deleted, got not deleted")
	}
}
