package tui

import (
	"testing"
)

func TestFormatBytes_ShouldReturnGb_GivenGigabyteValue(t *testing.T) {
	// Setup.
	bytes := int64(2 * 1024 * 1024 * 1024)

	// Execute.
	result := formatBytes(bytes)

	// Assert.
	if result != "2.0Gb" {
		t.Errorf("expected '2.0Gb', got '%s'", result)
	}
}

func TestFormatBytes_ShouldReturnFractionalGb_GivenFractionalGigabytes(t *testing.T) {
	// Setup.
	bytes := int64(1.5 * 1024 * 1024 * 1024)

	// Execute.
	result := formatBytes(bytes)

	// Assert.
	if result != "1.5Gb" {
		t.Errorf("expected '1.5Gb', got '%s'", result)
	}
}

func TestFormatBytes_ShouldReturnMb_GivenMegabyteValue(t *testing.T) {
	// Setup.
	bytes := int64(256 * 1024 * 1024)

	// Execute.
	result := formatBytes(bytes)

	// Assert.
	if result != "256Mb" {
		t.Errorf("expected '256Mb', got '%s'", result)
	}
}

func TestFormatBytes_ShouldReturnKb_GivenSmallValue(t *testing.T) {
	// Setup.
	bytes := int64(512 * 1024)

	// Execute.
	result := formatBytes(bytes)

	// Assert.
	if result != "512Kb" {
		t.Errorf("expected '512Kb', got '%s'", result)
	}
}

func TestFormatBytes_ShouldReturnZeroMb_GivenZero(t *testing.T) {
	// Setup.
	bytes := int64(0)

	// Execute.
	result := formatBytes(bytes)

	// Assert.
	if result != "0Mb" {
		t.Errorf("expected '0Mb', got '%s'", result)
	}
}

func TestFormatResourceLine_ShouldReturnEmpty_GivenNil(t *testing.T) {
	// Execute.
	result := formatResourceLine(nil)

	// Assert.
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFormatResourceLine_ShouldFormatAll_GivenResources(t *testing.T) {
	// Setup.
	r := &AgentResources{
		CPUPercent: 45,
		MemBytes:   int64(1.5 * 1024 * 1024 * 1024),
		MemPercent: 3,
		DiskBytes:  int64(156 * 1024 * 1024),
	}

	// Execute.
	result := formatResourceLine(r)

	// Assert.
	expected := "CPU: 45%  Mem: 1.5Gb (3%)  Disk: 156Mb"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFindDescendants_ShouldReturnAllDescendants_GivenProcessTree(t *testing.T) {
	// Setup.
	procs := map[int]*procInfo{
		1:   {pid: 1, ppid: 0, cpu: 0, rss: 0},
		100: {pid: 100, ppid: 1, cpu: 10, rss: 1000},
		200: {pid: 200, ppid: 100, cpu: 20, rss: 2000},
		300: {pid: 300, ppid: 100, cpu: 5, rss: 500},
		400: {pid: 400, ppid: 200, cpu: 15, rss: 1500},
		999: {pid: 999, ppid: 1, cpu: 50, rss: 5000},
	}

	// Execute.
	result := findDescendants(100, procs)

	// Assert.
	resultSet := make(map[int]bool)
	for _, pid := range result {
		resultSet[pid] = true
	}
	expectedPIDs := []int{100, 200, 300, 400}
	for _, pid := range expectedPIDs {
		if !resultSet[pid] {
			t.Errorf("expected PID %d in descendants", pid)
		}
	}
	if resultSet[1] {
		t.Error("PID 1 should not be in descendants")
	}
	if resultSet[999] {
		t.Error("PID 999 should not be in descendants")
	}
	if len(result) != 4 {
		t.Errorf("expected 4 descendants, got %d", len(result))
	}
}

func TestFindDescendants_ShouldReturnOnlyRoot_GivenNoChildren(t *testing.T) {
	// Setup.
	procs := map[int]*procInfo{
		1:   {pid: 1, ppid: 0},
		100: {pid: 100, ppid: 1},
		200: {pid: 200, ppid: 1},
	}

	// Execute.
	result := findDescendants(100, procs)

	// Assert.
	if len(result) != 1 {
		t.Errorf("expected 1 descendant (just root), got %d", len(result))
	}
	if result[0] != 100 {
		t.Errorf("expected root PID 100, got %d", result[0])
	}
}
