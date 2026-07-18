package agents

import (
	"os"
	"path/filepath"
	"strings"
)

func Markers(ruleName string) (start, end string) {
	u := strings.ToUpper(ruleName)
	return "<!-- " + u + "_START -->", "<!-- " + u + "_END -->"
}

// StripFrontmatter removes YAML frontmatter (--- ... ---) if present.
func StripFrontmatter(body string) string {
	s := strings.TrimPrefix(body, "\uFEFF")
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return strings.TrimSpace(s) + "\n"
	}
	rest := s[4:]
	if i := strings.Index(rest, "\n---\n"); i >= 0 {
		return strings.TrimSpace(rest[i+5:]) + "\n"
	}
	if i := strings.Index(rest, "\r\n---\r\n"); i >= 0 {
		return strings.TrimSpace(rest[i+7:]) + "\n"
	}
	// Malformed / single --- : return as-is trimmed.
	return strings.TrimSpace(s) + "\n"
}

// MarkedBody returns the marker-fenced section from body, or body with markers added.
func MarkedBody(ruleName, body string) string {
	start, end := Markers(ruleName)
	body = StripFrontmatter(body)
	if i := strings.Index(body, start); i >= 0 {
		if j := strings.Index(body[i:], end); j >= 0 {
			return strings.TrimSpace(body[i:i+j+len(end)]) + "\n"
		}
	}
	return start + "\n" + strings.TrimSpace(body) + "\n" + end + "\n"
}

func UpsertMarkedSection(file, block, start, end string) (action string, err error) {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return "", err
	}
	existing, readErr := os.ReadFile(file)
	if readErr != nil && !os.IsNotExist(readErr) {
		return "", readErr
	}
	block = strings.TrimSpace(block) + "\n"
	if os.IsNotExist(readErr) {
		if err := atomicWrite(file, block); err != nil {
			return "", err
		}
		return "created", nil
	}
	content := string(existing)
	si := strings.Index(content, start)
	ei := strings.Index(content, end)
	if si >= 0 && ei > si {
		old := content[si : ei+len(end)]
		if strings.TrimSpace(old) == strings.TrimSpace(block) {
			return "unchanged", nil
		}
		before := content[:si]
		after := content[ei+len(end):]
		next := before + block + after
		if err := atomicWrite(file, next); err != nil {
			return "", err
		}
		return "updated", nil
	}
	trimmed := strings.TrimRight(content, " \t\r\n")
	sep := ""
	if trimmed != "" {
		sep = "\n\n"
	}
	if err := atomicWrite(file, trimmed+sep+block); err != nil {
		return "", err
	}
	return "updated", nil
}

func RemoveMarkedSection(file, start, end string) (action string, err error) {
	b, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return "kept", nil
		}
		return "", err
	}
	content := string(b)
	si := strings.Index(content, start)
	ei := strings.Index(content, end)
	if si < 0 || ei <= si {
		return "not-found", nil
	}
	before := strings.TrimRight(content[:si], " \t\r\n")
	after := strings.TrimLeft(content[ei+len(end):], " \t\r\n")
	joined := before
	if before != "" && after != "" {
		joined += "\n\n"
	}
	joined += after
	if strings.TrimSpace(joined) == "" {
		_ = os.Remove(file)
		return "removed", nil
	}
	if err := atomicWrite(file, strings.TrimSpace(joined)+"\n"); err != nil {
		return "", err
	}
	return "removed", nil
}

func atomicWrite(file, content string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return err
	}
	// Windows rename fails if the destination exists.
	_ = os.Remove(file)
	if err := os.Rename(tmp, file); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// writeMarkdownRule upserts the marked section into an instructions file.
func writeMarkdownRule(file, ruleName, body string) (Result, error) {
	block := MarkedBody(ruleName, body)
	start, end := Markers(ruleName)
	action, err := UpsertMarkedSection(file, block, start, end)
	return Result{Path: file, Action: action}, err
}

func removeMarkdownRule(file, ruleName string) (Result, error) {
	start, end := Markers(ruleName)
	action, err := RemoveMarkedSection(file, start, end)
	return Result{Path: file, Action: action}, err
}
