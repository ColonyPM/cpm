//go:build integration

package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunSetupScriptContainers_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	images := []string{"hello-world:latest"}
	env := map[string]string{"TEST_ENV": "true"}

	err := RunSetupScriptContainers(ctx, images, env)

	require.NoError(t, err, "Expected setup container to run and exit successfully")
}

func TestRunSetupScriptContainers_ImageNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	t.Cleanup(cancel)

	images := []string{"this-image-100-percent-does-not-exist:12345"}

	err := RunSetupScriptContainers(ctx, images, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to run")
}