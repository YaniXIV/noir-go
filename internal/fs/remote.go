package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultGitRef = "HEAD"

func (r *Resolver) resolveRemoteDependency(dep map[string]string) (string, error) {
	gitURL := dep["git"]
	tag := dep["tag"]
	if tag == "" {
		tag = defaultGitRef
	}

	repoURL, err := url.Parse(gitURL)
	if err != nil {
		return "", fmt.Errorf("parse git dependency %q: %w", gitURL, err)
	}
	if repoURL.Host != "github.com" {
		return "", fmt.Errorf("unsupported git host %q", repoURL.Host)
	}

	archiveURL, err := resolveGithubCodeArchive(gitURL, tag)
	if err != nil {
		return "", err
	}

	archivePath, err := r.fetchArchive(archiveURL)
	if err != nil {
		return "", err
	}

	return r.extractArchive(repoURL, tag, dep["directory"], archivePath)
}

func (r *Resolver) fetchArchive(archiveURL *url.URL) (string, error) {
	archivePath := filepath.Join(r.cacheDir, "archives", safeFilename(archiveURL.Path))
	if fileExists(archivePath) {
		return archivePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return "", fmt.Errorf("create archive cache dir: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, archiveURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create archive request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download archive %q: %w", archiveURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download archive %q: unexpected status %s", archiveURL.String(), resp.Status)
	}

	tmpPath := archivePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp archive file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return "", fmt.Errorf("write archive %q: %w", tmpPath, err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close archive %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, archivePath); err != nil {
		return "", fmt.Errorf("move archive into cache: %w", err)
	}

	return archivePath, nil
}

func (r *Resolver) extractArchive(repoURL *url.URL, tag string, directory string, archivePath string) (string, error) {
	extractRoot := filepath.Join(r.cacheDir, "libs", safeFilename(repoURL.Path+"@"+tag))
	packagePath := filepath.Join(extractRoot, directory)
	if fileExists(packagePath) {
		return packagePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(extractRoot), 0755); err != nil {
		return "", fmt.Errorf("create extract cache dir: %w", err)
	}

	tmpRoot := extractRoot + ".tmp"
	if err := os.RemoveAll(tmpRoot); err != nil {
		return "", fmt.Errorf("reset temp extract dir: %w", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive %q: %w", archivePath, err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		name, err := stripArchiveRoot(f.Name)
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}

		target := filepath.Join(tmpRoot, filepath.FromSlash(name))
		if !strings.HasPrefix(target, tmpRoot+string(os.PathSeparator)) && target != tmpRoot {
			return "", fmt.Errorf("archive entry %q escapes extract root", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("create archive dir %q: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", fmt.Errorf("create parent dir for %q: %w", target, err)
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("open archive entry %q: %w", f.Name, err)
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
		if err != nil {
			rc.Close()
			return "", fmt.Errorf("create extracted file %q: %w", target, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return "", fmt.Errorf("extract archive entry %q: %w", f.Name, err)
		}
		if err := out.Close(); err != nil {
			rc.Close()
			return "", fmt.Errorf("close extracted file %q: %w", target, err)
		}
		if err := rc.Close(); err != nil {
			return "", fmt.Errorf("close archive entry %q: %w", f.Name, err)
		}
	}

	if err := os.Rename(tmpRoot, extractRoot); err != nil {
		return "", fmt.Errorf("move extracted dependency into cache: %w", err)
	}
	if !fileExists(packagePath) {
		return "", fmt.Errorf("dependency directory %q not found in archive", directory)
	}

	return packagePath, nil
}

func resolveGithubCodeArchive(rawGitURL string, ref string) (*url.URL, error) {
	repoURL, err := url.Parse(rawGitURL)
	if err != nil {
		return nil, fmt.Errorf("parse github url %q: %w", rawGitURL, err)
	}

	owner, repo, err := splitGithubRepo(repoURL)
	if err != nil {
		return nil, err
	}
	if err := validateGitRef(ref); err != nil {
		return nil, err
	}

	return &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   fmt.Sprintf("/%s/%s/archive/%s.zip", owner, repo, ref),
	}, nil
}

func splitGithubRepo(repoURL *url.URL) (string, string, error) {
	parts := strings.Split(strings.Trim(repoURL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid github repository url %q", repoURL.String())
	}
	repo := strings.TrimSuffix(parts[1], ".git")
	return parts[0], repo, nil
}

func validateGitRef(ref string) error {
	decoded, err := url.PathUnescape(ref)
	if err != nil {
		return fmt.Errorf("decode git ref %q: %w", ref, err)
	}
	if strings.Contains(decoded, "..") || strings.Contains(decoded, "/") || strings.Contains(decoded, `\`) {
		return fmt.Errorf("invalid git reference %q", ref)
	}
	return nil
}

func safeFilename(val string) string {
	if val == "" {
		panic("invalid value")
	}

	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	return strings.TrimLeft(replacer.Replace(val), "_")
}

func stripArchiveRoot(name string) (string, error) {
	clean := strings.Trim(strings.ReplaceAll(name, "\\", "/"), "/")
	if clean == "" {
		return "", nil
	}

	parts := strings.Split(clean, "/")
	if len(parts) < 2 {
		return "", nil
	}

	rest := strings.Join(parts[1:], "/")
	if rest == "." || rest == "" {
		return "", nil
	}
	if strings.Contains(rest, "..") {
		return "", fmt.Errorf("archive entry %q contains invalid traversal", name)
	}
	return rest, nil
}
