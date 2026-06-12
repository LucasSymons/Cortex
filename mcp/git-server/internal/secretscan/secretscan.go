// Package secretscan provides a lightweight, dependency-free content scan for
// high-signal secrets. It backs the server-side commit gate: changed files are
// scanned before a commit is created so a credential pasted into a memory
// file's body never reaches the remote.
//
// The ruleset is deliberately curated for low false positives rather than
// exhaustive coverage - the threat model is an accidental paste of a real
// credential into a profile/memory file, not adversarial obfuscation. Users who
// want a comprehensive ruleset can layer an optional gitleaks pre-commit hook on
// top (see the repository SECURITY.md); this in-server gate is the always-on
// backstop that needs no external tooling.
package secretscan

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	// maxFileSize caps how much of any single file is read. Profile and memory
	// files are small; anything larger is almost certainly a binary or data blob
	// and is skipped rather than scanned.
	maxFileSize = 5 << 20 // 5 MiB
	// maxLineLen bounds the scanner's per-line buffer. Base64-encoded keys can be
	// long, so allow generous lines without risking unbounded memory.
	maxLineLen = 1 << 20 // 1 MiB
)

// rule pairs a stable, human-readable name with a compiled detection pattern.
// The name is reported in findings; the matched text never is.
type rule struct {
	name    string
	pattern *regexp.Regexp
}

// rules is the curated high-signal ruleset. Each pattern targets a credential
// shape distinctive enough to keep false positives low on prose and config.
var rules = []rule{
	{"aws-access-key-id", regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{"private-key-block", regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`)},
	{"gitlab-pat", regexp.MustCompile(`\bglpat-[0-9A-Za-z_-]{20,}`)},
	{"github-pat", regexp.MustCompile(`\b(?:gh[pousr]_[0-9A-Za-z]{36,}|github_pat_[0-9A-Za-z_]{22,})`)},
	{"slack-token", regexp.MustCompile(`\bxox[baprs]-[0-9A-Za-z-]{10,}`)},
	{"google-api-key", regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	{"jwt", regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`)},
	{"generic-secret-assignment", regexp.MustCompile(`(?i)(?:password|passwd|secret|api[_-]?key|access[_-]?token|auth[_-]?token|client[_-]?secret)["']?\s*[:=]\s*["'][^"'\n]{8,}["']`)},
}

// Finding records one rule match within a file. It deliberately omits the
// matched text so the secret is never echoed back through logs or tool output -
// only its location and the rule that flagged it.
type Finding struct {
	Path string // path relative to the scan root
	Rule string // name of the rule that matched
	Line int    // 1-based line number of the match
}

// String renders a finding as "path:line (rule)".
func (f Finding) String() string {
	return fmt.Sprintf("%s:%d (%s)", f.Path, f.Line, f.Rule)
}

// BlockedError is returned by callers when a commit is refused because changed
// files contain likely secrets. Its message is actionable and never contains a
// secret value, so it is safe to surface directly to the user.
type BlockedError struct {
	Findings []Finding
}

// Error renders an actionable, secret-free summary of every finding.
func (e *BlockedError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "refusing to commit: %d potential secret(s) detected. "+
		"Remove the secret or add the path to .gitignore, then retry:", len(e.Findings))
	for _, f := range e.Findings {
		fmt.Fprintf(&b, "\n  - %s", f.String())
	}
	return b.String()
}

// ScanFiles scans each of paths (relative to root) for secrets and returns all
// findings, ordered by path then line for stable output. Missing files (e.g.
// staged deletions), directories, oversized files, and binary content are
// skipped. An error is returned only for unexpected I/O failures.
func ScanFiles(root string, paths []string) ([]Finding, error) {
	// Confine all file access beneath root with os.Root: any path that tries to
	// escape the scan root - via a "../" segment or a symlink pointing outside -
	// is rejected by the kernel rather than silently followed. Defence in depth
	// for a security-sensitive scanner reading caller-supplied relative paths.
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("opening scan root %s: %w", root, err)
	}
	defer func() { _ = r.Close() }() // read-only root; close error is immaterial

	var findings []Finding
	for _, p := range paths {
		fileFindings, err := scanFile(r, p)
		if err != nil {
			return nil, err
		}
		findings = append(findings, fileFindings...)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Line < findings[j].Line
	})
	return findings, nil
}

// scanFile scans a single file, reporting at most one finding per rule (the
// first match) to keep output concise. Files that do not exist, are directories,
// exceed maxFileSize, or contain a NUL byte (treated as binary) yield no
// findings and no error.
func scanFile(root *os.Root, relPath string) ([]Finding, error) {
	info, err := root.Stat(relPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // deleted in the worktree; nothing to scan
		}
		return nil, fmt.Errorf("stat %s: %w", relPath, err)
	}
	if info.IsDir() || info.Size() > maxFileSize {
		return nil, nil
	}

	f, err := root.Open(relPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", relPath, err)
	}
	defer func() { _ = f.Close() }() // read-only file; close error is immaterial

	data, err := io.ReadAll(io.LimitReader(f, maxFileSize))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}
	if bytes.IndexByte(data, 0) != -1 {
		return nil, nil // binary file
	}

	var findings []Finding
	seen := make(map[string]bool, len(rules))
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), maxLineLen)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Bytes()
		for _, r := range rules {
			if seen[r.name] {
				continue
			}
			if r.pattern.Match(line) {
				findings = append(findings, Finding{Path: relPath, Rule: r.name, Line: lineNo})
				seen[r.name] = true
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", relPath, err)
	}
	return findings, nil
}
