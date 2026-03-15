package deploycmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	exists    bool
	existsErr error
	saveErr   error

	savedPkg     string
	savedVersion string
	savedAt      time.Time
	savedExecs   []SpawnedExecutor
	saveCalls    int
}

func (f *fakeRepo) RevisionExists(ctx context.Context, pkgName, version string) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeRepo) SaveDeployment(ctx context.Context, pkgName, version string, deployedAt time.Time, executors []SpawnedExecutor) error {
	f.saveCalls++
	f.savedPkg = pkgName
	f.savedVersion = version
	f.savedAt = deployedAt
	f.savedExecs = append([]SpawnedExecutor(nil), executors...)
	return f.saveErr
}

type createCall struct {
	anchor string
	image  string
}

type removeCall struct {
	anchor      string
	containerID string
}

type createResult struct {
	exec SpawnedExecutor
	err  error
}

type fakeRuntime struct {
	anchors    []string
	anchorsErr error

	createResults []createResult
	createCalls   []createCall

	removeCalls []removeCall
	removeErrs  map[string]error
}

func (f *fakeRuntime) ApprovedAnchors(ctx context.Context) ([]string, error) {
	return f.anchors, f.anchorsErr
}

func (f *fakeRuntime) CreateExecutor(ctx context.Context, anchorName, image string) (SpawnedExecutor, error) {
	f.createCalls = append(f.createCalls, createCall{
		anchor: anchorName,
		image:  image,
	})

	if len(f.createResults) == 0 {
		return SpawnedExecutor{}, errors.New("unexpected CreateExecutor call")
	}

	result := f.createResults[0]
	f.createResults = f.createResults[1:]
	return result.exec, result.err
}

func (f *fakeRuntime) RemoveExecutor(ctx context.Context, anchorName, containerID string) error {
	f.removeCalls = append(f.removeCalls, removeCall{
		anchor:      anchorName,
		containerID: containerID,
	})

	if f.removeErrs == nil {
		return nil
	}
	return f.removeErrs[containerID]
}

func newTestDeployer(repo *fakeRepo, runtime *fakeRuntime, load func(string) ([]string, error), out io.Writer) *deployer {
	return &deployer{
		repo:               repo,
		runtime:            runtime,
		loadExecutorImages: load,
		now: func() time.Time {
			return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		},
		out: out,
	}
}

func TestDeploy_InvalidPackageRef(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return nil, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "badref")
	require.Error(t, err)
	require.ErrorContains(t, err, "package must be in the format name@version")
}

func TestDeploy_AlreadyDeployed(t *testing.T) {
	repo := &fakeRepo{exists: true}
	runtime := &fakeRuntime{}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		t.Fatal("loadExecutorImages should not be called")
		return nil, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "hello@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "hello@1.0.0 has already been deployed")
}

func TestDeploy_LoadExecutorImagesError(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return nil, errors.New("manifest failed")
	}, io.Discard)

	err := d.Deploy(context.Background(), "hello@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "manifest failed")
}

func TestDeploy_ApprovedAnchorsError(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{
		anchorsErr: errors.New("anchors failed"),
	}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return []string{"img1"}, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "hello@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "anchors failed")
}

func TestDeploy_Success(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{
		anchors: []string{"anchor-a", "anchor-b"},
		createResults: []createResult{
			{exec: SpawnedExecutor{ExecutorName: "exec-1", AnchorName: "anchor-a", ContainerID: "cont-1", ImgName: "img-1"}},
			{exec: SpawnedExecutor{ExecutorName: "exec-2", AnchorName: "anchor-a", ContainerID: "cont-2", ImgName: "img-2"}},
			{exec: SpawnedExecutor{ExecutorName: "exec-3", AnchorName: "anchor-b", ContainerID: "cont-3", ImgName: "img-1"}},
			{exec: SpawnedExecutor{ExecutorName: "exec-4", AnchorName: "anchor-b", ContainerID: "cont-4", ImgName: "img-2"}},
		},
	}

	var out bytes.Buffer
	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return []string{"img-1", "img-2"}, nil
	}, &out)

	err := d.Deploy(context.Background(), "mypkg@2.3.4")
	require.NoError(t, err)

	require.Equal(t, 1, repo.saveCalls)
	require.Equal(t, "mypkg", repo.savedPkg)
	require.Equal(t, "2.3.4", repo.savedVersion)
	require.Equal(t, time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC), repo.savedAt)

	require.Len(t, repo.savedExecs, 4)
	require.Equal(t, []createCall{
		{anchor: "anchor-a", image: "img-1"},
		{anchor: "anchor-a", image: "img-2"},
		{anchor: "anchor-b", image: "img-1"},
		{anchor: "anchor-b", image: "img-2"},
	}, runtime.createCalls)

	require.Contains(t, out.String(), "Spawning executors on anchor 'anchor-a'")
	require.Contains(t, out.String(), "Spawning executors on anchor 'anchor-b'")
}

func TestDeploy_CreateFailure_RollsBackInReverseOrder(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{
		anchors: []string{"anchor-a"},
		createResults: []createResult{
			{exec: SpawnedExecutor{ExecutorName: "exec-1", AnchorName: "anchor-a", ContainerID: "cont-1", ImgName: "img-1"}},
			{exec: SpawnedExecutor{ExecutorName: "exec-2", AnchorName: "anchor-a", ContainerID: "cont-2", ImgName: "img-2"}},
			{err: errors.New("boom on img-3")},
		},
	}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return []string{"img-1", "img-2", "img-3"}, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "mypkg@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "boom on img-3")

	require.Equal(t, []removeCall{
		{anchor: "anchor-a", containerID: "cont-2"},
		{anchor: "anchor-a", containerID: "cont-1"},
	}, runtime.removeCalls)

	require.Equal(t, 0, repo.saveCalls)
}

func TestDeploy_CreateFailure_RollbackFailureIsIncluded(t *testing.T) {
	repo := &fakeRepo{}
	runtime := &fakeRuntime{
		anchors: []string{"anchor-a"},
		createResults: []createResult{
			{exec: SpawnedExecutor{ExecutorName: "exec-1", AnchorName: "anchor-a", ContainerID: "cont-1", ImgName: "img-1"}},
			{err: errors.New("create failed")},
		},
		removeErrs: map[string]error{
			"cont-1": errors.New("remove failed"),
		},
	}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return []string{"img-1", "img-2"}, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "mypkg@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "create failed")
	require.ErrorContains(t, err, "rollback failed")
	require.ErrorContains(t, err, "remove cont-1")
}

func TestDeploy_SaveFailure_RollsBackSpawnedExecutors(t *testing.T) {
	repo := &fakeRepo{
		saveErr: errors.New("db save failed"),
	}
	runtime := &fakeRuntime{
		anchors: []string{"anchor-a"},
		createResults: []createResult{
			{exec: SpawnedExecutor{ExecutorName: "exec-1", AnchorName: "anchor-a", ContainerID: "cont-1", ImgName: "img-1"}},
			{exec: SpawnedExecutor{ExecutorName: "exec-2", AnchorName: "anchor-a", ContainerID: "cont-2", ImgName: "img-2"}},
		},
	}

	d := newTestDeployer(repo, runtime, func(string) ([]string, error) {
		return []string{"img-1", "img-2"}, nil
	}, io.Discard)

	err := d.Deploy(context.Background(), "mypkg@1.0.0")
	require.Error(t, err)
	require.ErrorContains(t, err, "db save failed")

	require.Equal(t, []removeCall{
		{anchor: "anchor-a", containerID: "cont-2"},
		{anchor: "anchor-a", containerID: "cont-1"},
	}, runtime.removeCalls)

	require.Equal(t, 1, repo.saveCalls)
}
