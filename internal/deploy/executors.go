package deploy

import (
	"context"
	"fmt"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/container"

	"github.com/briandowns/spinner"
)

type Executor struct {
	Name  string
	Image string
	Env   map[string]string
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// Atomic semantics:
//   - success: all containers running, caller owns them and must terminate.
//   - failure: all already-started containers have been terminated above, and we return nil, err.
func StartExecutors(
	ctx context.Context,
	executors []Executor,
	env map[string]string,
) ([]*container.Container, error) {
	containers := make([]*container.Container, 0, len(executors))

	n := len(executors)
	fmt.Printf("Starting %d %s:\n", n, pluralize(n, "executor", "executors"))

	for _, executor := range executors {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = fmt.Sprintf("⮡ Starting executor %s ", executor.Name)
		s.Start()

		ctr, err := container.Run(
			ctx,
			container.WithImage(executor.Image),
			container.WithEnv(executor.Env),
			container.WithHostConfigModifier(func(hc *dockercontainer.HostConfig) {
				hc.NetworkMode = "host"
			}),
		)

		s.Stop()

		if err != nil {
			for _, c := range containers {
				_ = container.Terminate(c)
			}

			return nil, fmt.Errorf("starting %s: %w", executor.Image, err)
		}

		fmt.Printf("⮡ Started executor %s\n", executor.Name)
		containers = append(containers, ctr)
	}

	return containers, nil
}
