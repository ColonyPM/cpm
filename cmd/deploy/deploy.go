package deploycmd

import (
	"fmt"
	"os"

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
	"AWS_S3_REGION",
	"AWS_S3_BUCKET",
	"AWS_S3_TLS",
	"AWS_S3_SKIPVERIFY",

	"MINIO_USER",
	"MINIO_PASSWORD",
}

func BuildEnv(executorName string) (map[string]string, error) {
	env, err := LoadEnvMap(envKeys)
	if err != nil {
		return nil, err
	}

	// Add your additional key
	env["EXECUTOR_NAME"] = executorName
	return env, nil
}

func deploy(cmd *cobra.Command, args []string) error {
	fmt.Println("deploy called")
	
	executorName := "my-executor" // or passed in as arg / flag, etc.

	env, err := BuildEnv(executorName)
	if err != nil {
		return err
	}
	
	return nil
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <pkg>",
		Short: "Deployment related commands",
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
