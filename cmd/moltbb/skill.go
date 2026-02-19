package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/utils"
)

const defaultSkillName = "moltbb-agent-diary-publish"

var tagRefPattern = regexp.MustCompile(`^v\d`)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage MoltBB agent skills",
	}

	cmd.AddCommand(newSkillInstallCmd())
	return cmd
}

func newSkillInstallCmd() *cobra.Command {
	var skillDir string
	var repo string
	var ref string
	var force bool

	cmd := &cobra.Command{
		Use:   "install [skill-name]",
		Short: "Install a skill into local Codex skills directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := defaultSkillName
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				skillName = strings.TrimSpace(args[0])
			}

			return runSkillInstall(skillName, skillDir, repo, ref, force)
		},
	}

	cmd.Flags().StringVar(&skillDir, "dir", "~/.codex/skills", "Target directory for installed skills")
	cmd.Flags().StringVar(&repo, "repo", defaultReleaseRepo, "GitHub repo in owner/name format")
	cmd.Flags().StringVar(&ref, "ref", "main", "Git ref for skill source (branch, tag, or latest)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing installed skill")

	return cmd
}

func runSkillInstall(skillName, skillDir, repo, ref string, force bool) error {
	skillName = strings.TrimSpace(skillName)
	if skillName == "" {
		return errors.New("skill name is required")
	}

	repo = strings.TrimSpace(repo)
	if repo == "" {
		repo = defaultReleaseRepo
	}
	if !strings.Contains(repo, "/") {
		return fmt.Errorf("invalid --repo value: %s (expected owner/name)", repo)
	}

	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = "main"
	}
	if strings.EqualFold(ref, "latest") {
		tag, err := fetchLatestTag(repo)
		if err != nil {
			return err
		}
		ref = tag
	}

	skillDirPath, err := utils.ExpandPath(skillDir)
	if err != nil {
		return fmt.Errorf("resolve --dir: %w", err)
	}
	if err := utils.EnsureDir(skillDirPath, 0o755); err != nil {
		return err
	}

	targetDir := filepath.Join(skillDirPath, skillName)
	if utils.FileExists(targetDir) {
		if !force {
			return fmt.Errorf("target already exists: %s (use --force to overwrite)", targetDir)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("remove existing skill dir: %w", err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "moltbb-skill-install-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "repo.tar.gz")
	archiveURL, err := downloadSkillArchive(repo, ref, archivePath)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading skill source: %s\n", archiveURL)
	if err := extractSkillFromArchive(archivePath, skillName, targetDir); err != nil {
		return err
	}

	fmt.Printf("Installed skill: %s\n", targetDir)
	return nil
}

func downloadSkillArchive(repo, ref, outputPath string) (string, error) {
	order := []string{"heads", "tags"}
	if tagRefPattern.MatchString(ref) {
		order = []string{"tags", "heads"}
	}

	var lastErr error
	for _, refType := range order {
		url := fmt.Sprintf("https://codeload.github.com/%s/tar.gz/refs/%s/%s", repo, refType, ref)
		if err := downloadSkillFile(url, outputPath); err == nil {
			return url, nil
		} else {
			lastErr = err
		}
	}

	return "", fmt.Errorf("download skill archive failed for %s@%s: %w", repo, ref, lastErr)
}

func downloadSkillFile(url, outputPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("request failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractSkillFromArchive(archivePath, skillName, targetDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	found := false

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		rel, ok := archiveSkillRelativePath(hdr.Name, skillName)
		if !ok {
			continue
		}
		found = true

		dstPath := filepath.Join(targetDir, filepath.FromSlash(rel))
		relCheck, err := filepath.Rel(targetDir, dstPath)
		if err != nil || strings.HasPrefix(relCheck, "..") {
			return fmt.Errorf("invalid skill archive path: %s", hdr.Name)
		}

		mode := hdr.FileInfo().Mode().Perm()
		if mode == 0 {
			mode = 0o644
		}
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("skill not found in archive: %s", skillName)
	}
	return nil
}

func archiveSkillRelativePath(entryPath, skillName string) (string, bool) {
	cleaned := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(entryPath)), "./")
	parts := strings.Split(cleaned, "/")

	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "skills" && parts[i+1] == skillName {
			if i+2 >= len(parts) {
				return ".", true
			}
			return strings.Join(parts[i+2:], "/"), true
		}
	}
	return "", false
}
