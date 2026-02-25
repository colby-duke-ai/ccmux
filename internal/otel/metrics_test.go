package otel

import (
	"testing"
	"time"
)

func TestFormatCostLine_ShouldReturnEmpty_GivenNil(t *testing.T) {
	// Execute.
	result := FormatCostLine(nil)

	// Assert.
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFormatCostLine_ShouldReturnEmpty_GivenZeroCost(t *testing.T) {
	// Setup.
	m := &AgentMetrics{}

	// Execute.
	result := FormatCostLine(m)

	// Assert.
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFormatCostLine_ShouldFormatCost_GivenRealisticValue(t *testing.T) {
	// Setup.
	m := &AgentMetrics{CostUSD: 0.42}

	// Execute.
	result := FormatCostLine(m)

	// Assert.
	if result != "$0.42" {
		t.Errorf("expected '$0.42', got '%s'", result)
	}
}

func TestFormatCostLine_ShouldFormatLargeCost_GivenDollarAmount(t *testing.T) {
	// Setup.
	m := &AgentMetrics{CostUSD: 12.50}

	// Execute.
	result := FormatCostLine(m)

	// Assert.
	if result != "$12.50" {
		t.Errorf("expected '$12.50', got '%s'", result)
	}
}

func TestFormatMetricsDetail_ShouldReturnEmpty_GivenNil(t *testing.T) {
	// Execute.
	cost, tokens, active := FormatMetricsDetail(nil)

	// Assert.
	if cost != "" || tokens != "" || active != "" {
		t.Errorf("expected all empty, got cost='%s' tokens='%s' active='%s'", cost, tokens, active)
	}
}

func TestFormatMetricsDetail_ShouldReturnAll_GivenRealisticValues(t *testing.T) {
	// Setup.
	m := &AgentMetrics{
		CostUSD:        0.42,
		TokensIn:       125000,
		TokensOut:      8200,
		TokensCacheRead: 89000,
		ActiveTimeSec:  201,
		LastUpdated:    time.Now(),
	}

	// Execute.
	cost, tokens, active := FormatMetricsDetail(m)

	// Assert.
	if cost != "$0.42" {
		t.Errorf("expected '$0.42', got '%s'", cost)
	}
	if tokens != "In: 125.0k  Out: 8.2k  Cache: 89.0k" {
		t.Errorf("unexpected tokens: '%s'", tokens)
	}
	if active != "3m21s" {
		t.Errorf("expected '3m21s', got '%s'", active)
	}
}

func TestFormatTokenCount_ShouldReturnRaw_GivenSmallNumber(t *testing.T) {
	// Execute.
	result := formatTokenCount(999)

	// Assert.
	if result != "999" {
		t.Errorf("expected '999', got '%s'", result)
	}
}

func TestFormatTokenCount_ShouldReturnK_GivenThousands(t *testing.T) {
	// Execute.
	result := formatTokenCount(12500)

	// Assert.
	if result != "12.5k" {
		t.Errorf("expected '12.5k', got '%s'", result)
	}
}

func TestFormatTokenCount_ShouldReturnM_GivenMillions(t *testing.T) {
	// Execute.
	result := formatTokenCount(1500000)

	// Assert.
	if result != "1.5M" {
		t.Errorf("expected '1.5M', got '%s'", result)
	}
}

func TestFormatTokenCount_ShouldReturnZero_GivenZero(t *testing.T) {
	// Execute.
	result := formatTokenCount(0)

	// Assert.
	if result != "0" {
		t.Errorf("expected '0', got '%s'", result)
	}
}

func TestFormatDuration_ShouldReturnSeconds_GivenShortDuration(t *testing.T) {
	// Execute.
	result := formatDuration(45)

	// Assert.
	if result != "45s" {
		t.Errorf("expected '45s', got '%s'", result)
	}
}

func TestFormatDuration_ShouldReturnMinutes_GivenMediumDuration(t *testing.T) {
	// Execute.
	result := formatDuration(201)

	// Assert.
	if result != "3m21s" {
		t.Errorf("expected '3m21s', got '%s'", result)
	}
}

func TestFormatDuration_ShouldReturnHours_GivenLongDuration(t *testing.T) {
	// Execute.
	result := formatDuration(4320)

	// Assert.
	if result != "1h12m" {
		t.Errorf("expected '1h12m', got '%s'", result)
	}
}
