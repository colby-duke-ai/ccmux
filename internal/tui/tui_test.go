package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/CDFalcon/ccmux/internal/project"
)

func newTestModel() model {
	branchInput := textinput.New()
	branchInput.Placeholder = "origin/master"
	branchInput.Width = 50
	branchInput.CharLimit = 100

	branchFilter := textinput.New()
	branchFilter.Placeholder = "Type to search branches..."
	branchFilter.Width = 50
	branchFilter.CharLimit = 100

	taskInput := newAutoGrowTextarea("Describe the task...", 60)

	progress := new(int64)

	return model{
		view:             ViewNewTaskBranch,
		branchInput:      branchInput,
		branchFilter:     branchFilter,
		taskInput:        taskInput,
		projectImports:   make(map[string]*projImportProcess),
		downloadProgress: progress,
	}
}

func TestBranchEntries_ShouldIncludeDefaultAndManual_GivenNoFilter(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main", "develop"}

	// Execute.
	entries := m.branchEntries()

	// Assert.
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	if entries[0].value != "origin/master" {
		t.Errorf("expected first entry 'origin/master', got '%s'", entries[0].value)
	}
	if !entries[1].isManual {
		t.Error("expected second entry to be manual")
	}
	if entries[2].value != "main" {
		t.Errorf("expected third entry 'main', got '%s'", entries[2].value)
	}
}

func TestBranchEntries_ShouldShowFilteredResults_GivenFilter(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main", "develop"}
	m.branchFilter.SetValue("dev")
	m.filteredBranches = []string{"develop"}

	// Execute.
	entries := m.branchEntries()

	// Assert.
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (default + manual + 1 filtered), got %d", len(entries))
	}
	if entries[2].value != "develop" {
		t.Errorf("expected filtered entry 'develop', got '%s'", entries[2].value)
	}
}

func TestHandleNewTaskBranchInputKeys_ShouldDefaultToOriginMaster_GivenEmptyInput(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewNewTaskBranchInput
	m.branchInput.SetValue("")

	// Execute.
	result, _ := m.handleNewTaskBranchInputKeys(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert.
	rm := result.(model)
	if rm.spawnBranch != "origin/master" {
		t.Errorf("expected 'origin/master', got '%s'", rm.spawnBranch)
	}
}

func TestHandleNewTaskBranchInputKeys_ShouldUseCustomBranch_GivenInput(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewNewTaskBranchInput
	m.branchInput.SetValue("my-custom-branch")

	// Execute.
	result, _ := m.handleNewTaskBranchInputKeys(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert.
	rm := result.(model)
	if rm.spawnBranch != "my-custom-branch" {
		t.Errorf("expected 'my-custom-branch', got '%s'", rm.spawnBranch)
	}
}

func TestHandleNewTaskBranchInputKeys_ShouldGoBack_GivenEsc(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewNewTaskBranchInput
	m.branchInput.SetValue("something")

	// Execute.
	result, _ := m.handleNewTaskBranchInputKeys(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert.
	rm := result.(model)
	if rm.view != ViewNewTaskBranch {
		t.Errorf("expected ViewNewTaskBranch, got %d", rm.view)
	}
	if rm.branchInput.Value() != "" {
		t.Errorf("expected branch input cleared, got '%s'", rm.branchInput.Value())
	}
}

func TestHandleNewTaskBranchKeys_ShouldNavigateDown_GivenDownKey(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main", "develop"}
	m.selectedIndex = 0

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyDown})

	// Assert.
	rm := result.(model)
	if rm.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", rm.selectedIndex)
	}
}

func TestHandleNewTaskBranchKeys_ShouldNavigateUp_GivenUpKey(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main", "develop"}
	m.selectedIndex = 2

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyUp})

	// Assert.
	rm := result.(model)
	if rm.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", rm.selectedIndex)
	}
}

func TestHandleNewTaskBranchKeys_ShouldSelectDefault_GivenEnterOnFirst(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main"}
	m.selectedIndex = 0

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert.
	rm := result.(model)
	if rm.spawnBranch != "origin/master" {
		t.Errorf("expected 'origin/master', got '%s'", rm.spawnBranch)
	}
	if rm.view != ViewNewTaskInput {
		t.Errorf("expected ViewNewTaskInput, got %d", rm.view)
	}
}

func TestHandleNewTaskBranchKeys_ShouldGoToManualInput_GivenEnterOnSecond(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchOptions = []string{"main"}
	m.selectedIndex = 1

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert.
	rm := result.(model)
	if rm.view != ViewNewTaskBranchInput {
		t.Errorf("expected ViewNewTaskBranchInput, got %d", rm.view)
	}
}

func TestHandleNewTaskBranchKeys_ShouldClearFilter_GivenEscWithFilter(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchFilter.SetValue("something")
	m.filteredBranches = []string{"something"}

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert.
	rm := result.(model)
	if rm.branchFilter.Value() != "" {
		t.Errorf("expected filter cleared, got '%s'", rm.branchFilter.Value())
	}
	if rm.view != ViewNewTaskBranch {
		t.Errorf("expected to stay on ViewNewTaskBranch, got %d", rm.view)
	}
}

func TestHandleNewTaskBranchKeys_ShouldGoBack_GivenEscWithNoFilter(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.branchFilter.SetValue("")

	// Execute.
	result, _ := m.handleNewTaskBranchKeys(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert.
	rm := result.(model)
	if rm.view != ViewSelectProject {
		t.Errorf("expected ViewSelectProject, got %d", rm.view)
	}
}

func TestIsProjectImporting_ShouldReturnTrue_GivenActiveImport(t *testing.T) {
	// Setup.
	m := newTestModel()
	proc := &projImportProcess{}
	m.projectImports["test-proj"] = proc

	// Execute.
	result := m.isProjectImporting("test-proj")

	// Assert.
	if !result {
		t.Error("expected isProjectImporting to return true for active import")
	}
}

func TestIsProjectImporting_ShouldReturnFalse_GivenDoneImport(t *testing.T) {
	// Setup.
	m := newTestModel()
	proc := &projImportProcess{}
	proc.finish(nil)
	m.projectImports["test-proj"] = proc

	// Execute.
	result := m.isProjectImporting("test-proj")

	// Assert.
	if result {
		t.Error("expected isProjectImporting to return false for done import")
	}
}

func TestIsProjectImporting_ShouldReturnFalse_GivenNoImport(t *testing.T) {
	// Setup.
	m := newTestModel()

	// Execute.
	result := m.isProjectImporting("nonexistent")

	// Assert.
	if result {
		t.Error("expected isProjectImporting to return false for nonexistent project")
	}
}

func TestHandleAddProjectFastWTKeys_ShouldGoBack_GivenEsc(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewAddProjectFastWT
	m.projectForm = newProjectForm()

	// Execute.
	result, _ := m.handleAddProjectFastWTKeys(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert.
	rm := result.(model)
	if rm.view != ViewAddProjectPath {
		t.Errorf("expected ViewAddProjectPath, got %d", rm.view)
	}
}

func TestHandleAddProjectPathKeys_ShouldGoToFastWT_GivenEnterWithPath(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewAddProjectPath
	m.projectForm = newProjectForm()
	m.projectForm.pathInput.SetValue("/some/path")

	// Execute.
	result, _ := m.handleAddProjectPathKeys(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert.
	rm := result.(model)
	if rm.view != ViewAddProjectFastWT {
		t.Errorf("expected ViewAddProjectFastWT, got %d", rm.view)
	}
	if rm.newProjectPath != "/some/path" {
		t.Errorf("expected path '/some/path', got '%s'", rm.newProjectPath)
	}
}

func TestHandleProjImportOutputKeys_ShouldGoBack_GivenEsc(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewProjImportOutput

	// Execute.
	result, _ := m.handleProjImportOutputKeys(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert.
	rm := result.(model)
	if rm.view != ViewManageProjects {
		t.Errorf("expected ViewManageProjects, got %d", rm.view)
	}
}

func TestProjImportProcess_ShouldStreamLines_GivenAppendAndGet(t *testing.T) {
	// Setup.
	proc := &projImportProcess{}

	// Execute.
	proc.appendLine("line 1")
	proc.appendLine("line 2")
	lines := proc.getLines()

	// Assert.
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "line 1" || lines[1] != "line 2" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestProjImportProcess_ShouldReportDone_GivenFinishCalled(t *testing.T) {
	// Setup.
	proc := &projImportProcess{}
	proc.appendLine("output")

	// Execute.
	proc.finish(nil)
	done, err := proc.isDone()

	// Assert.
	if !done {
		t.Error("expected done to be true")
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHandleManageProjectsEnter_ShouldShowImportOutput_GivenImportingProject(t *testing.T) {
	// Setup.
	m := newTestModel()
	m.view = ViewManageProjects
	m.selectedIndex = 0
	m.projects = []*project.Project{{Name: "importing-proj", Path: "/tmp"}}
	proc := &projImportProcess{}
	m.projectImports["importing-proj"] = proc

	// Execute.
	result, _ := m.handleManageProjectsEnter()

	// Assert.
	rm := result.(model)
	if rm.view != ViewProjImportOutput {
		t.Errorf("expected ViewProjImportOutput, got %d", rm.view)
	}
}
