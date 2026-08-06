// Command starlight is the golden-path CLI for scaffolding a new
// microservice into this repository's GitOps layout.
package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed templates/chart templates/app.yaml
var templatesFS embed.FS

const placeholderName = "sample-service"

var serviceNamePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// isValidServiceName reports whether name is a valid kebab-case Kubernetes
// resource name: lowercase alphanumeric characters and hyphens only, per
// the naming convention in CLAUDE.md.
func isValidServiceName(name string) bool {
	return name != "" && serviceNamePattern.MatchString(name)
}

// promptServiceName reads a service name from r, re-prompting on w until a
// valid kebab-case name is entered.
func promptServiceName(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "Service name: ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		name := strings.TrimSpace(line)
		if isValidServiceName(name) {
			return name, nil
		}
		fmt.Fprintf(w, "Invalid service name %q: use lowercase letters, numbers, and hyphens only (kebab-case), e.g. \"order-events-consumer\".\n", name)
	}
}

// writeTemplateDir walks the embedded chart template and writes it to
// destDir, substituting every occurrence of the placeholder service name
// with serviceName.
func writeTemplateDir(destDir, serviceName string) error {
	const embedRoot = "templates/chart"
	return fs.WalkDir(templatesFS, embedRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, embedRoot), "/")
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := templatesFS.ReadFile(p)
		if err != nil {
			return err
		}
		rendered := strings.ReplaceAll(string(content), placeholderName, serviceName)
		return os.WriteFile(target, []byte(rendered), 0o644)
	})
}

// writeApplicationManifest renders the embedded Argo CD Application
// template for serviceName and writes it to destPath.
func writeApplicationManifest(destPath, serviceName string) error {
	content, err := templatesFS.ReadFile("templates/app.yaml")
	if err != nil {
		return err
	}
	rendered := strings.ReplaceAll(string(content), placeholderName, serviceName)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, []byte(rendered), 0o644)
}

func runCreateService(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("create-service", flag.ContinueOnError)
	flags.SetOutput(stderr)
	nameFlag := flags.String("name", "", "service name in kebab-case (skips the interactive prompt)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	var serviceName string
	if *nameFlag != "" {
		if !isValidServiceName(*nameFlag) {
			return fmt.Errorf("invalid -name %q: use lowercase letters, numbers, and hyphens only (kebab-case)", *nameFlag)
		}
		serviceName = *nameFlag
	} else {
		name, err := promptServiceName(bufio.NewReader(stdin), stdout)
		if err != nil {
			return fmt.Errorf("reading service name: %w", err)
		}
		serviceName = name
	}

	chartDir := filepath.Join("platform", "charts", serviceName)
	appPath := filepath.Join("platform", "gitops", serviceName+"-app.yaml")

	if err := writeTemplateDir(chartDir, serviceName); err != nil {
		return fmt.Errorf("generating chart: %w", err)
	}
	if err := writeApplicationManifest(appPath, serviceName); err != nil {
		return fmt.Errorf("generating Argo CD Application manifest: %w", err)
	}

	fmt.Fprintf(stdout, `
Generated service %q:
  - %s
  - %s

This CLI only generates files — nothing has been committed, pushed, or
applied to the cluster. Review the generated files, then:

  1. git add %s %s
  2. git commit -m "feat: add %s service"
  3. git push
  4. kubectl apply -f %s   # registers the service with Argo CD

`, serviceName, chartDir, appPath, chartDir, appPath, serviceName, appPath)

	return nil
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "create-service" {
		fmt.Fprintln(os.Stderr, "usage: starlight create-service [-name <service-name>]")
		os.Exit(1)
	}

	if err := runCreateService(os.Args[2:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
