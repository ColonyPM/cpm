package pkgcmd

import (
	"testing"
	"os"
	"path/filepath"
)


func TestCreatFile_Readme(t *testing.T) {
	tempDir := t.TempDir()
	
	
	pkgDir := filepath.Join(tempDir, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal("mkdir", err)
	}
	
	if err := createFile(pkgDir, "readme.md"); err != nil {
		t.Fatal("createFile", err)
	}
	
	content, err := os.ReadFile(filepath.Join(pkgDir, "readme.md"))
	if err != nil {
		t.Fatal("readFile", err)
	}
	
	if string(content) != "# mypkg\n" {
		t.Errorf("Unexpected content: %s", string(content))
	}
}

func TestCreatFile_Values(t *testing.T) {
	tempDir := t.TempDir()
	
	pkgDir := filepath.Join(tempDir, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal("mkdir", err)
	}
	
	if err := createFile(pkgDir, "values.yaml"); err != nil {
		t.Fatal("createFile", err)
	}
	
	content, err := os.ReadFile(filepath.Join(pkgDir, "values.yaml"))
	if err != nil {
		t.Fatal("readFile", err)
	}
	
	
	if string(content) != "global:\n" {
		t.Errorf("contents mismatch:\n got: \nwant:")
	}

}

func TestCreatFile_unkown(t *testing.T) {
	tempDir := t.TempDir()
	
	pkgDir := filepath.Join(tempDir, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal("mkdir", err)
	}
	
	err := createFile(pkgDir, "unknown.yaml")
		
	if err == nil {
		t.Fatal("expected error got nil")
	}
}
