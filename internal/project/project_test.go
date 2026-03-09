package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestStore(t *testing.T) (*Store, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "ccmux-project-test")
	if err != nil {
		t.Fatal(err)
	}

	s := &Store{
		filePath: filepath.Join(tmpDir, "projects.json"),
	}

	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return s, repoDir, cleanup
}

func TestAdd_ShouldStoreProject_GivenValidProject(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	project := &Project{
		Name: "test-project",
		Path: repoDir,
	}

	// Execute.
	err := store.Add(project)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	retrieved, err := store.Get("test-project")
	if err != nil {
		t.Fatalf("failed to retrieve project: %v", err)
	}
	if retrieved.Name != "test-project" {
		t.Errorf("expected name 'test-project', got '%s'", retrieved.Name)
	}
}

func TestAdd_ShouldFail_GivenNonGitRepo(t *testing.T) {
	// Setup.
	store, _, cleanup := setupTestStore(t)
	defer cleanup()
	tmpDir, _ := os.MkdirTemp("", "non-git")
	defer os.RemoveAll(tmpDir)

	project := &Project{
		Name: "bad-project",
		Path: tmpDir,
	}

	// Execute.
	err := store.Add(project)

	// Assert.
	if err == nil {
		t.Error("expected error for non-git repo, got nil")
	}
}

func TestAdd_ShouldFail_GivenDuplicateName(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	store.Add(&Project{Name: "dup", Path: repoDir})

	// Execute.
	err := store.Add(&Project{Name: "dup", Path: repoDir})

	// Assert.
	if err == nil {
		t.Error("expected error for duplicate name, got nil")
	}
}

func TestList_ShouldReturnAllProjects_GivenMultipleProjects(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	store.Add(&Project{Name: "proj-a", Path: repoDir})
	store.Add(&Project{Name: "proj-b", Path: repoDir})

	// Execute.
	projects, err := store.List()

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestList_ShouldReturnSortedByName_GivenMultipleProjects(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	store.Add(&Project{Name: "zebra", Path: repoDir})
	store.Add(&Project{Name: "alpha", Path: repoDir})

	// Execute.
	projects, err := store.List()

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projects[0].Name != "alpha" {
		t.Errorf("expected first project to be 'alpha', got '%s'", projects[0].Name)
	}
}

func TestAdd_ShouldPersistFastWorktrees_GivenFlagEnabled(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()

	// Execute.
	err := store.Add(&Project{Name: "fast-proj", Path: repoDir, UseFastWorktrees: true})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	retrieved, err := store.Get("fast-proj")
	if err != nil {
		t.Fatalf("failed to retrieve: %v", err)
	}
	if !retrieved.UseFastWorktrees {
		t.Error("expected UseFastWorktrees to be true")
	}
}

func TestUpdate_ShouldModifyProject_GivenValidName(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	store.Add(&Project{Name: "upd", Path: repoDir})

	// Execute.
	err := store.Update("upd", func(p *Project) {
		p.UseFastWorktrees = true
	})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	retrieved, _ := store.Get("upd")
	if !retrieved.UseFastWorktrees {
		t.Error("expected UseFastWorktrees to be true after update")
	}
}

func TestUpdate_ShouldFail_GivenInvalidName(t *testing.T) {
	// Setup.
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	// Execute.
	err := store.Update("nonexistent", func(p *Project) {
		p.UseFastWorktrees = true
	})

	// Assert.
	if err == nil {
		t.Error("expected error for nonexistent project, got nil")
	}
}

func TestIsProjDirectory_ShouldReturnTrue_GivenDotRepoExists(t *testing.T) {
	// Setup.
	tmpDir, _ := os.MkdirTemp("", "proj-test")
	defer os.RemoveAll(tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".repo"), 0755)

	// Execute.
	result := IsProjDirectory(tmpDir)

	// Assert.
	if !result {
		t.Error("expected IsProjDirectory to return true")
	}
}

func TestIsProjDirectory_ShouldReturnFalse_GivenNoDotRepo(t *testing.T) {
	// Setup.
	tmpDir, _ := os.MkdirTemp("", "proj-test")
	defer os.RemoveAll(tmpDir)

	// Execute.
	result := IsProjDirectory(tmpDir)

	// Assert.
	if result {
		t.Error("expected IsProjDirectory to return false")
	}
}

func TestFindProjTemplateDir_ShouldFindTemplate_GivenZeroPrefixDir(t *testing.T) {
	// Setup.
	tmpDir, _ := os.MkdirTemp("", "proj-test")
	defer os.RemoveAll(tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, "00-template"), 0755)

	// Execute.
	result, err := FindProjTemplateDir(tmpDir)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "00-template" {
		t.Errorf("expected '00-template', got '%s'", filepath.Base(result))
	}
}

func TestFindProjTemplateDir_ShouldFail_GivenNoTemplate(t *testing.T) {
	// Setup.
	tmpDir, _ := os.MkdirTemp("", "proj-test")
	defer os.RemoveAll(tmpDir)

	// Execute.
	_, err := FindProjTemplateDir(tmpDir)

	// Assert.
	if err == nil {
		t.Error("expected error when no template dir found")
	}
}

func TestRemove_ShouldDeleteProject_GivenValidName(t *testing.T) {
	// Setup.
	store, repoDir, cleanup := setupTestStore(t)
	defer cleanup()
	store.Add(&Project{Name: "to-remove", Path: repoDir})

	// Execute.
	err := store.Remove("to-remove")

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	projects, _ := store.List()
	if len(projects) != 0 {
		t.Errorf("expected 0 projects after removal, got %d", len(projects))
	}
}

