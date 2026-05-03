package pluginexec

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLookPluginBinary_envOverride(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "my-douyin")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KBSINK_PLUGIN_DOUYIN", fake)
	got, err := LookPluginBinary("douyin")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %q want %q", got, fake)
	}
}

func TestLookPluginBinary_notOnPath_errorsIsNotFound(t *testing.T) {
	t.Parallel()
	_, err := LookPluginBinary("zzzz_no_such_plugin_name_ever")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound in chain, got %v", err)
	}
}

func TestLookPluginBinary_envOverride_missingFile(t *testing.T) {
	t.Setenv("KBSINK_PLUGIN_DOUYIN", filepath.Join(t.TempDir(), "nope-not-here"))
	_, err := LookPluginBinary("douyin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "KBSINK_PLUGIN_DOUYIN") {
		t.Fatalf("got %v", err)
	}
}

func TestRunConvert_success(t *testing.T) {
	t.Parallel()
	bin := buildFakePlugin(t)
	ctx := context.Background()
	res, err := RunConvert(ctx, bin, Request{
		ArticleURL: "https://example.com/x",
		OutputRoot: "output",
		VideoMode:  "link",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "Fake Title" {
		t.Fatalf("title %q", res.Title)
	}
	if res.MarkdownPath != "out/fake/fake.md" {
		t.Fatalf("markdown_path %q", res.MarkdownPath)
	}
	if res.Images != 2 || res.Videos != 1 {
		t.Fatalf("images=%d videos=%d", res.Images, res.Videos)
	}
}

func TestRunConvert_pluginErrorJSON(t *testing.T) {
	t.Parallel()
	bin := buildFakePlugin(t)
	ctx := context.Background()
	_, err := RunConvert(ctx, bin, Request{
		ArticleURL: "error:",
		OutputRoot: "output",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "simulated failure") {
		t.Fatalf("got %v", err)
	}
}

func TestRunConvert_contextCancel(t *testing.T) {
	t.Parallel()
	bin := buildFakePlugin(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := RunConvert(ctx, bin, Request{
		ArticleURL: "https://slow.example/",
		OutputRoot: "output",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunConvert_validation(t *testing.T) {
	t.Parallel()
	bin := buildFakePlugin(t)
	ctx := context.Background()
	if _, err := RunConvert(ctx, bin, Request{ArticleURL: "", OutputRoot: "o"}); err == nil {
		t.Fatal("want error for empty article_url")
	}
	if _, err := RunConvert(ctx, bin, Request{ArticleURL: "u", OutputRoot: ""}); err == nil {
		t.Fatal("want error for empty output_root")
	}
}

func buildFakePlugin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "fake_kbsink_plugin"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(dir, name)
	srcDir, err := filepath.Abs(filepath.Join("testdata", "fake_kbsink_plugin"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, outb)
	}
	return out
}
