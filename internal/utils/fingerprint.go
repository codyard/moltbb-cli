package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func HostInfo() (hostname, osName, arch string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return "", "", "", fmt.Errorf("resolve hostname: %w", err)
	}
	return hostname, runtime.GOOS, runtime.GOARCH, nil
}

func BuildFingerprint(hostname, osName, machineID, cliVersion string) string {
	source := fmt.Sprintf("%s|%s|%s|%s", hostname, osName, machineID, cliVersion)
	hash := sha256.Sum256([]byte(source))
	return hex.EncodeToString(hash[:])
}

func StableFingerprint(cliVersion string) (fingerprint, hostname, osLabel, machineID string, err error) {
	hostname, osName, arch, err := HostInfo()
	if err != nil {
		return "", "", "", "", err
	}

	machineID = resolveMachineID()
	osLabel = fmt.Sprintf("%s/%s", osName, arch)
	fingerprint = BuildFingerprint(hostname, osLabel, machineID, cliVersion)
	return fingerprint, hostname, osLabel, machineID, nil
}

func resolveMachineID() string {
	if value := strings.TrimSpace(os.Getenv("MOLTBB_MACHINE_ID")); value != "" {
		return value
	}

	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if data, err := os.ReadFile(path); err == nil {
			if value := strings.TrimSpace(string(data)); value != "" {
				return value
			}
		}
	}

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		out, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "IOPlatformUUID") {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						value := strings.Trim(parts[1], " \"")
						if value != "" {
							return value
						}
					}
				}
			}
		}
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("wmic", "csproduct", "get", "UUID")
		out, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				value := strings.TrimSpace(line)
				if value != "" && !strings.EqualFold(value, "UUID") {
					return value
				}
			}
		}
	}

	// Safe fallback, keeps fingerprint stable enough for common single-user hosts.
	host, err := os.Hostname()
	if err != nil {
		return ""
	}
	return host
}
