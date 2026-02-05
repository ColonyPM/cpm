package deploycmd

import (
	"bytes"
	"context"
	"testing"
	"time"
	"github.com/ColonyPM/cpm/internal/db"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"errors"
)

type mockDeployQ struct {
	mock.Mock
}

func (m *mockDeployQ) ListDeployments(ctx context.Context) ([]db.Deployment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
			return nil, args.Error(1)
		}
		return args.Get(0).([]db.Deployment), args.Error(1)
	return args.Get(0).([]db.Deployment), args.Error(1)
}

func (m *mockDeployQ) ListExecutorsByDeployment(ctx context.Context, deploymentID int64) ([]db.Executor, error) {
	args := m.Called(ctx, deploymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Executor), args.Error(1)
}

func TestRunList_NoDeployment(t *testing.T){
	ctx := context.Background()
	
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	cmd.SetContext(context.Background())
	
	test := new(mockDeployQ)
	
	test.On("ListDeployments", ctx).Return([]db.Deployment{}, nil)

	
	err := runList(cmd, nil, test)
	require.NoError(t, err)
	
	out := output.String()
	assert.Contains(t, out, "Package")
	assert.Contains(t, out, "Executors")
	assert.Contains(t, out, "Created At")
	
	test.AssertExpectations(t)

}

func TestRunList_Deployment_NoExecutors(t *testing.T){
	ctx := context.Background()
	
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	cmd.SetContext(context.Background())
	
	test := new(mockDeployQ)
	
	deployments := []db.Deployment{
		{ID: 1, PkgName: "mypkg", DeployedAt: time.Now()},
	}
	
	test.On("ListDeployments", ctx).Return(deployments, nil)
	test.On("ListExecutorsByDeployment", ctx, int64(1)).Return([]db.Executor{}, nil)

	
	err := runList(cmd, nil, test)
	require.NoError(t, err)
	
	test.AssertExpectations(t)

}

func TestRunList_Deployment_WithExecutors(t *testing.T){
	ctx := context.Background()
	
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	cmd.SetContext(context.Background())
	
	test := new(mockDeployQ)
	
	deployments := []db.Deployment{
		{ID: 5, PkgName: "mypkg", DeployedAt: time.Now()},
	}
	
	executors := []db.Executor{
		{ID: 1, DeploymentID: 5, 	ExecutorName: "executor1", ContainerID: "conatiner1"},
		{ID: 2, DeploymentID: 5, 	ExecutorName: "executor2", ContainerID: "conatiner2"},
	}
	
	test.On("ListDeployments", ctx).Return(deployments, nil)
	test.On("ListExecutorsByDeployment", ctx, int64(5)).Return(executors, nil)

	
	err := runList(cmd, nil, test)
	require.NoError(t, err)
	
	out := output.String()
	assert.Contains(t, out, "mypkg")
	assert.Contains(t, out, "executor1")
	assert.Contains(t, out, "executor2")
	assert.Contains(t, out, "conatiner1")
	assert.Contains(t, out, "conatiner2")
	assert.Contains(t, out, time.Now().Format(time.RFC3339))
	
	test.AssertExpectations(t)

}

func TestRunList_ListDeploymentsError(t *testing.T){
	ctx := context.Background()
	
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	cmd.SetContext(context.Background())
	
	test := new(mockDeployQ)
	wantErr := errors.New("db down")

	test.On("ListDeployments", ctx).Return([]db.Deployment{}, wantErr)

	
	err := runList(cmd, nil, test)
	require.ErrorIs(t, err, wantErr)
	
	assert.Equal(t, "", output.String())

	test.AssertExpectations(t)
}

