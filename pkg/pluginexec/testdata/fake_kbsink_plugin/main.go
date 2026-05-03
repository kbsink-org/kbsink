// Fake external plugin for pluginexec tests: reads JSON Request from stdin, writes fixed JSON to stdout.
package main

import (
	"encoding/json"
	"io"
	"os"
)

type request struct {
	ArticleURL string `json:"article_url"`
	OutputRoot string `json:"output_root"`
	VideoMode  string `json:"video_mode"`
}

func main() {
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(2)
	}
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		_, _ = os.Stdout.WriteString(`{"ok":false,"error":"bad json"}` + "\n")
		os.Exit(1)
	}
	if req.ArticleURL == "error:" {
		_, _ = os.Stdout.WriteString(`{"ok":false,"error":"simulated failure"}` + "\n")
		os.Exit(1)
	}
	out := `{"ok":true,"title":"Fake Title","markdown_path":"out/fake/fake.md","images":2,"videos":1}` + "\n"
	_, _ = os.Stdout.WriteString(out)
}
