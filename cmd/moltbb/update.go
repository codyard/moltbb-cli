package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultReleaseRepo = "codyard/moltbb-cli"

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func newUpdateCmd() *cobra.Command {
	var targetVersion string
	var repo string
	var force bool

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade"},
		Short:   "Self-update MoltBB CLI from GitHub releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelfUpdate(targetVersion, repo, force)
		},
	}

	cmd.Flags().StringVar(&targetVersion, "version", "latest", "Target version (e.g. v0.4.1 or latest)")
	cmd.Flags().StringVar(&repo, "repo", defaultReleaseRepo, "GitHub repo in owner/name format")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall even if current version matches")

	return cmd
}

func runSelfUpdate(targetVersion, repo string, force bool) error {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		repo = defaultReleaseRepo
	}

	tag, err := resolveTargetTag(targetVersion, repo)
	if err != nil {
		return err
	}

	if !force && tag == version {
		fmt.Printf("Already on %s. Use --force to reinstall.\n", version)
		return nil
	}

	assetName, err := releaseAssetName(tag)
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, assetName)
	fmt.Printf("Downloading %s\n", downloadURL)

	tmpDir, err := os.MkdirTemp("", "moltbb-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return err
	}

	extractedBinary, err := extractBinary(archivePath, tmpDir)
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve current executable: %w", err)
	}
	execPath, _ = filepath.EvalSymlinks(execPath)

	if runtime.GOOS == "windows" {
		pendingPath := execPath + ".new.exe"
		if err := copyFile(extractedBinary, pendingPath, 0o755); err != nil {
			return err
		}
		fmt.Printf("Downloaded update to %s\n", pendingPath)
		fmt.Println("Windows cannot replace a running executable in-place. Close current process and replace manually.")
		return nil
	}

	stagingPath := execPath + ".new"
	if err := copyFile(extractedBinary, stagingPath, 0o755); err != nil {
		return err
	}

	if err := os.Rename(stagingPath, execPath); err != nil {
		return fmt.Errorf("replace binary failed (try with proper permission): %w", err)
	}

	fmt.Printf("Updated successfully to %s\n", tag)
	fmt.Printf("Binary path: %s\n", execPath)
	return nil
}

func resolveTargetTag(targetVersion, repo string) (string, error) {
	targetVersion = strings.TrimSpace(targetVersion)
	if targetVersion == "" || strings.EqualFold(targetVersion, "latest") {
		return fetchLatestTag(repo)
	}
	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = "v" + targetVersion
	}
	return targetVersion, nil
}

func fetchLatestTag(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("request latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("query latest release failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload latestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode latest release response: %w", err)
	}
	if strings.TrimSpace(payload.TagName) == "" {
		return "", errors.New("latest release did not include tag_name")
	}
	return payload.TagName, nil
}

func releaseAssetName(tag string) (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported architecture for self-update: %s", arch)
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return fmt.Sprintf("moltbb_%s_%s_%s.tar.gz", tag, runtime.GOOS, arch), nil
	case "windows":
		return fmt.Sprintf("moltbb_%s_%s_%s.zip", tag, runtime.GOOS, arch), nil
	default:
		return "", fmt.Errorf("unsupported OS for self-update: %s", runtime.GOOS)
	}
}

func downloadFile(url, outputPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("download failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create asset file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("save asset file: %w", err)
	}
	return nil
}

func extractBinary(archivePath, dstDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractFromTarGz(archivePath, dstDir)
	}
	if strings.HasSuffix(archivePath, ".zip") {
		return extractFromZip(archivePath, dstDir)
	}
	return "", fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractFromTarGz(archivePath, dstDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(hdr.Name)
		if name != "moltbb" {
			continue
		}

		outPath := filepath.Join(dstDir, "moltbb")
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return "", err
		}
		_ = outFile.Close()
		return outPath, nil
	}
	return "", errors.New("moltbb binary not found in tar.gz")
}

func extractFromZip(archivePath, dstDir string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if name != "moltbb.exe" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		outPath := filepath.Join(dstDir, "moltbb.exe")
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			rc.Close()
			return "", err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", err
		}
		rc.Close()
		outFile.Close()
		return outPath, nil
	}
	return "", errors.New("moltbb.exe not found in zip")
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("open target file: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy file: %w", err)
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Chmod(dst, mode); err != nil {
		return err
	}
	return nil
}
