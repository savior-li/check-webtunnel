//go:build e2e
// +build e2e

package tests

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "tor-bridge-collector")

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/server")
	cmd.Dir = "../"
	err := cmd.Run()
	require.NoError(t, err)

	return outputPath
}

func TestE2E_InitCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	cmd := exec.Command(binPath, "init")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "config_created")
	assert.Contains(t, string(output), "db_created")

	assert.FileExists(t, filepath.Join(tempDir, "config.yaml"))
}

func TestE2E_InitWithCustomConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()
	customConfig := filepath.Join(tempDir, "myconfig.yaml")

	cmd := exec.Command(binPath, "init", "--config", customConfig)
	cmd.Dir = tempDir
	_, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.FileExists(t, customConfig)
}

func TestE2E_Version(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--version")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "tor-bridge-collector")
}

func TestE2E_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--help")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "tor-bridge-collector")
	assert.Contains(t, string(output), "init")
	assert.Contains(t, string(output), "fetch")
	assert.Contains(t, string(output), "validate")
	assert.Contains(t, string(output), "export")
	assert.Contains(t, string(output), "stats")
}

func TestE2E_StatsCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err)

	cmd := exec.Command(binPath, "stats")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "stats")
}

func TestE2E_ExportCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err)

	outputDir := filepath.Join(tempDir, "output")
	cmd := exec.Command(binPath, "export", "--output", outputDir)
	cmd.Dir = tempDir
	_, err = cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.DirExists(t, outputDir)
	assert.FileExists(t, filepath.Join(outputDir, "bridges.txt"))
}

func TestE2E_ExportJSONFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err)

	outputDir := filepath.Join(tempDir, "output")
	cmd := exec.Command(binPath, "export", "--format", "json", "--output", outputDir)
	cmd.Dir = tempDir
	_, err = cmd.CombinedOutput()
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "bridges.json"))
}

func TestE2E_InvalidCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "nonexistent")
	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(string(output)), "unknown command")
}

func TestE2E_LanguageSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init", "--lang", "en")
	initCmd.Dir = tempDir
	output, err := initCmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "config_created")
}

func TestE2E_MultipleExports(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err)

	outputDir1 := filepath.Join(tempDir, "output1")
	cmd1 := exec.Command(binPath, "export", "--output", outputDir1)
	cmd1.Dir = tempDir
	_, err = cmd1.CombinedOutput()
	require.NoError(t, err)

	outputDir2 := filepath.Join(tempDir, "output2")
	cmd2 := exec.Command(binPath, "export", "--format", "all", "--output", outputDir2)
	cmd2.Dir = tempDir
	_, err = cmd2.CombinedOutput()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir1, "bridges.txt"))
	assert.FileExists(t, filepath.Join(outputDir2, "bridges.txt"))
	assert.FileExists(t, filepath.Join(outputDir2, "bridges.json"))
}

func TestE2E_StatsPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	binPath := buildBinary(t)
	tempDir := t.TempDir()

	initCmd := exec.Command(binPath, "init")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err)

	periods := []string{"day", "week", "month"}
	for _, period := range periods {
		t.Run(period, func(t *testing.T) {
			cmd := exec.Command(binPath, "stats", "--period", period)
			cmd.Dir = tempDir
			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, fmt.Sprintf("period %s failed", period))
			assert.Contains(t, string(output), "stats")
		})
	}
}
