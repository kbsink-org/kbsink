// Package cli implements the kb-sink-md flag parsing and conversion entrypoint.
// It lives next to the main package so kb-sink-md and other binaries can share one implementation.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/kbsink-org/kbsink/internal/plugin" // register built-in plugins (wechat, xhs)
	kbsink "github.com/kbsink-org/kbsink/pkg"
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/pluginexec"
	"github.com/kbsink-org/kbsink/pkg/pluginreg"
)

// Run parses argv like kb-sink-md: flags then a single URL argument.
// Exit codes: 0 success, 1 error, 2 usage.
func Run(args []string) int {
	if len(args) < 1 {
		return 2
	}
	prog := filepath.Base(args[0])
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	outputRoot := fs.String("o", "output", "output root directory (local filesystem; kb-sink-md does not configure S3)")
	timeout := fs.Duration("timeout", 60*time.Second, "timeout for the conversion")
	videoMode := fs.String("video-mode", "link", "video markdown mode: link|embed")
	plugin := fs.String("plugin", "", "required: registered plugin name (parser + driver), e.g. wechat, xhs")

	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage:\n  %s --plugin <name> [other flags] <article-url>\n\nFlags:\n", prog)
		fs.PrintDefaults()
		_, _ = fmt.Fprintf(fs.Output(), "\n--plugin is required (no URL-based auto selection). Storage is always local disk in this binary.\n")
		if names := pluginreg.Names(); len(names) > 0 {
			_, _ = fmt.Fprintf(fs.Output(), "Plugins in this build: %s\n", strings.Join(names, ", "))
		}
		_, _ = fmt.Fprintf(fs.Output(), "Other plugin ids may work if a matching kbsink-plugin-<id> binary is on PATH (or set KBSINK_PLUGIN_<ID> to its path).\n")
	}

	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	articleURL := fs.Arg(0)

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	mode, err := resolveVideoMode(*videoMode)
	if err != nil {
		emitError(err)
		return 1
	}

	pluginName := strings.TrimSpace(*plugin)
	if pluginName == "" {
		emitError(fmt.Errorf("--plugin is required (e.g. --plugin wechat); plugins in this build: %s",
			strings.Join(pluginreg.Names(), ", ")))
		return 1
	}
	pluginName = strings.ToLower(pluginName)

	httpClient := httpClientForCLI(*timeout)

	pl, ok := pluginreg.Lookup(pluginName)
	if ok {
		parser, driver, err := pl.NewComponents(httpClient)
		if err != nil {
			emitError(fmt.Errorf("plugin %q: %w", pluginName, err))
			return 1
		}
		if parser == nil {
			emitError(fmt.Errorf("plugin %q returned nil parser", pluginName))
			return 1
		}

		opts := []kbsink.Option{
			kbsink.WithHTTPClient(httpClient),
			kbsink.WithParser(parser),
		}
		if driver != nil {
			opts = append(opts, kbsink.WithDriver(driver))
		}
		converter := kbsink.NewConverter(opts...)

		res, err := converter.Convert(ctx, articleURL, core.ConvertOptions{
			OutputRoot: *outputRoot,
			VideoMode:  mode,
		})
		if err != nil {
			emitError(err)
			return 1
		}

		emitSuccess(res)
		return 0
	}

	binPath, err := pluginexec.LookPluginBinary(pluginName)
	if err != nil {
		if errors.Is(err, pluginexec.ErrNotFound) {
			emitError(fmt.Errorf("unknown plugin %q; registered: %s; external: install kbsink-plugin-%s on PATH or set KBSINK_PLUGIN_%s",
				pluginName, strings.Join(pluginreg.Names(), ", "), pluginName, pluginEnvSuffix(pluginName)))
			return 1
		}
		emitError(err)
		return 1
	}

	extRes, err := pluginexec.RunConvert(ctx, binPath, pluginexec.Request{
		ArticleURL: articleURL,
		OutputRoot: *outputRoot,
		VideoMode:  string(mode),
		Timeout:    pluginexec.FormatTimeout(*timeout),
	})
	if err != nil {
		emitError(err)
		return 1
	}
	emitSuccessExternal(extRes)
	return 0
}

func pluginEnvSuffix(pluginName string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(pluginName), "-", "_"))
}

func httpClientForCLI(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		return http.DefaultClient
	}
	return &http.Client{Timeout: timeout}
}

func emitError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
}

func emitSuccess(res *core.ArticleResult) {
	_, _ = fmt.Fprintf(os.Stdout, "title: %s\n", res.Title)
	_, _ = fmt.Fprintf(os.Stdout, "markdown: %s\n", res.MarkdownPath)
	_, _ = fmt.Fprintf(os.Stdout, "images: %d\n", len(res.Images))
	videoCount := 0
	for _, asset := range res.Assets {
		if asset.Type == core.AssetTypeVideo {
			videoCount++
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "videos: %d\n", videoCount)
}

func emitSuccessExternal(res *pluginexec.Result) {
	_, _ = fmt.Fprintf(os.Stdout, "title: %s\n", res.Title)
	_, _ = fmt.Fprintf(os.Stdout, "markdown: %s\n", res.MarkdownPath)
	_, _ = fmt.Fprintf(os.Stdout, "images: %d\n", res.Images)
	_, _ = fmt.Fprintf(os.Stdout, "videos: %d\n", res.Videos)
}

func resolveVideoMode(raw string) (core.VideoMode, error) {
	mode := strings.TrimSpace(strings.ToLower(raw))
	switch mode {
	case "", string(core.VideoModeLink):
		return core.VideoModeLink, nil
	case string(core.VideoModeEmbed):
		return core.VideoModeEmbed, nil
	default:
		return "", fmt.Errorf("unsupported video mode %q, expected link|embed", raw)
	}
}
