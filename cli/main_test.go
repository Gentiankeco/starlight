package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsValidServiceName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple lowercase", "payments", true},
		{"kebab-case", "order-events-consumer", true},
		{"alphanumeric with hyphen", "service-2", true},
		{"empty string rejected", "", false},
		{"uppercase rejected", "Payments-Api", false},
		{"spaces rejected", "payments api", false},
		{"underscore rejected", "payments_api", false},
		{"special characters rejected", "payments@api!", false},
		{"leading hyphen rejected", "-payments", false},
		{"trailing hyphen rejected", "payments-", false},
		{"double hyphen rejected", "payments--api", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isValidServiceName(c.input)
			if got != c.want {
				t.Errorf("isValidServiceName(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestPromptServiceNameRePromptsOnInvalidInput(t *testing.T) {
	input := "Payments Api\npayments-api\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer

	got, err := promptServiceName(r, &out)
	if err != nil {
		t.Fatalf("promptServiceName returned error: %v", err)
	}
	if got != "payments-api" {
		t.Errorf("promptServiceName() = %q, want %q", got, "payments-api")
	}
	if !strings.Contains(out.String(), "Invalid service name") {
		t.Errorf("expected re-prompt message for invalid input, got output: %q", out.String())
	}
}

func TestPromptImageAcceptsValidInput(t *testing.T) {
	input := "ghcr.io/gentiankeco/starlight-demo-app:latest\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer

	got, err := promptImage(r, &out)
	if err != nil {
		t.Fatalf("promptImage returned error: %v", err)
	}
	if got != "ghcr.io/gentiankeco/starlight-demo-app:latest" {
		t.Errorf("promptImage() = %q, want %q", got, "ghcr.io/gentiankeco/starlight-demo-app:latest")
	}
}

func TestPromptImageRePromptsOnEmptyInput(t *testing.T) {
	input := "\n   \nnginx:latest\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer

	got, err := promptImage(r, &out)
	if err != nil {
		t.Fatalf("promptImage returned error: %v", err)
	}
	if got != "nginx:latest" {
		t.Errorf("promptImage() = %q, want %q", got, "nginx:latest")
	}
	if strings.Count(out.String(), "must not be empty") != 2 {
		t.Errorf("expected two re-prompt messages for blank input, got output: %q", out.String())
	}
}

func TestSplitImageRef(t *testing.T) {
	cases := []struct {
		name           string
		image          string
		wantRepository string
		wantTag        string
	}{
		{"simple image with tag", "nginx:latest", "nginx", "latest"},
		{"registry path with tag", "ghcr.io/you/your-app:v1", "ghcr.io/you/your-app", "v1"},
		{"multi-segment path with tag", "ghcr.io/gentiankeco/starlight-demo-app:latest", "ghcr.io/gentiankeco/starlight-demo-app", "latest"},
		{"no tag defaults to latest", "nginx", "nginx", "latest"},
		{"registry with port and no tag", "localhost:5000/myapp", "localhost:5000/myapp", "latest"},
		{"registry with port and tag", "localhost:5000/myapp:v2", "localhost:5000/myapp", "v2"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotRepository, gotTag := splitImageRef(c.image)
			if gotRepository != c.wantRepository || gotTag != c.wantTag {
				t.Errorf("splitImageRef(%q) = (%q, %q), want (%q, %q)", c.image, gotRepository, gotTag, c.wantRepository, c.wantTag)
			}
		})
	}
}

func TestRepoRootFromFindsGitRootRegardlessOfStartingDepth(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate this project's real layout: a nested Go module (like
	// cli/go.mod) that sits below the true repo root and must not be
	// mistaken for it.
	nested := filepath.Join(root, "cli", "sub", "deeper")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cli", "go.mod"), []byte("module fake\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	wantRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, start := range []string{root, filepath.Join(root, "cli"), nested} {
		t.Run(start, func(t *testing.T) {
			got, err := repoRootFrom(start)
			if err != nil {
				t.Fatalf("repoRootFrom(%q) returned error: %v", start, err)
			}
			gotRoot, err := filepath.EvalSymlinks(got)
			if err != nil {
				t.Fatal(err)
			}
			if gotRoot != wantRoot {
				t.Errorf("repoRootFrom(%q) = %q, want %q", start, gotRoot, wantRoot)
			}
		})
	}
}
