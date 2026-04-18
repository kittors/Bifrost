package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

func TestIntegrationDeviceDisabledBlocksClientRefresh(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service, db := newSeededIntegrationService(t, ctx)
	login := bootstrapSeedClient(t, ctx, service, "alice")

	if _, err := db.ExecContext(
		ctx,
		`UPDATE devices SET status = 'disabled' WHERE id = $1`,
		login.Device.ID,
	); err != nil {
		t.Fatalf("disable device: %v", err)
	}

	if _, err := service.RefreshSession(ctx, auth.RefreshInput{
		DeviceID:     login.Device.ID,
		RefreshToken: login.RefreshToken,
	}); err == nil {
		t.Fatal("expected disabled device to block refresh")
	} else {
		var serviceErr *auth.ServiceError
		if !errors.As(err, &serviceErr) {
			t.Fatalf("expected service error, got %T: %v", err, err)
		}
		if serviceErr.Code != contracts.ErrorCodeDeviceDisabled {
			t.Fatalf("expected device disabled, got %s", serviceErr.Code)
		}
	}
}
