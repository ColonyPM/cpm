package deploycmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ColonyPM/cpm/internal/config"
	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/deployment"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func createUniqueExecutorName(prefix string, n int) string {
	id := uuid.New()
	hex := fmt.Sprintf("%x", id[:])
	if n > len(hex) {
		n = len(hex)
	}
	return prefix + hex[:n]
}

// buildEnvFromConfig converts config values to the map[string]string format expected by executors.
func buildEnvFromConfig(cfg *config.Config, executorName string) map[string]string {
	env := map[string]string{
		"COLONIES_SERVER_TLS":    strconv.FormatBool(cfg.Colonies.TLS),
		"COLONIES_SERVER_HOST":   cfg.Colonies.Host,
		"COLONIES_SERVER_PORT":   strconv.Itoa(cfg.Colonies.Port),
		"COLONIES_COLONY_NAME":   cfg.Colonies.ColonyName,
		"COLONIES_COLONY_PRVKEY": cfg.Colonies.ColonyPrvkey,
		"COLONIES_PRVKEY":        cfg.Colonies.Prvkey,

		"AWS_S3_ENDPOINT":   cfg.S3.Endpoint,
		"AWS_S3_ACCESSKEY":  cfg.S3.Accesskey,
		"AWS_S3_SECRETKEY":  cfg.S3.Secretkey,
		"AWS_S3_REGION_KEY": cfg.S3.Region,
		"AWS_S3_BUCKET":     cfg.S3.Bucket,
		"AWS_S3_TLS":        strconv.FormatBool(cfg.S3.TLS),
		"AWS_S3_SKIPVERIFY": strconv.FormatBool(cfg.S3.SkipVerify),

		"MINIO_USER":     cfg.Minio.User,
		"MINIO_PASSWORD": cfg.Minio.Password,

		"EXECUTOR_NAME": createUniqueExecutorName(executorName+"-", 8),
	}
	return env
}

func prepareExecutors(cfg *config.Config, executorSpecs []pkg.ExecutorSpec) []deployment.Executor {
	execs := make([]deployment.Executor, 0, len(executorSpecs))
	for _, execSpec := range executorSpecs {
		execs = append(execs, deployment.Executor{
			Name:  execSpec.Name,
			Image: execSpec.Image,
			Env:   buildEnvFromConfig(cfg, execSpec.Name),
		})
	}
	return execs
}

func deploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()
	cfg := storectx.GetConfig(ctx)

	// 1. Get Package
	manifest, err := pkg.GetPackageManifest(args[0])
	if err != nil {
		return err
	}

	// 2. Setup scripts
	//    * Runs sequentially, in the order specified in the manifest
	//    * Waits for each script to exit before starting the next one
	env := buildEnvFromConfig(cfg, "")

	if err := deployment.RunSetupScriptContainers(ctx, manifest.Deployments.Setup, env); err != nil {
		return err
	}

	// 3. Start Executors
	executors := prepareExecutors(cfg, manifest.Deployments.Executors)

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
