package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/CDFalcon/ccmux/internal/agent"
	"github.com/CDFalcon/ccmux/internal/tmux"
)

type AgentResources struct {
	CPUPercent float64
	MemBytes   int64
	MemPercent float64
	DiskBytes  int64
}

type procInfo struct {
	pid  int
	ppid int
	cpu  float64
	rss  int64
}

func queryAllAgentResources(agents []*agent.Agent, tmuxMgr *tmux.Manager, totalMemKB int64) map[string]*AgentResources {
	procs := listAllProcesses()
	resources := make(map[string]*AgentResources)

	type diskResult struct {
		agentID string
		bytes   int64
	}
	var wg sync.WaitGroup
	diskCh := make(chan diskResult, len(agents))

	for _, a := range agents {
		if a.WorktreePath != "" {
			wg.Add(1)
			go func(id, path string) {
				defer wg.Done()
				diskCh <- diskResult{id, getDiskUsage(path)}
			}(a.ID, a.WorktreePath)
		}
	}

	go func() {
		wg.Wait()
		close(diskCh)
	}()

	diskMap := make(map[string]int64)
	for r := range diskCh {
		diskMap[r.agentID] = r.bytes
	}

	for _, a := range agents {
		res := &AgentResources{}

		if a.TmuxWindow != "" {
			panePID, err := tmuxMgr.GetPanePID(a.TmuxWindow)
			if err == nil && panePID > 0 {
				descendants := findDescendants(panePID, procs)
				var totalCPU float64
				var totalRSS int64
				for _, pid := range descendants {
					if p, ok := procs[pid]; ok {
						totalCPU += p.cpu
						totalRSS += p.rss
					}
				}
				res.CPUPercent = totalCPU
				res.MemBytes = totalRSS * 1024
				if totalMemKB > 0 {
					res.MemPercent = float64(totalRSS) / float64(totalMemKB) * 100
				}
			}
		}

		res.DiskBytes = diskMap[a.ID]
		resources[a.ID] = res
	}

	return resources
}

func listAllProcesses() map[int]*procInfo {
	cmd := exec.Command("ps", "-e", "--no-headers", "-o", "pid:1,ppid:1,pcpu:1,rss:1")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	procs := make(map[int]*procInfo)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		rss, _ := strconv.ParseInt(fields[3], 10, 64)
		procs[pid] = &procInfo{pid: pid, ppid: ppid, cpu: cpu, rss: rss}
	}
	return procs
}

func findDescendants(rootPID int, procs map[int]*procInfo) []int {
	children := make(map[int][]int)
	for pid, p := range procs {
		children[p.ppid] = append(children[p.ppid], pid)
	}

	var result []int
	queue := []int{rootPID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, children[current]...)
	}
	return result
}

func getDiskUsage(path string) int64 {
	cmd := exec.Command("du", "-sb", path)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 1 {
		return 0
	}
	size, _ := strconv.ParseInt(fields[0], 10, 64)
	return size
}

func getTotalMemoryKB() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseInt(fields[1], 10, 64)
				return kb
			}
		}
	}
	return 0
}

func formatResourceLine(r *AgentResources) string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("CPU: %.0f%%  Mem: %s (%.0f%%)  Disk: %s",
		r.CPUPercent,
		formatBytes(r.MemBytes),
		r.MemPercent,
		formatBytes(r.DiskBytes),
	)
}

func formatBytes(bytes int64) string {
	const (
		gb = 1024 * 1024 * 1024
		mb = 1024 * 1024
	)
	if bytes >= gb {
		return fmt.Sprintf("%.1fGb", float64(bytes)/float64(gb))
	}
	if bytes >= mb {
		return fmt.Sprintf("%.0fMb", float64(bytes)/float64(mb))
	}
	if bytes > 0 {
		return fmt.Sprintf("%.0fKb", float64(bytes)/1024)
	}
	return "0Mb"
}
