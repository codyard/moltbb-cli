package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
)

func HostInfo() (hostname, osName, arch string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return "", "", "", fmt.Errorf("resolve hostname: %w", err)
	}
	return hostname, runtime.GOOS, runtime.GOARCH, nil
}

func BuildFingerprint(hostname, osName, arch, cliVersion string) string {
	source := fmt.Sprintf("%s|%s|%s|%s", hostname, osName, arch, cliVersion)
	hash := sha256.Sum256([]byte(source))
	return hex.EncodeToString(hash[:])
}
