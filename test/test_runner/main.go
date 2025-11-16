package main

import (
	"fmt"
	"integration-tests/pkg"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
)

func main() {
	fmt.Println("### Initializing Test Env")
	// Start the OTel collector
	// Note: The OTel collector is configured to write telemetry data to "<workdir>/<signal>.jsonl"
	startColCmd := buildCmd(true, "otelcol-contrib", "--config=test/collector_config.yaml")
	if err := startColCmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer startColCmd.Process.Kill()

	// Give the OTel collector time to initialize
	time.Sleep(5 * time.Second)

	testData, err := os.ReadFile("test/test_data.yaml")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var testCases []pkg.TestCase
	if err := yaml.Unmarshal(testData, &testCases); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	var failures []string
	failedCount := 0

	for _, tc := range testCases {
		fmt.Printf("### Testing %q\n", tc.Path)
		testFailed := false

		if err := buildAndRunExampleApp(tc.Path); err != nil {
			failures = append(failures, fmt.Sprintf("%s -> build/run: %v", tc.Path, err))
			testFailed = true
		}

		if !testFailed {
			if err := pkg.TestMetrics(tc.Data.Metrics); err != nil {
				failures = append(failures, fmt.Sprintf("%q -> metrics: %v", tc.Path, err))
				fmt.Printf("--- FAILED: %q metrics\n", tc.Path)
				testFailed = true
			}
			if err := pkg.TestTracing(tc.Data.Traces); err != nil {
				failures = append(failures, fmt.Sprintf("%q -> tracing: %v", tc.Path, err))
				fmt.Printf("--- FAILED: %q tracing\n", tc.Path)
				testFailed = true
			}
			if err := pkg.TestLogs(tc.Data.Logs); err != nil {
				failures = append(failures, fmt.Sprintf("%q -> logs: %v", tc.Path, err))
				fmt.Printf("--- FAILED: %q logs\n", tc.Path)
				testFailed = true
			}
		}

		if testFailed {
			failedCount++
		}

		if err := cleanupTelemetryFiles(); err != nil {
			log.Fatalf("Failed to cleanup telemetry files: %v", err)
		}
	}

	if len(failures) > 0 {
		fmt.Printf("\n##### TEST SUMMARY #####\n")
		fmt.Printf("Failed: %d/%d test cases\n", failedCount, len(testCases))
		for _, failure := range failures {
			fmt.Printf("  • %s\n", failure)
		}
		os.Exit(1)
	}

	fmt.Printf("\n##### ALL TESTS PASSED (%d/%d) #####\n", len(testCases), len(testCases))
}

// Reset the *.jsonl files for the next testcase
func cleanupTelemetryFiles() error {
	files := []string{"metrics.jsonl", "traces.jsonl", "logs.jsonl"}
	for _, f := range files {
		// Deleting the files will cause problems, so we make them empty instead
		if err := os.Truncate(f, 0); err != nil {
			return err
		}
	}

	return nil
}

func buildAndRunExampleApp(path string) error {
	spinBuildCmd := buildCmd(true, "spin", "build", "-f", path)
	if err := spinBuildCmd.Run(); err != nil {
		return err
	}

	spinUpCmd := buildCmd(true, "spin", "up", "-f", path)
	spinUpCmd.Env = append(os.Environ(), "OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318")
	if err := spinUpCmd.Start(); err != nil {
		return err
	}
	defer spinUpCmd.Process.Kill()

	// Give the application time to initialize
	time.Sleep(time.Duration(5 * time.Second))

	resp, err := http.Get("http://localhost:3000")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	expectedResponse := "Hello, world!"
	if string(body) != expectedResponse {
		return fmt.Errorf("The response body for the %q app is not correct.\nWant: %q, Got: %q", path, expectedResponse, string(body))
	}

	// Give the collector time to write files to disk
	time.Sleep(time.Duration(5 * time.Second))
	return nil
}

// Builds an executor that prints to os.Stdout and os.Stdin
func buildCmd(mute bool, args ...string) *exec.Cmd {
	if len(args) == 0 {
		panic("buildCmd requires at least one argument")
	}
	cmd := exec.Command(args[0], args[1:]...)
	// This assumes that the command will be run immediately after being build.
	// If it's not, this will show up in weird places
	fmt.Printf("--- Executing %q\n", strings.Join(args, " "))
	if !mute {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}
