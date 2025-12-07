package deployment

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/briandowns/spinner"
	"github.com/colonyos/colonies/pkg/client"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-sdk/container"
	"github.com/docker/go-sdk/container/wait"
)

type Executor struct {
	Name  string
	Image string
	Env   map[string]string
}

const (
	green = "\x1b[32m"
	red   = "\x1b[31m"
	reset = "\x1b[0m"

	greenOK = green + "✔" + reset
	redX    = red + "❌" + reset
)

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func waitForColoniesExecutor(cc *client.ColoniesClient, env map[string]string, timeout time.Duration) *wait.NopStrategy {
	return wait.ForNop(func(ctx context.Context, _ wait.StrategyTarget) error {
		ticker := time.NewTicker(1 * time.Second) // poll interval
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()

			case <-ticker.C:
				exec, err := cc.GetExecutor(env["COLONIES_COLONY_NAME"], env["EXECUTOR_NAME"], env["COLONIES_PRVKEY"])
				if err == nil && exec != nil {
					return nil
				}

				if err != nil {
					msg := err.Error()

					if strings.Contains(msg, "Failed to get executor, executor is nil") {
						continue
					}

					return err
				}
			}
		}
	}).WithTimeout(timeout)
}

func RunSetupScriptContainers(ctx context.Context, images []string, env map[string]string) error {
	fmt.Printf("[+] Running %d setup %s\n", len(images), pluralize(len(images), "script", "scripts"))

	scriptTimeoutTime := 5 * time.Minute

	for _, img := range images {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = " => "
		s.Suffix = fmt.Sprintf(" %s", img)
		s.Start()

		ctr, err := container.Run(
			ctx,
			container.WithImage(img),
			container.WithEnv(env),
			container.WithHostConfigModifier(func(hc *dockercontainer.HostConfig) {
				hc.NetworkMode = "host"
			}),
			container.WithWaitStrategy(wait.ForExit().WithTimeout(scriptTimeoutTime)),
		)

		s.Stop()

		if err != nil {
			return fmt.Errorf("failed to run %q: %w", img, err)
		}

		state, err := ctr.State(ctx)
		if err != nil {
			return fmt.Errorf("failed to get container state for %q: %w", ctr.ID(), err)
		}
		if state.ExitCode != 0 {
			fmt.Printf(" => %s %s\n", redX, img)

			logs, logErr := ctr.Logs(ctx)
			if logErr != nil {
				fmt.Fprintf(os.Stderr, "failed to fetch logs for %s (%s): %v\n", img, ctr.ID(), logErr)
			} else {
				defer logs.Close()

				fmt.Fprintf(os.Stderr, "\n=== logs for %s ===\n", img)

				if _, copyErr := io.Copy(os.Stderr, logs); copyErr != nil {
					fmt.Fprintf(os.Stderr, "failed to copy logs for %s: %v\n", img, copyErr)
				}

				fmt.Fprintln(os.Stderr, "=== end logs ===")
			}

			return fmt.Errorf("%s exited with code %d", img, state.ExitCode)
		}

		if err := ctr.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container %q: %w", ctr.ID(), err)
		}

		fmt.Printf(" => %s %s\n", greenOK, img)
	}

	return nil
}

// Atomic semantics:
//   - success: all containers running, caller owns them and must terminate.
//   - failure: all already-started containers have been terminated above, and we return nil, err.
func StartExecutors(ctx context.Context, executors []Executor, env map[string]string) ([]*container.Container, error) {
	containers := make([]*container.Container, 0, len(executors))

	cc := storectx.GetColoniesClient(ctx)

	executorTimeout := 1 * time.Minute

	n := len(executors)
	fmt.Printf("[+] Starting %d %s\n", n, pluralize(n, "executor", "executors"))

	for _, executor := range executors {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = " => "
		s.Suffix = fmt.Sprintf(" %s", executor.Name)
		s.Start()

		ws := waitForColoniesExecutor(cc, executor.Env, executorTimeout)

		ctr, err := container.Run(
			ctx,
			container.WithImage(executor.Image),
			container.WithEnv(executor.Env),
			container.WithHostConfigModifier(func(hc *dockercontainer.HostConfig) {
				hc.NetworkMode = "host"
			}),
			container.WithWaitStrategy(ws),
		)

		s.Stop()

		if err != nil {
			for _, c := range containers {
				_ = container.Terminate(c)
			}

			return nil, fmt.Errorf("starting %s: %w", executor.Image, err)
		}

		fmt.Printf(" => %s %s\n", greenOK, executor.Name)
		containers = append(containers, ctr)
	}

	return containers, nil
}
