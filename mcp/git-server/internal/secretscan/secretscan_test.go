package secretscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFilesDetectsSecrets(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		wantRule string
	}{
		{"aws access key", "deploy uses AKIAIOSFODNN7EXAMPLE for s3\n", "aws-access-key-id"},
		{"private key block", "-----BEGIN OPENSSH PRIVATE KEY-----\nbody\n", "private-key-block"},
		{"gitlab pat", "token: glpat-ABCDEFGHIJ1234567890\n", "gitlab-pat"},
		{"github pat", "GH_TOKEN=ghp_0123456789abcdefghijklmnopqrstuvwxyz\n", "github-pat"},
		{"slack token", "slack = xoxb-1234567890-abcdef\n", "slack-token"},
		{"google api key", "key AIzaabcdefghijklmnopqrstuvwxyz012345678 used\n", "google-api-key"},
		{"jwt", "auth eyJhbGciOiJIUzI1.eyJzdWIiOiIxMjM0.dBjftJeZ4CVP\n", "jwt"},
		{"generic assignment", "password = \"hunter2hunter2\"\n", "generic-secret-assignment"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, "memory.md", c.content)

			findings, err := ScanFiles(root, []string{"memory.md"})
			if err != nil {
				t.Fatalf("ScanFiles: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("findings = %v, want exactly 1", findings)
			}
			if findings[0].Rule != c.wantRule {
				t.Fatalf("rule = %q, want %q", findings[0].Rule, c.wantRule)
			}
			if findings[0].Line != 1 {
				t.Fatalf("line = %d, want 1", findings[0].Line)
			}
			// The matched secret value must never leak into the finding.
			if strings.Contains(findings[0].String(), strings.TrimSpace(c.content)) {
				t.Fatalf("finding %q leaked the secret content", findings[0].String())
			}
		})
	}
}

func TestScanFilesCleanContent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "notes.md", "# Notes\n\nNothing sensitive here, just prose about AWS and tokens.\n")

	findings, err := ScanFiles(root, []string{"notes.md"})
	if err != nil {
		t.Fatalf("ScanFiles: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
}

func TestScanFilesSkipsMissingDirAndBinary(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Binary content with a NUL byte and a key-shaped string that must be ignored.
	writeFile(t, root, "blob.bin", "AKIAIOSFODNN7EXAMPLE\x00\x01\x02")

	findings, err := ScanFiles(root, []string{"does-not-exist.md", "subdir", "blob.bin"})
	if err != nil {
		t.Fatalf("ScanFiles: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none (missing/dir/binary all skipped)", findings)
	}
}

func TestScanFilesReportsLineAndPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "creds.md", "line one\nline two\nslack = xoxb-1234567890-abcdef\n")

	findings, err := ScanFiles(root, []string{"creds.md"})
	if err != nil {
		t.Fatalf("ScanFiles: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %v, want 1", findings)
	}
	if got := findings[0]; got.Path != "creds.md" || got.Line != 3 {
		t.Fatalf("finding = %s, want creds.md:3", got)
	}
}

func TestBlockedErrorMessageIsActionableAndSafe(t *testing.T) {
	err := &BlockedError{Findings: []Finding{
		{Path: "memory.md", Rule: "gitlab-pat", Line: 4},
	}}
	msg := err.Error()
	for _, want := range []string{"refusing to commit", "memory.md:4", "gitlab-pat", ".gitignore"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error message %q missing %q", msg, want)
		}
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
