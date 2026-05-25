package vx_puppet

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type LightpandaInstance struct {
	BinaryPath string
	DataDir    string
	Port       int

	cmd *exec.Cmd
}

func DownloadLightpanda(
	ctx context.Context,
	destDir string,
) (string, error) {

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	url, binaryName, err := lightpandaDownloadInfo()
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(destDir, binaryName)

	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed downloading lightpanda: %s",
			resp.Status,
		)
	}

	out, err := os.Create(binaryPath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()

	if err != nil {
		return "", err
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return "", err
	}

	return binaryPath, nil
}

func StartLightpanda(
	ctx context.Context,
	binaryPath string,
	port int,
) (*LightpandaInstance, error) {

	dataDir := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("lightpanda-%d", time.Now().UnixNano()),
	)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(
		ctx,
		binaryPath,
		"serve",
		"--host",
		"127.0.0.1",
		"--port",
		fmt.Sprintf("%d", port),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	instance := &LightpandaInstance{
		BinaryPath: binaryPath,
		DataDir:    dataDir,
		Port:       port,
		cmd:        cmd,
	}

	if err := waitForCDP(
		ctx,
		port,
	); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	return instance, nil
}

func (l *LightpandaInstance) WebSocketURL() string {
	return fmt.Sprintf(
		"ws://127.0.0.1:%d",
		l.Port,
	)
}

func (l *LightpandaInstance) Close() error {

	if l.cmd != nil && l.cmd.Process != nil {
		_ = l.cmd.Process.Kill()
	}

	return os.RemoveAll(l.DataDir)
}

func waitForCDP(
	ctx context.Context,
	port int,
) error {

	url := fmt.Sprintf(
		"http://127.0.0.1:%d/json/version",
		port,
	)

	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := http.Get(url)

		if err == nil {
			resp.Body.Close()
			return nil
		}

		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("lightpanda failed to start")
}

func lightpandaDownloadInfo() (
	url string,
	binaryName string,
	err error,
) {

	switch runtime.GOOS {

	case "linux":

		switch runtime.GOARCH {
		case "amd64":
			return "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-x86_64-linux", "lightpanda", nil

		case "arm64":
			return "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-aarch64-linux", "lightpanda", nil
		}

	case "darwin":

		switch runtime.GOARCH {
		case "amd64":
			return "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-x86_64-macos", "lightpanda", nil

		case "arm64":
			return "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-aarch64-macos", "lightpanda", nil
		}
	}

	return "", "", fmt.Errorf(
		"unsupported platform: %s/%s",
		runtime.GOOS,
		runtime.GOARCH,
	)
}
