package fs

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolverResolveGithubDependency(t *testing.T) {
	t.Helper()

	archive := buildTestZip(t, map[string]string{
		"repo-v1.2.3/pkg/Nargo.toml": "[package]\nname = \"dep\"\ntype = \"lib\"\n\n[dependencies]\n",
		"repo-v1.2.3/pkg/src/lib.nr": "fn dep() {}",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner/repo/archive/v1.2.3.zip" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "Nargo.toml"), []byte(strings.Join([]string{
		"[package]",
		"name = \"root\"",
		"type = \"bin\"",
		"",
		"[dependencies.dep]",
		"git = \"https://github.com/owner/repo\"",
		"tag = \"v1.2.3\"",
		"directory = \"pkg\"",
		"",
	}, "\n")), 0644)
	if err != nil {
		t.Fatalf("write root Nargo.toml: %v", err)
	}
	err = os.MkdirAll(filepath.Join(root, "src"), 0755)
	if err != nil {
		t.Fatalf("create src dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(root, "src", "main.nr"), []byte("fn main() {}"), 0644)
	if err != nil {
		t.Fatalf("write root main.nr: %v", err)
	}

	client := &http.Client{
		Transport: rewriteGithubTransport(t, server.URL),
	}
	r := newResolverForTest(t.TempDir(), client)

	if err := r.Resolve(root); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	var foundRemoteLib bool
	for path, content := range r.AllFiles {
		if strings.HasSuffix(path, filepath.Join("pkg", "src", "lib.nr")) {
			foundRemoteLib = true
			if content != "fn dep() {}" {
				t.Fatalf("unexpected remote lib content: %q", content)
			}
		}
	}
	if !foundRemoteLib {
		t.Fatalf("expected remote dependency lib.nr to be loaded")
	}
}

func buildTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	zw := zip.NewWriter(buf)
	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func rewriteGithubTransport(t *testing.T, baseURL string) http.RoundTripper {
	t.Helper()

	serverURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "github.com" {
			cloned := req.Clone(req.Context())
			cloned.URL.Scheme = serverURL.Scheme
			cloned.URL.Host = serverURL.Host
			cloned.Host = serverURL.Host
			req = cloned
		}
		return http.DefaultTransport.RoundTrip(req)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResolveGithubCodeArchiveRejectsTraversal(t *testing.T) {
	_, err := resolveGithubCodeArchive("https://github.com/owner/repo", "../bad")
	if err == nil {
		t.Fatalf("expected invalid ref error")
	}
	if !strings.Contains(err.Error(), "invalid git reference") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSafeFilename(t *testing.T) {
	got := safeFilename("/owner/repo/archive/v1.2.3.zip")
	want := "owner_repo_archive_v1.2.3.zip"
	if got != want {
		t.Fatalf("safeFilename mismatch: got %q want %q", got, want)
	}
}
