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
	"strconv"
	"strings"
)

//go:embed templates/chart templates/app.yaml templates/ci-workflow.yml
var templatesFS embed.FS

const placeholderName = "sample-service"
const placeholderImageRepository = "nginx"
const placeholderImageTag = "1.27"
const placeholderContainerPort = "80"
const defaultContainerPort = "8080"
const placeholderGitHubUsername = "sample-owner"

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

// promptImage reads a container image reference from r, re-prompting on w
// until a non-empty value is entered.
func promptImage(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "Container image (e.g. nginx:latest or ghcr.io/you/your-app:v1): ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		image := strings.TrimSpace(line)
		if image != "" {
			return image, nil
		}
		fmt.Fprintln(w, "Container image must not be empty.")
	}
}

// isValidContainerPort reports whether s is a valid TCP port number: an
// integer between 1 and 65535.
func isValidContainerPort(s string) bool {
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n >= 1 && n <= 65535
}

// promptContainerPort reads a container port from r, re-prompting on w
// until a valid port (1-65535) is entered. An empty line accepts the
// default of 8080.
func promptContainerPort(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "Container port (default 8080, press Enter to accept): ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		port := strings.TrimSpace(line)
		if port == "" {
			return defaultContainerPort, nil
		}
		if isValidContainerPort(port) {
			return port, nil
		}
		fmt.Fprintf(w, "Invalid container port %q: enter a number between 1 and 65535.\n", port)
	}
}

// isValidGitHubUsername reports whether s is a non-empty string containing
// no spaces, suitable for use in a ghcr.io image path.
func isValidGitHubUsername(s string) bool {
	return s != "" && !strings.Contains(s, " ")
}

// promptGitHubUsername reads a GitHub username or org from r, re-prompting
// on w until a valid (non-empty, space-free) value is entered.
func promptGitHubUsername(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "GitHub username or org (for ghcr.io image path, e.g. Gentiankeco): ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		username := strings.TrimSpace(line)
		if isValidGitHubUsername(username) {
			return username, nil
		}
		fmt.Fprintf(w, "Invalid GitHub username %q: must be non-empty and contain no spaces.\n", username)
	}
}

// splitImageRef splits a container image reference into its repository and
// tag, e.g. "ghcr.io/gentiankeco/starlight-demo-app:latest" becomes
// ("ghcr.io/gentiankeco/starlight-demo-app", "latest"). The split happens
// on the last colon that appears after the last slash, so registry hosts
// with a port (e.g. "localhost:5000/app") are not mistaken for a tag
// separator. If no tag is present, "latest" is returned.
func splitImageRef(image string) (repository, tag string) {
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon > lastSlash {
		return image[:lastColon], image[lastColon+1:]
	}
	return image, "latest"
}

// writeTemplateDir walks the embedded chart template and writes it to
// destDir, substituting every occurrence of the placeholder service name
// with serviceName, and — in values.yaml only — the placeholder image
// repository, tag, and container port with imageRepository, imageTag, and
// containerPort.
func writeTemplateDir(destDir, serviceName, imageRepository, imageTag, containerPort string) error {
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
		if rel == "values.yaml" {
			rendered = strings.Replace(rendered, "repository: "+placeholderImageRepository, "repository: "+imageRepository, 1)
			rendered = strings.Replace(rendered, `tag: "`+placeholderImageTag+`"`, `tag: "`+imageTag+`"`, 1)
			rendered = strings.Replace(rendered, "containerPort: "+placeholderContainerPort, "containerPort: "+containerPort, 1)
		}
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

// writeCIWorkflow renders the embedded GitHub Actions CI workflow template
// for serviceName and githubUsername and writes it to destPath.
func writeCIWorkflow(destPath, serviceName, githubUsername string) error {
	content, err := templatesFS.ReadFile("templates/ci-workflow.yml")
	if err != nil {
		return err
	}
	rendered := strings.ReplaceAll(string(content), placeholderName, serviceName)
	rendered = strings.ReplaceAll(rendered, placeholderGitHubUsername, githubUsername)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, []byte(rendered), 0o644)
}

// repoRootFrom walks up from startDir looking for a ".git" entry (a
// directory in a normal checkout, or a file in a git worktree), which
// marks the root of this repository. This is deliberately independent of
// cli/go.mod: the CLI's own module lives in a subdirectory (cli/), not at
// the repo root, so anchoring on go.mod would resolve to cli/ itself when
// the binary is run from inside that directory.
func repoRootFrom(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate repository root: no .git found above %s", startDir)
		}
		dir = parent
	}
}

func runCreateService(args []string, startDir string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("create-service", flag.ContinueOnError)
	flags.SetOutput(stderr)
	nameFlag := flags.String("name", "", "service name in kebab-case (skips the interactive prompt)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	stdinReader := bufio.NewReader(stdin)

	var serviceName string
	if *nameFlag != "" {
		if !isValidServiceName(*nameFlag) {
			return fmt.Errorf("invalid -name %q: use lowercase letters, numbers, and hyphens only (kebab-case)", *nameFlag)
		}
		serviceName = *nameFlag
	} else {
		name, err := promptServiceName(stdinReader, stdout)
		if err != nil {
			return fmt.Errorf("reading service name: %w", err)
		}
		serviceName = name
	}

	image, err := promptImage(stdinReader, stdout)
	if err != nil {
		return fmt.Errorf("reading container image: %w", err)
	}
	imageRepository, imageTag := splitImageRef(image)

	containerPort, err := promptContainerPort(stdinReader, stdout)
	if err != nil {
		return fmt.Errorf("reading container port: %w", err)
	}

	githubUsername, err := promptGitHubUsername(stdinReader, stdout)
	if err != nil {
		return fmt.Errorf("reading GitHub username: %w", err)
	}

	repoRoot, err := repoRootFrom(startDir)
	if err != nil {
		return err
	}

	chartDir := filepath.Join(repoRoot, "platform", "charts", serviceName)
	appPath := filepath.Join(repoRoot, "platform", "gitops", serviceName+"-app.yaml")
	ciPath := filepath.Join(repoRoot, "platform", "ci-templates", serviceName+"-ci.yml")

	if err := writeTemplateDir(chartDir, serviceName, imageRepository, imageTag, containerPort); err != nil {
		return fmt.Errorf("generating chart: %w", err)
	}
	if err := writeApplicationManifest(appPath, serviceName); err != nil {
		return fmt.Errorf("generating Argo CD Application manifest: %w", err)
	}
	if err := writeCIWorkflow(ciPath, serviceName, githubUsername); err != nil {
		return fmt.Errorf("generating CI workflow template: %w", err)
	}

	fmt.Fprintf(stdout, `
Generated service %q in repo %s:
  - %s
  - %s
  - %s

This CLI only generates files — nothing has been committed, pushed, or
applied to the cluster. Review the generated files, then:

  1. git add %s %s %s
  2. git commit -m "feat: add %s service"
  3. git push
  4. kubectl apply -f %s   # registers the service with Argo CD
  5. Copy %s into your app repo at .github/workflows/build-and-push.yml
     — or trigger the Starlight bootstrap workflow to deliver it
     automatically.

`, serviceName, repoRoot, chartDir, appPath, ciPath, chartDir, appPath, ciPath, serviceName, appPath, ciPath)

	return nil
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "create-service" {
		fmt.Fprintln(os.Stderr, "usage: starlight create-service [-name <service-name>]")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := runCreateService(os.Args[2:], cwd, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
