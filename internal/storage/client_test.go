package storage

import (
	"context"
	"fmt"
	"os/signal"
	"salsasync/internal/config"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	cfg := config.Config{
		StorageApi:    "http://localhost:9001/api/v1/",
		StorageApiKey: "HkYDkgRpFISd8tyL5G9pCAdKKpSKTJFl",
	}

	storageApi := New(cfg.StorageApi, cfg.StorageApiKey)
	users, err := storageApi.GetUsers(ctx)
	fmt.Println(users)
	assert.NoError(t, err)
}
