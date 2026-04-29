package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

func TestIntegrationCreateUserPersistsRoleAssignments(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service, _ := newSeededIntegrationService(t, ctx)
	adminLogin := bootstrapSeedAdmin(t, ctx, service)

	created, err := service.CreateAdminUser(ctx, auth.CreateAdminUserInput{
		AccessToken: adminLogin.AccessToken,
		RequestID:   "req_integration_user_role_create",
		Username:    "foxtrot",
		DisplayName: "Foxtrot",
		Email:       "foxtrot@example.com",
		Password:    "ChangeMe123!",
		RoleIDs:     []string{"role_developer"},
	})
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	if created.Roles[0] != "role_developer" {
		t.Fatalf("expected developer role, got %#v", created.Roles)
	}

	users, err := service.ListAdminUsers(ctx, auth.ListAdminUsersInput{
		AccessToken: adminLogin.AccessToken,
		Keyword:     "foxtrot",
		Page:        1,
		PageSize:    10,
		RoleID:      "role_developer",
	})
	if err != nil {
		t.Fatalf("list admin users: %v", err)
	}

	if len(users.Items) != 1 {
		t.Fatalf("expected one filtered user, got %d", len(users.Items))
	}
	if users.Items[0].ID != created.ID {
		t.Fatalf("expected created user %s, got %s", created.ID, users.Items[0].ID)
	}
}
