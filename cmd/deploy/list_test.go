package deploycmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/spf13/cobra"
)

type fakeDeployQ struct {
	revisions      []db.Revision
	revisionsErr   error
	executorsByRev map[int64][]db.Executor
	executorsErr   map[int64]error
	calledIDs      []int64
}

func (f *fakeDeployQ) ListRevisions(ctx context.Context) ([]db.Revision, error) {
	if f.revisionsErr != nil {
		return nil, f.revisionsErr
	}
	return f.revisions, nil
}

func (f *fakeDeployQ) ListExecutorsByRevision(ctx context.Context, revisionID int64) ([]db.Executor, error) {
	f.calledIDs = append(f.calledIDs, revisionID)

	if err := f.executorsErr[revisionID]; err != nil {
		return nil, err
	}
	return f.executorsByRev[revisionID], nil
}

func newTestCmd(buf *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())
	return cmd
}

func cleanTableOutput(s string) string {
	return strings.ReplaceAll(s, zeroWidthSpace, "")
}

func TestRunList_RendersRevisionsAndExecutors(t *testing.T) {
	var out bytes.Buffer
	cmd := newTestCmd(&out)

	deployTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	q := &fakeDeployQ{
		revisions: []db.Revision{
			{
				ID:          1,
				PackageName: "pkg-a",
				Version:     "1.0.0",
				DeployTime:  deployTime,
			},
			{
				ID:          2,
				PackageName: "pkg-b",
				Version:     "2.0.0",
				DeployTime:  deployTime,
			},
		},
		executorsByRev: map[int64][]db.Executor{
			1: {
				{
					ExecutorName: "exec-1",
					AnchorName:   "anchor-1",
					ContainerID:  "1234567890abcdef",
					ImgName:      "img-1",
				},
			},
			2: nil,
		},
		executorsErr: map[int64]error{},
	}

	err := runList(cmd, nil, q)
	if err != nil {
		t.Fatalf("runList returned error: %v", err)
	}

	got := cleanTableOutput(out.String())

	if !strings.Contains(got, "pkg-a") {
		t.Fatalf("expected output to contain pkg-a, got:\n%s", got)
	}
	if !strings.Contains(got, "1.0.0") {
		t.Fatalf("expected output to contain version 1.0.0, got:\n%s", got)
	}
	if !strings.Contains(got, "exec-1") {
		t.Fatalf("expected output to contain executor name, got:\n%s", got)
	}
	if !strings.Contains(got, "anchor-1") {
		t.Fatalf("expected output to contain anchor name, got:\n%s", got)
	}
	if !strings.Contains(got, "1234567890...") {
		t.Fatalf("expected truncated container ID, got:\n%s", got)
	}
	if !strings.Contains(got, "img-1") {
		t.Fatalf("expected image name, got:\n%s", got)
	}
	if !strings.Contains(got, deployTime.Format(time.RFC3339)) {
		t.Fatalf("expected deploy time, got:\n%s", got)
	}
	if !strings.Contains(got, "No Executors") {
		t.Fatalf("expected output to contain No Executors row, got:\n%s", got)
	}

	if len(q.calledIDs) != 2 || q.calledIDs[0] != 1 || q.calledIDs[1] != 2 {
		t.Fatalf("expected ListExecutorsByRevision called with [1 2], got %v", q.calledIDs)
	}
}

func TestRunList_ListRevisionsError(t *testing.T) {
	var out bytes.Buffer
	cmd := newTestCmd(&out)

	wantErr := errors.New("db exploded")
	q := &fakeDeployQ{
		revisionsErr:   wantErr,
		executorsByRev: map[int64][]db.Executor{},
		executorsErr:   map[int64]error{},
	}

	err := runList(cmd, nil, q)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestRunList_ListExecutorsByRevisionError(t *testing.T) {
	var out bytes.Buffer
	cmd := newTestCmd(&out)

	wantErr := errors.New("executor lookup failed")

	q := &fakeDeployQ{
		revisions: []db.Revision{
			{
				ID:          7,
				PackageName: "pkg",
				Version:     "1.2.3",
				DeployTime:  time.Now(),
			},
		},
		executorsByRev: map[int64][]db.Executor{},
		executorsErr: map[int64]error{
			7: wantErr,
		},
	}

	err := runList(cmd, nil, q)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestNewDeployListCmd(t *testing.T) {
	cmd := newDeployListCmd()

	if cmd.Use != "list" {
		t.Fatalf("expected Use=list, got %q", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatal("expected RunE to be set")
	}
}
