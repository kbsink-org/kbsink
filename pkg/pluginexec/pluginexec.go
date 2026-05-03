// Package pluginexec runs external kbsink CLI plugins as subprocesses (protocol v1).
//
// Parent kb-sink-md invokes a binary named kbsink-plugin-<name> found on PATH,
// or overridden by KBSINK_PLUGIN_<NAME> where NAME is the uppercased plugin id (e.g. KBSINK_PLUGIN_DOUYIN).
//
// Protocol v1 — parent → child (stdin, one JSON object, UTF-8):
//   article_url (string, required), output_root (string, required),
//   video_mode (string, "link" or "embed"), timeout (string, optional, Go duration).
//
// Protocol v1 — child → parent (stdout, one line JSON object):
//   Success: {"ok":true,"title":"...","markdown_path":"...","images":N,"videos":M}
//   Failure: {"ok":false,"error":"..."} with non-zero exit code.
package pluginexec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Request is the JSON body sent on stdin to kbsink-plugin-<name>.
type Request struct {
	ArticleURL string `json:"article_url"`
	OutputRoot string `json:"output_root"`
	VideoMode  string `json:"video_mode,omitempty"`
	Timeout    string `json:"timeout,omitempty"`
}

// Result is the successful outcome returned from a plugin subprocess.
type Result struct {
	Title        string `json:"title"`
	MarkdownPath string `json:"markdown_path"`
	Images       int    `json:"images"`
	Videos       int    `json:"videos"`
}

type responseEnvelope struct {
	OK            bool   `json:"ok"`
	Title         string `json:"title"`
	MarkdownPath  string `json:"markdown_path"`
	Images        int    `json:"images"`
	Videos        int    `json:"videos"`
	Error         string `json:"error"`
}

// ErrNotFound means no external plugin binary was resolved for the name.
var ErrNotFound = errors.New("pluginexec: no external plugin binary found")

// LookPluginBinary returns the executable path for plugin name (already normalized, e.g. "douyin").
// It checks KBSINK_PLUGIN_<UPPERNAME> first, then exec.LookPath("kbsink-plugin-"+name).
func LookPluginBinary(name string) (string, error) {
	n := strings.TrimSpace(strings.ToLower(name))
	if n == "" {
		return "", ErrNotFound
	}
	envKey := "KBSINK_PLUGIN_" + strings.ToUpper(strings.ReplaceAll(n, "-", "_"))
	if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
		st, err := os.Stat(p)
		if err != nil {
			return "", fmt.Errorf("pluginexec: %s is set but not usable: %w", envKey, err)
		}
		if st.IsDir() {
			return "", fmt.Errorf("pluginexec: %s points to a directory", envKey)
		}
		return filepath.Clean(p), nil
	}
	exeName := "kbsink-plugin-" + n
	path, err := exec.LookPath(exeName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return path, nil
}

// RunConvert starts the plugin binary, writes req as JSON to stdin, and parses stdout JSON.
func RunConvert(ctx context.Context, pluginPath string, req Request) (*Result, error) {
	if strings.TrimSpace(req.ArticleURL) == "" {
		return nil, fmt.Errorf("pluginexec: article_url is required")
	}
	if strings.TrimSpace(req.OutputRoot) == "" {
		return nil, fmt.Errorf("pluginexec: output_root is required")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("pluginexec: marshal request: %w", err)
	}

	cmd := exec.CommandContext(ctx, pluginPath)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	out := bytes.TrimSpace(stdout.Bytes())

	var env responseEnvelope
	if len(out) > 0 {
		if uerr := json.Unmarshal(out, &env); uerr != nil {
			if runErr != nil {
				return nil, fmt.Errorf("pluginexec: %w; stderr: %s", runErr, strings.TrimSpace(stderr.String()))
			}
			return nil, fmt.Errorf("pluginexec: invalid plugin stdout JSON: %w", uerr)
		}
	}

	if runErr != nil {
		if env.OK || env.Error != "" {
			msg := strings.TrimSpace(env.Error)
			if msg == "" {
				msg = runErr.Error()
			}
			return nil, fmt.Errorf("pluginexec: %s", msg)
		}
		s := strings.TrimSpace(stderr.String())
		if s != "" {
			return nil, fmt.Errorf("pluginexec: %w: %s", runErr, s)
		}
		return nil, fmt.Errorf("pluginexec: %w", runErr)
	}

	if !env.OK {
		msg := strings.TrimSpace(env.Error)
		if msg == "" {
			msg = "plugin returned ok=false"
		}
		return nil, fmt.Errorf("pluginexec: %s", msg)
	}

	return &Result{
		Title:        env.Title,
		MarkdownPath: env.MarkdownPath,
		Images:       env.Images,
		Videos:       env.Videos,
	}, nil
}

// FormatTimeout returns a duration string for Request.Timeout, or empty if non-positive.
func FormatTimeout(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.String()
}
