// Package windowlayout handles parsing window layout configs for agent windows.
// Supports two formats:
//
//  1. Simple ccmux format (top-level "panes" with direction/size/command):
//
//     panes:
//     - direction: h
//     size: 40%
//     command: "vim ."
//
//  2. Tmuxinator format (windows list with named window containing panes):
//
//     name: my-project
//     windows:
//     - my-window:
//     layout: main-vertical
//     panes:
//     - vim .
//     - bash
//     - make watch
package windowlayout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PaneDef defines a single additional tmux pane to create in an agent window.
type PaneDef struct {
	// Direction of the split: "h" (horizontal/left-right) or "v" (vertical/top-bottom).
	// Only used when Layout is empty. Defaults to "h".
	Direction string
	// Size of the new pane, e.g. "40%" or "40".
	// Only used when Layout is empty. Defaults to "50%".
	Size string
	// Command to run in the pane. Empty means an interactive shell.
	// Shell variables like $WORKTREE_PATH and $AGENT_ID are expanded.
	Command string
}

// Config defines the window layout for agent windows.
type Config struct {
	Panes  []PaneDef
	// Layout is a tmux layout name (e.g. "main-vertical", "tiled") to apply after
	// creating all panes. When set, Direction and Size on individual panes are ignored.
	Layout string
}

// simplePaneDef is the YAML shape for our own simple format.
type simplePaneDef struct {
	Direction string `yaml:"direction"`
	Size      string `yaml:"size"`
	Command   string `yaml:"command"`
}

// simpleConfig is the YAML shape for our own simple format.
type simpleConfig struct {
	Panes  []simplePaneDef `yaml:"panes"`
	Layout string          `yaml:"layout"`
}

// tmuxinatorWindow is the YAML shape for a tmuxinator window entry.
// Each entry in the windows list is a single-key map: { <name>: { layout, panes } }.
type tmuxinatorWindow struct {
	Layout string        `yaml:"layout"`
	Panes  []interface{} `yaml:"panes"`
}

// Load reads and parses a window layout config file.
// Supports ~ in path.
// Automatically detects whether the file is tmuxinator format or the simple format.
func Load(path string) (*Config, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read window layout config %q: %w", path, err)
	}

	// Parse into a generic map first to detect format.
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse window layout config %q: %w", path, err)
	}

	if _, hasTmuxWindows := raw["windows"]; hasTmuxWindows {
		return parseTmuxinator(data, path)
	}
	return parseSimple(data, path)
}

func parseSimple(data []byte, path string) (*Config, error) {
	var sc simpleConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("failed to parse window layout config %q: %w", path, err)
	}

	config := &Config{Layout: sc.Layout}
	for _, p := range sc.Panes {
		dir := strings.ToLower(strings.TrimSpace(p.Direction))
		switch dir {
		case "horizontal", "h", "":
			dir = "h"
		case "vertical", "v":
			dir = "v"
		default:
			dir = "h"
		}
		size := p.Size
		if size == "" {
			size = "50%"
		}
		config.Panes = append(config.Panes, PaneDef{
			Direction: dir,
			Size:      size,
			Command:   p.Command,
		})
	}
	return config, nil
}

// parseTmuxinator reads the tmuxinator format. It takes the FIRST window's
// pane list and layout. Each pane entry becomes an extra pane alongside Claude.
func parseTmuxinator(data []byte, path string) (*Config, error) {
	var root struct {
		Windows []yaml.Node `yaml:"windows"`
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse tmuxinator config %q: %w", path, err)
	}
	if len(root.Windows) == 0 {
		return &Config{}, nil
	}

	// Each windows entry is a mapping node with one key (the window name).
	// Decode the first window.
	firstWindowNode := root.Windows[0]
	if firstWindowNode.Kind != yaml.MappingNode || len(firstWindowNode.Content) < 2 {
		return &Config{}, nil
	}
	// Content[0] = key (window name), Content[1] = value (window config)
	var win tmuxinatorWindow
	if err := firstWindowNode.Content[1].Decode(&win); err != nil {
		return nil, fmt.Errorf("failed to decode tmuxinator window config: %w", err)
	}

	config := &Config{Layout: win.Layout}
	for _, paneEntry := range win.Panes {
		cmd := parseTmuxinatorPaneEntry(paneEntry)
		config.Panes = append(config.Panes, PaneDef{Command: cmd})
	}
	return config, nil
}

// parseTmuxinatorPaneEntry handles the various forms a tmuxinator pane can take:
//   - a plain string: "vim ."
//   - a single-key map: { pane_name: "command" }
func parseTmuxinatorPaneEntry(entry interface{}) string {
	switch v := entry.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Single-key map: { name: command }
		for _, val := range v {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}

// GenerateSetupScript returns a bash snippet that creates the additional panes.
// The snippet uses $MAIN_PANE and $WORKTREE_PATH variables which must be set
// in the calling script before this snippet runs.
func (c *Config) GenerateSetupScript() string {
	if len(c.Panes) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("# Create additional window panes (from window layout config)\n")

	for _, pane := range c.Panes {
		if c.Layout != "" {
			// When a layout will be applied, skip direction/size — let tmux arrange them.
			if pane.Command == "" {
				sb.WriteString(`tmux split-window -t "$MAIN_PANE" -c "$WORKTREE_PATH" 2>/dev/null || true` + "\n")
			} else {
				escaped := escapeForDoubleQuotes(pane.Command)
				sb.WriteString(fmt.Sprintf(`tmux split-window -t "$MAIN_PANE" -c "$WORKTREE_PATH" "%s" 2>/dev/null || true`+"\n", escaped))
			}
		} else {
			if pane.Command == "" {
				sb.WriteString(fmt.Sprintf(
					`tmux split-window -%s -t "$MAIN_PANE" -c "$WORKTREE_PATH" -l %s 2>/dev/null || true`+"\n",
					pane.Direction, pane.Size,
				))
			} else {
				escaped := escapeForDoubleQuotes(pane.Command)
				sb.WriteString(fmt.Sprintf(
					`tmux split-window -%s -t "$MAIN_PANE" -c "$WORKTREE_PATH" -l %s "%s" 2>/dev/null || true`+"\n",
					pane.Direction, pane.Size, escaped,
				))
			}
		}
	}

	if c.Layout != "" {
		sb.WriteString(fmt.Sprintf(`tmux select-layout -t "$MAIN_PANE" %s 2>/dev/null || true`+"\n", c.Layout))
	}
	sb.WriteString(`tmux select-pane -t "$MAIN_PANE"` + "\n")
	return sb.String()
}

// escapeForDoubleQuotes escapes backslashes and double-quotes so the string can
// be placed inside a double-quoted bash argument. Shell variables (e.g.
// $WORKTREE_PATH) are intentionally left unescaped so they expand at runtime.
func escapeForDoubleQuotes(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
