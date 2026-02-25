package otel

import (
	"fmt"
	"time"
)

type AgentMetrics struct {
	CostUSD        float64
	TokensIn       int64
	TokensOut      int64
	TokensCacheRead    int64
	TokensCacheCreate  int64
	ActiveTimeSec  float64
	LastUpdated    time.Time
}

func FormatCostLine(m *AgentMetrics) string {
	if m == nil || m.CostUSD == 0 {
		return ""
	}
	return fmt.Sprintf("$%.2f", m.CostUSD)
}

func FormatMetricsDetail(m *AgentMetrics) (cost, tokens, activeTime string) {
	if m == nil {
		return "", "", ""
	}
	if m.CostUSD > 0 {
		cost = fmt.Sprintf("$%.2f", m.CostUSD)
	}
	if m.TokensIn > 0 || m.TokensOut > 0 || m.TokensCacheRead > 0 {
		tokens = fmt.Sprintf("In: %s  Out: %s  Cache: %s",
			formatTokenCount(m.TokensIn),
			formatTokenCount(m.TokensOut),
			formatTokenCount(m.TokensCacheRead))
	}
	if m.ActiveTimeSec > 0 {
		activeTime = formatDuration(m.ActiveTimeSec)
	}
	return cost, tokens, activeTime
}

func formatTokenCount(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	if seconds < 3600 {
		m := int(seconds) / 60
		s := int(seconds) % 60
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	return fmt.Sprintf("%dh%02dm", h, m)
}
