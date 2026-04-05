package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantContains []string
	}{
		{
			name:         "normal output",
			args:         []string{"--hosts", "100", "--deadline", "2026-04-10", "--today", "2026-04-06", "--once"},
			wantContains: []string{"5 weekdays remaining", "20 hosts/night", "100 hosts"},
		},
		{
			name:         "deadline passed",
			args:         []string{"--hosts", "100", "--deadline", "2026-04-01", "--today", "2026-04-05", "--once"},
			wantContains: []string{"Deadline passed", "100 hosts remaining"},
		},
		{
			name:         "large host count has commas",
			args:         []string{"--hosts", "10000", "--deadline", "2026-04-10", "--today", "2026-04-06", "--once"},
			wantContains: []string{"10,000 hosts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.args, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run() returned error: %v", err)
			}
			out := stdout.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q\ngot: %s", want, out)
				}
			}
		})
	}
}

func TestRunJSON(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "normal json output",
			args: []string{"--hosts", "100", "--deadline", "2026-04-10", "--today", "2026-04-06", "--json"},
			want: map[string]any{
				"hosts_per_night": float64(20),
				"weekdays_left":   float64(5),
				"total_hosts":     float64(100),
				"deadline":        "2026-04-10",
				"today":           "2026-04-06",
				"deadline_passed": false,
			},
		},
		{
			name: "deadline passed json",
			args: []string{"--hosts", "100", "--deadline", "2026-04-01", "--today", "2026-04-05", "--json"},
			want: map[string]any{
				"hosts_per_night": float64(0),
				"weekdays_left":   float64(0),
				"total_hosts":     float64(100),
				"deadline":        "2026-04-01",
				"today":           "2026-04-05",
				"deadline_passed": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.args, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run() returned error: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON output: %v\nraw: %s", err, stdout.String())
			}

			for k, want := range tt.want {
				if got[k] != want {
					t.Errorf("json[%q] = %v, want %v", k, got[k], want)
				}
			}
		})
	}
}

func TestRunValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing hosts",
			args:    []string{"--deadline", "2026-04-10", "--once"},
			wantErr: "--hosts is required",
		},
		{
			name:    "zero hosts",
			args:    []string{"--hosts", "0", "--deadline", "2026-04-10", "--once"},
			wantErr: "--hosts is required",
		},
		{
			name:    "negative hosts via flag",
			args:    []string{"--hosts", "-5", "--deadline", "2026-04-10", "--once"},
			wantErr: "value must be positive, got -5",
		},
		{
			name:    "missing deadline",
			args:    []string{"--hosts", "100", "--once"},
			wantErr: "--deadline is required",
		},
		{
			name:    "bad deadline format",
			args:    []string{"--hosts", "100", "--deadline", "not-a-date", "--once"},
			wantErr: "invalid deadline format",
		},
		{
			name:    "bad today format",
			args:    []string{"--hosts", "100", "--deadline", "2026-04-10", "--today", "nope", "--once"},
			wantErr: "invalid today format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(--version) returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "tick") {
		t.Errorf("version output should contain 'tick', got: %s", stdout.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--help"}, &stdout, &stderr)
	// flag.ErrHelp is returned by FlagSet with ContinueOnError
	if err == nil {
		t.Fatal("expected error for --help")
	}
	if !strings.Contains(stderr.String(), "Examples:") {
		t.Error("help output should contain usage examples")
	}
}

func TestRunHostsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("500\n"), 0o644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--hosts-file", path, "--deadline", "2026-04-10", "--today", "2026-04-06", "--once"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() returned error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "500 hosts") {
		t.Errorf("output should contain '500 hosts', got: %s", out)
	}
}

func TestRunHostsFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("300\n"), 0o644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--hosts-file", path, "--deadline", "2026-04-10", "--today", "2026-04-06", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() returned error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["total_hosts"] != float64(300) {
		t.Errorf("total_hosts = %v, want 300", got["total_hosts"])
	}
}

func TestRunHostsFileMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("100\n"), 0o644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--hosts", "100", "--hosts-file", path, "--deadline", "2026-04-10", "--once"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for --hosts + --hosts-file")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error = %q, want it to mention 'mutually exclusive'", err.Error())
	}
}

func TestRunHostsFileMissing(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--hosts-file", "/nonexistent/hosts.txt", "--deadline", "2026-04-10", "--once"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing hosts file")
	}
}

func TestRunHostsFileBadContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("not-a-number\n"), 0o644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--hosts-file", path, "--deadline", "2026-04-10", "--once"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid hosts file content")
	}
}
