package deploycmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/deployment"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func LoadEnvMap(keys []string) (map[string]string, error) {
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil, fmt.Errorf("missing required env var %s", k)
		}
		m[k] = v
	}
	return m, nil
}

var envKeys = []string{
	"COLONIES_SERVER_TLS",
	"COLONIES_SERVER_HOST",
	"COLONIES_SERVER_PORT",
	"COLONIES_COLONY_NAME",
	"COLONIES_COLONY_PRVKEY",
	"COLONIES_PRVKEY",

	"AWS_S3_ENDPOINT",
	"AWS_S3_ACCESSKEY",
	"AWS_S3_SECRETKEY",
	"AWS_S3_REGION_KEY",
	"AWS_S3_BUCKET",
	"AWS_S3_TLS",
	"AWS_S3_SKIPVERIFY",

	"MINIO_USER",
	"MINIO_PASSWORD",
}

func createUniqueExecutorName(prefix string, n int) string {
	id := uuid.New()
	hex := fmt.Sprintf("%x", id[:])
	if n > len(hex) {
		n = len(hex)
	}
	return prefix + hex[:n]
}

func buildEnv(executorName string) (map[string]string, error) {
	env, err := LoadEnvMap(envKeys)
	if err != nil {
		return nil, err
	}

	// Add your additional key
	env["EXECUTOR_NAME"] = createUniqueExecutorName(executorName+"-", 8)
	return env, nil
}

func prepareExecutors(executorSpecs []pkg.ExecutorSpec) ([]deployment.Executor, error) {
	execs := make([]deployment.Executor, 0, len(executorSpecs))
	for _, execSpec := range executorSpecs {
		env, err := buildEnv(execSpec.Name)
		if err != nil {
			return nil, err
		}

		execs = append(execs, deployment.Executor{
			Name:  execSpec.Name,
			Image: execSpec.Image,
			Env:   env,
		})
	}
	return execs, nil
}

func deploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()

	// 1. Get Package
	manifest, err := pkg.GetPackageManifest(args[0])
	if err != nil {
		return err
	}

	// 2. Setup scripts
	//    * Runs sequentially, in the order specified in the manifest
	//    * Waits for each script to exit before starting the next one
	env, err := buildEnv("")
	if err != nil {
		return err
	}

	if err := deployment.RunSetupScriptContainers(ctx, manifest.Deployments.Setup, env); err != nil {
		return err
	}

	// 3. Start Executors
	executors, err := prepareExecutors(manifest.Deployments.Executors)
	if err != nil {
		return err
	}

	containers, err := deployment.StartExecutors(ctx, executors, env)
	if err != nil {
		return err
	}

	// 4. Register deployment (dbConn)
	dbConn, q := storectx.GetDb(ctx)

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := q.WithTx(tx)

	dep, err := qtx.CreateDeployment(ctx, db.CreateDeploymentParams{
		PkgName:    args[0],
		DeployedAt: time.Now().UTC(),
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	for i, container := range containers {
		_, err := qtx.CreateExecutor(ctx, db.CreateExecutorParams{
			DeploymentID: dep.ID,
			ExecutorName: executors[i].Env["EXECUTOR_NAME"],
			ContainerID:  container.ShortID(),
		})
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("\nDeployed %s\n", dep.PkgName)

	// 5. (Maybe print info and suggestions of cpm deploy
	//     function/workflow with the possible fnSpecs / wfs)
	// ...
	// ...
	// ...

	return nil
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <pkg>",
		Short: "Deploy a package to the colony",
		Args:  cobra.ExactArgs(1),
		RunE:  deploy,
	}

	cmd.AddCommand(newDeployGetCmd())
	cmd.AddCommand(newDeployListCmd())
	cmd.AddCommand(newDeployRemoveCmd())
	cmd.AddCommand(newDeployFunctionCmd())
	cmd.AddCommand(newDeployWorkflowCmd())

	return cmd
}
