package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

func TestIntegrationUserServiceDenyOverridesRoleAllow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service, _ := newSeededIntegrationService(t, ctx)
	login := bootstrapSeedClient(t, ctx, service, "bob")

	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: login.AccessToken,
		ServiceKey:  "jenkins",
		RequestID:   "req_integration_policy",
	}); err == nil {
		t.Fatal("expected deny override to block jenkins access")
	} else {
		var serviceErr *auth.ServiceError
		if !errors.As(err, &serviceErr) {
			t.Fatalf("expected service error, got %T: %v", err, err)
		}
		if serviceErr.Code != contracts.ErrorCodePolicyAccessDenied {
			t.Fatalf("expected policy access denied, got %s", serviceErr.Code)
		}
	}
}
