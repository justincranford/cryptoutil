package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TestProfile struct {
	Name        string                            `json:"name"`
	VUs         int                               `json:"vus"`
	Duration    string                            `json:"duration"`
	RampUp      string                            `json:"rampUp,omitempty"`
	Thresholds  map[string]map[string]interface{} `json:"thresholds"`
	Description string                            `json:"description"`
}

type K6Metrics struct {
	IterationDuration struct {
		Values struct {
			Avg float64 `json:"avg"`
		} `json:"values"`
	} `json:"iteration_duration"`
	HTTPReqDuration struct {
		Values struct {
			Avg  float64 `json:"avg"`
			Rate float64 `json:"rate"`
			P95  float64 `json:"p(95)"`
			P99  float64 `json:"p(99)"`
		} `json:"values"`
	} `json:"http_req_duration"`
	HTTPReqFailed struct {
		Values struct {
			Rate float64 `json:"rate"`
		} `json:"values"`
	} `json:"http_req_failed"`
	HTTPReqs struct {
		Values struct {
			Count float64 `json:"count"`
		} `json:"values"`
	} `json:"http_reqs"`
	KeyGenerationDuration *struct {
		Values struct {
			Avg float64 `json:"avg"`
		} `json:"values"`
	} `json:"key_generation_duration,omitempty"`
	EncryptionDuration *struct {
		Values struct {
			Avg float64 `json:"avg"`
		} `json:"values"`
	} `json:"encryption_duration,omitempty"`
	DecryptionDuration *struct {
		Values struct {
			Avg float64 `json:"avg"`
		} `json:"values"`
	} `json:"decryption_duration,omitempty"`
}

type K6Results struct {
	Metrics K6Metrics `json:"metrics"`
}

type PerformanceResult struct {
	Timestamp         string  `json:"timestamp"`
	FileName          string  `json:"fileName"`
	Duration          float64 `json:"duration"`
	RequestsPerSecond float64 `json:"requestsPerSecond"`
	AvgResponseTime   float64 `json:"avgResponseTime"`
	P95ResponseTime   float64 `json:"p95ResponseTime"`
	P99ResponseTime   float64 `json:"p99ResponseTime"`
	ErrorRate         float64 `json:"errorRate"`
	TotalRequests     float64 `json:"totalRequests"`
	KeyGenDuration    float64 `json:"keyGenDuration"`
	EncryptDuration   float64 `json:"encryptDuration"`
	DecryptDuration   float64 `json:"decryptDuration"`
}

type TrendAnalysis struct {
	ResponseTimeChange float64 `json:"responseTimeChange"`
	ErrorRateChange    float64 `json:"errorRateChange"`
	ThroughputChange   float64 `json:"throughputChange"`
	Status             string  `json:"status"`
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "run":
		runTests(args)
	case "analyze":
		analyzeResults(args)
	case "generate":
		generateScript(args)
	case "quick":
		runQuickTest(args)
	case "help", "-h", "--help":
		showUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("perf-test - Performance testing utilities for cryptoutil")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./scripts/perf-test.go <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  run       Run performance tests using k6")
	fmt.Println("  analyze   Analyze performance test results and generate reports")
	fmt.Println("  generate  Generate k6 test scripts")
	fmt.Println("  quick     Run quick performance test (generate + run)")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Use 'go run ./scripts/perf-test.go <subcommand> -h' for subcommand-specific help.")
}

func runTests(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	profile := fs.String("profile", "quick", "Test profile (quick, full, deep)")
	baseURL := fs.String("base-url", "http://localhost:8080", "Base URL of the cryptoutil API")
	outputDir := fs.String("output", "./performance-results", "Directory to store test results")
	dryRun := fs.Bool("dry-run", false, "Show what would be executed without running tests")
	verbose := fs.Bool("verbose", false, "Enable verbose output")
	help := fs.Bool("help", false, "Show help message")

	fs.Parse(args)

	if *help {
		showRunHelp()
		return
	}

	if err := runTestsInternal(*profile, *baseURL, *outputDir, *dryRun, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showRunHelp() {
	fmt.Println("perf-test run - Run performance tests using k6")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./scripts/perf-test.go run [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -profile string    Test profile (quick, full, deep) (default \"quick\")")
	fmt.Println("  -base-url string   Base URL of the cryptoutil API (default \"http://localhost:8080\")")
	fmt.Println("  -output string     Directory to store test results (default \"./performance-results\")")
	fmt.Println("  -dry-run           Show what would be executed without running tests")
	fmt.Println("  -verbose           Enable verbose output")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go run ./scripts/perf-test.go run -profile quick")
	fmt.Println("  go run ./scripts/perf-test.go run -profile full -base-url https://api.example.com")
	fmt.Println("  go run ./scripts/perf-test.go run -dry-run -verbose")
}

func runTestsInternal(profile, baseURL, outputDir string, dryRun, verbose bool) error {
	// Validate profile
	validProfiles := []string{"quick", "full", "deep"}
	if !contains(validProfiles, profile) {
		return fmt.Errorf("invalid profile '%s'. Valid profiles: %s", profile, strings.Join(validProfiles, ", "))
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get test configuration
	testConfig := getTestProfile(profile)

	// Generate timestamp for results
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	resultFile := filepath.Join(outputDir, fmt.Sprintf("perf-test-%s-%s.json", profile, timestamp))
	scriptFile := filepath.Join(outputDir, fmt.Sprintf("perf-test-%s.js", profile))

	if verbose {
		fmt.Printf("🚀 Starting %s performance test against %s\n", profile, baseURL)
		fmt.Printf("📊 Configuration: %d VUs, %s duration\n", testConfig.VUs, testConfig.Duration)
		fmt.Printf("📁 Results will be saved to: %s\n", resultFile)
	}

	// Generate k6 test script
	if err := generateK6ScriptInternal(scriptFile, testConfig, baseURL); err != nil {
		return fmt.Errorf("failed to generate k6 script: %w", err)
	}

	if dryRun {
		fmt.Println("🔍 Dry run mode - would execute:")
		fmt.Printf("k6 run --out json=%s %s\n", resultFile, scriptFile)
		return nil
	}

	// Check if k6 is available
	if _, err := exec.LookPath("k6"); err != nil {
		return fmt.Errorf("k6 is not installed or not in PATH. Please install k6 first: https://k6.io/docs/get-started/installation/")
	}

	// Run k6 test
	if verbose {
		fmt.Println("📈 Running k6 performance test...")
	}

	cmdArgs := []string{"run", "--out", fmt.Sprintf("json=%s", resultFile), scriptFile}
	k6Cmd := exec.Command("k6", cmdArgs...)

	if verbose {
		k6Cmd.Stdout = os.Stdout
		k6Cmd.Stderr = os.Stderr
	}

	if err := k6Cmd.Run(); err != nil {
		return fmt.Errorf("k6 test failed: %w", err)
	}

	if verbose {
		fmt.Println("✅ Performance test completed successfully")
		fmt.Printf("📊 Results saved to: %s\n", resultFile)
	}

	// Clean up temporary script
	if err := os.Remove(scriptFile); err != nil && verbose {
		fmt.Printf("⚠️  Warning: failed to clean up temporary script: %v\n", err)
	}

	return nil
}

func analyzeResults(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	resultsDir := fs.String("results-dir", "./performance-results", "Directory containing performance test results")
	outputDir := fs.String("output-dir", "./performance-reports", "Directory to store analysis reports")
	help := fs.Bool("help", false, "Show help message")

	fs.Parse(args)

	if *help {
		showAnalyzeHelp()
		return
	}

	if err := analyzeResultsInternal(*resultsDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showAnalyzeHelp() {
	fmt.Println("perf-test analyze - Analyze performance test results and generate reports")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./scripts/perf-test.go analyze [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -results-dir string   Directory containing performance test results (default \"./performance-results\")")
	fmt.Println("  -output-dir string    Directory to store analysis reports (default \"./performance-reports\")")
	fmt.Println("  -help                 Show this help message")
	fmt.Println()
	fmt.Println("OUTPUTS:")
	fmt.Println("  - performance-history.json: Historical performance data")
	fmt.Println("  - performance-dashboard.html: Interactive dashboard with charts")
	fmt.Println("  - performance-summary-YYYY-MM-DD_HH-mm-ss.md: Markdown summary report")
}

func analyzeResultsInternal(resultsDir, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("🔍 Analyzing performance test results...")

	// Find all result files
	resultFiles, err := findResultFiles(resultsDir)
	if err != nil {
		return fmt.Errorf("failed to find result files: %w", err)
	}

	if len(resultFiles) == 0 {
		return fmt.Errorf("no performance test result files found in %s", resultsDir)
	}

	// Sort by modification time (newest first)
	sort.Slice(resultFiles, func(i, j int) bool {
		iInfo, _ := os.Stat(resultFiles[i])
		jInfo, _ := os.Stat(resultFiles[j])
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	// Load historical data
	history, err := loadHistoricalData(outputDir)
	if err != nil {
		log.Printf("Warning: Failed to load historical data: %v", err)
		history = []PerformanceResult{}
	}

	// Process latest results
	latestFile := resultFiles[0]
	fmt.Printf("📈 Processing latest results: %s\n", filepath.Base(latestFile))

	latestResults, err := parseK6Results(latestFile)
	if err != nil {
		return fmt.Errorf("failed to parse latest results file: %w", err)
	}

	// Add to history
	history = append(history, *latestResults)

	// Keep only last 50 entries
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	// Save updated history
	if err := saveHistoricalData(history, outputDir); err != nil {
		log.Printf("Warning: Failed to save historical data: %v", err)
	}

	// Generate trend analysis
	trend := generateTrendAnalysis(history)

	// Generate dashboard
	if err := generatePerformanceDashboard(history, outputDir); err != nil {
		log.Printf("Warning: Failed to generate dashboard: %v", err)
	}

	// Generate summary report
	if err := generateSummaryReport(history, trend, outputDir); err != nil {
		log.Printf("Warning: Failed to generate summary report: %v", err)
	}

	fmt.Printf("📊 Dashboard available at: %s\n", filepath.Join(outputDir, "performance-dashboard.html"))

	// Output key metrics for CI/CD
	fmt.Printf("::set-output name=avg_response_time::%.2f\n", latestResults.AvgResponseTime)
	fmt.Printf("::set-output name=error_rate::%.2f\n", latestResults.ErrorRate*100)
	fmt.Printf("::set-output name=p95_response_time::%.2f\n", latestResults.P95ResponseTime)
	fmt.Printf("::set-output name=throughput::%.2f\n", latestResults.RequestsPerSecond)
	fmt.Printf("::set-output name=performance_status::%s\n", trend.Status)

	return nil
}

func generateScript(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	baseURL := fs.String("base-url", "http://localhost:8080", "Base URL for the application")
	outputFile := fs.String("output", "cryptoutil-performance-test.js", "Output file for the generated k6 script")
	profileName := fs.String("profile", "quick", "Test profile to use (quick, full, deep)")
	vus := fs.Int("vus", 0, "Number of virtual users (overrides profile)")
	duration := fs.String("duration", "", "Test duration (overrides profile)")
	help := fs.Bool("help", false, "Show help message")

	fs.Parse(args)

	if *help {
		showGenerateHelp()
		return
	}

	if err := generateScriptInternal(*baseURL, *outputFile, *profileName, *vus, *duration); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showGenerateHelp() {
	fmt.Println("perf-test generate - Generate k6 test scripts")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./scripts/perf-test.go generate [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -base-url string   Base URL for the application (default \"http://localhost:8080\")")
	fmt.Println("  -output string     Output file for the generated k6 script (default \"cryptoutil-performance-test.js\")")
	fmt.Println("  -profile string    Test profile to use (quick, full, deep) (default \"quick\")")
	fmt.Println("  -vus int           Number of virtual users (overrides profile)")
	fmt.Println("  -duration string   Test duration (overrides profile)")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go run ./scripts/perf-test.go generate -profile quick -output my-test.js")
	fmt.Println("  go run ./scripts/perf-test.go generate -profile full -base-url https://my-app.com")
}

func generateScriptInternal(baseURL, outputFile, profileName string, vus int, duration string) error {
	// Define test profiles
	profiles := map[string]TestProfile{
		"quick": {
			Name:     "quick",
			VUs:      1,
			Duration: "30s",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<500": true},
				"http_req_failed":         {"rate<0.1": true},
				"key_generation_duration": {"p(95)<1000": true},
				"encryption_duration":     {"p(95)<500": true},
				"decryption_duration":     {"p(95)<500": true},
			},
			Description: "Quick performance test with minimal load",
		},
		"full": {
			Name:     "full",
			VUs:      5,
			Duration: "2m",
			RampUp:   "30s",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<750": true},
				"http_req_failed":         {"rate<0.05": true},
				"key_generation_duration": {"p(95)<1500": true},
				"encryption_duration":     {"p(95)<750": true},
				"decryption_duration":     {"p(95)<750": true},
			},
			Description: "Full performance test with moderate load",
		},
		"deep": {
			Name:     "deep",
			VUs:      10,
			Duration: "5m",
			RampUp:   "1m",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<1000": true},
				"http_req_failed":         {"rate<0.02": true},
				"key_generation_duration": {"p(95)<2000": true},
				"encryption_duration":     {"p(95)<1000": true},
				"decryption_duration":     {"p(95)<1000": true},
			},
			Description: "Deep performance test with high load",
		},
	}

	// Get selected profile
	profile, exists := profiles[profileName]
	if !exists {
		return fmt.Errorf("unknown profile '%s'. Available profiles: quick, full, deep", profileName)
	}

	// Override profile settings if specified
	if vus > 0 {
		profile.VUs = vus
	}
	if duration != "" {
		profile.Duration = duration
	}

	// Generate the k6 script
	script := generateK6Script(profile, baseURL)

	// Write to output file
	if err := os.WriteFile(outputFile, []byte(script), 0o644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", outputFile, err)
	}

	fmt.Printf("✅ Generated k6 test script: %s\n", outputFile)
	fmt.Printf("📊 Profile: %s (%s)\n", profile.Name, profile.Description)
	fmt.Printf("👥 Virtual Users: %d\n", profile.VUs)
	fmt.Printf("⏱️  Duration: %s\n", profile.Duration)
	if profile.RampUp != "" {
		fmt.Printf("📈 Ramp Up: %s\n", profile.RampUp)
	}
	fmt.Println("\nTo run the test:")
	fmt.Printf("  k6 run %s\n", outputFile)

	return nil
}

func runQuickTest(args []string) {
	fs := flag.NewFlagSet("quick", flag.ExitOnError)
	baseURL := fs.String("base-url", "http://localhost:8080", "Base URL of the cryptoutil API")
	outputDir := fs.String("output", "./performance-results", "Directory to store test results")
	help := fs.Bool("help", false, "Show help message")

	fs.Parse(args)

	if *help {
		showQuickHelp()
		return
	}

	// Generate quick test script
	tempScript := filepath.Join(*outputDir, "perf-test-quick.js")
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	profile := getTestProfile("quick")
	if err := generateK6ScriptInternal(tempScript, profile, *baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating test script: %v\n", err)
		os.Exit(1)
	}

	// Run the test
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	resultFile := filepath.Join(*outputDir, fmt.Sprintf("perf-test-quick-%s.json", timestamp))

	fmt.Println("🚀 Running quick performance test...")

	// Check if k6 is available
	if _, err := exec.LookPath("k6"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: k6 is not installed or not in PATH. Please install k6 first: https://k6.io/docs/get-started/installation/\n")
		os.Exit(1)
	}

	cmdArgs := []string{"run", "--out", fmt.Sprintf("json=%s", resultFile), tempScript}
	k6Cmd := exec.Command("k6", cmdArgs...)
	k6Cmd.Stdout = os.Stdout
	k6Cmd.Stderr = os.Stderr

	if err := k6Cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: k6 test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Quick performance test completed\n")
	fmt.Printf("📊 Results saved to: %s\n", resultFile)

	// Clean up temporary script
	os.Remove(tempScript)
}

func showQuickHelp() {
	fmt.Println("perf-test quick - Run quick performance test (generate + run)")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run ./scripts/perf-test.go quick [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -base-url string   Base URL of the cryptoutil API (default \"http://localhost:8080\")")
	fmt.Println("  -output string     Directory to store test results (default \"./performance-results\")")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("This command generates a quick performance test script and runs it immediately.")
}

// Helper functions (shared between subcommands)

func getTestProfile(profileName string) TestProfile {
	profiles := map[string]TestProfile{
		"quick": {
			Name:     "quick",
			VUs:      1,
			Duration: "30s",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<500": true},
				"http_req_failed":         {"rate<0.1": true},
				"key_generation_duration": {"p(95)<1000": true},
				"encryption_duration":     {"p(95)<500": true},
				"decryption_duration":     {"p(95)<500": true},
			},
			Description: "Quick performance test with minimal load",
		},
		"full": {
			Name:     "full",
			VUs:      5,
			Duration: "2m",
			RampUp:   "30s",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<750": true},
				"http_req_failed":         {"rate<0.05": true},
				"key_generation_duration": {"p(95)<1500": true},
				"encryption_duration":     {"p(95)<750": true},
				"decryption_duration":     {"p(95)<750": true},
			},
			Description: "Full performance test with moderate load",
		},
		"deep": {
			Name:     "deep",
			VUs:      10,
			Duration: "5m",
			RampUp:   "1m",
			Thresholds: map[string]map[string]interface{}{
				"http_req_duration":       {"p(95)<1000": true},
				"http_req_failed":         {"rate<0.02": true},
				"key_generation_duration": {"p(95)<2000": true},
				"encryption_duration":     {"p(95)<1000": true},
				"decryption_duration":     {"p(95)<1000": true},
			},
			Description: "Deep performance test with high load",
		},
	}

	return profiles[profileName]
}

func generateK6Script(profile TestProfile, baseURL string) string {
	// Convert thresholds to JavaScript object format
	thresholds := make([]string, 0, len(profile.Thresholds))
	for metric, conditions := range profile.Thresholds {
		conditionsStr := make([]string, 0, len(conditions))
		for condition := range conditions {
			conditionsStr = append(conditionsStr, fmt.Sprintf("'%s'", condition))
		}
		thresholds = append(thresholds, fmt.Sprintf("    %s: [%s]", metric, strings.Join(conditionsStr, ", ")))
	}

	// Generate options section
	options := fmt.Sprintf(`export let options = {
  vus: %d,
  duration: '%s',`, profile.VUs, profile.Duration)

	if profile.RampUp != "" {
		options += fmt.Sprintf(`
  stages: [
    { duration: '%s', target: %d },
    { duration: '%s', target: %d },
  ],`, profile.RampUp, profile.VUs, profile.Duration, profile.VUs)
	}

	options += fmt.Sprintf(`
  thresholds: {
%s
  },
};`, strings.Join(thresholds, ",\n"))

	// Generate the complete script
	script := fmt.Sprintf(`import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics for cryptoutil operations
const keyGenDuration = new Trend('key_generation_duration');
const encryptDuration = new Trend('encryption_duration');
const decryptDuration = new Trend('decryption_duration');

%s

// Base URL for the application
const BASE_URL = '%s';

// Test scenarios for cryptoutil
export default function () {
  // Health check
  const healthResponse = http.get(`+"`${BASE_URL}/health`"+`);
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  });

  // Key generation test (RSA)
  const keyGenStart = new Date().getTime();
  const keyGenResponse = http.post(`+"`${BASE_URL}/api/v1/keys/generate`"+`, JSON.stringify({
    algorithm: 'RSA',
    keySize: 2048,
  }), {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  const keyGenEnd = new Date().getTime();
  keyGenDuration.add(keyGenEnd - keyGenStart);

  check(keyGenResponse, {
    'key generation status is 201': (r) => r.status === 201,
    'key generation response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  // If key generation succeeded, test encryption/decryption
  if (keyGenResponse.status === 201) {
    const keyData = JSON.parse(keyGenResponse.body);
    const keyId = keyData.id;

    // Test data for encryption
    const testData = 'Hello, World! This is test data for encryption.';
    const encryptPayload = {
      keyId: keyId,
      plaintext: btoa(testData), // Base64 encode for JSON transport
      algorithm: 'RSA-OAEP',
    };

    // Encryption test
    const encryptStart = new Date().getTime();
    const encryptResponse = http.post(`+"`${BASE_URL}/api/v1/crypto/encrypt`"+`, JSON.stringify(encryptPayload), {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    const encryptEnd = new Date().getTime();
    encryptDuration.add(encryptEnd - encryptStart);

    check(encryptResponse, {
      'encryption status is 200': (r) => r.status === 200,
      'encryption response time < 1000ms': (r) => r.timings.duration < 1000,
    });

    // If encryption succeeded, test decryption
    if (encryptResponse.status === 200) {
      const encryptData = JSON.parse(encryptResponse.body);
      const decryptPayload = {
        keyId: keyId,
        ciphertext: encryptData.ciphertext,
        algorithm: 'RSA-OAEP',
      };

      // Decryption test
      const decryptStart = new Date().getTime();
      const decryptResponse = http.post(`+"`${BASE_URL}/api/v1/crypto/decrypt`"+`, JSON.stringify(decryptPayload), {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      const decryptEnd = new Date().getTime();
      decryptDuration.add(decryptEnd - decryptStart);

      check(decryptResponse, {
        'decryption status is 200': (r) => r.status === 200,
        'decryption response time < 1000ms': (r) => r.timings.duration < 1000,
        'decrypted data matches original': (r) => {
          if (r.status === 200) {
            const decryptData = JSON.parse(r.body);
            return atob(decryptData.plaintext) === testData;
          }
          return false;
        },
      });
    }
  }

  // Random sleep to simulate user think time
  sleep(Math.random() * 2 + 1); // 1-3 seconds
}

// Setup function - runs before the test starts
export function setup() {
  console.log(`+"`Starting performance test against ${BASE_URL}`"+`);

  // Warm-up request to ensure application is ready
  const warmupResponse = http.get(`+"`${BASE_URL}/health`"+`);
  if (warmupResponse.status !== 200) {
    console.error(`+"`Warm-up failed: ${warmupResponse.status} - ${warmupResponse.body}`"+`);
    // Don't fail the test, but log the issue
  }

  return {};
}

// Teardown function - runs after the test completes
export function teardown(data) {
  console.log('Performance test completed');
}
`, options, baseURL)

	return script
}

func generateK6ScriptInternal(scriptPath string, config TestProfile, baseURL string) error {
	k6Script := generateK6Script(config, baseURL)
	return os.WriteFile(scriptPath, []byte(k6Script), 0o644)
}

func findResultFiles(resultsDir string) ([]string, error) {
	pattern := filepath.Join(resultsDir, "perf-test-*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func parseK6Results(filePath string) (*PerformanceResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var k6Data K6Results
	if err := json.Unmarshal(data, &k6Data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &PerformanceResult{
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
		FileName:          filepath.Base(filePath),
		Duration:          k6Data.Metrics.IterationDuration.Values.Avg,
		RequestsPerSecond: k6Data.Metrics.HTTPReqDuration.Values.Rate,
		AvgResponseTime:   k6Data.Metrics.HTTPReqDuration.Values.Avg,
		P95ResponseTime:   k6Data.Metrics.HTTPReqDuration.Values.P95,
		P99ResponseTime:   k6Data.Metrics.HTTPReqDuration.Values.P99,
		ErrorRate:         k6Data.Metrics.HTTPReqFailed.Values.Rate,
		TotalRequests:     k6Data.Metrics.HTTPReqs.Values.Count,
		KeyGenDuration:    0,
		EncryptDuration:   0,
		DecryptDuration:   0,
	}

	if k6Data.Metrics.KeyGenerationDuration != nil {
		result.KeyGenDuration = k6Data.Metrics.KeyGenerationDuration.Values.Avg
	}
	if k6Data.Metrics.EncryptionDuration != nil {
		result.EncryptDuration = k6Data.Metrics.EncryptionDuration.Values.Avg
	}
	if k6Data.Metrics.DecryptionDuration != nil {
		result.DecryptDuration = k6Data.Metrics.DecryptionDuration.Values.Avg
	}

	return result, nil
}

func loadHistoricalData(outputDir string) ([]PerformanceResult, error) {
	historyFile := filepath.Join(outputDir, "performance-history.json")
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []PerformanceResult{}, nil
		}
		return nil, err
	}

	var history []PerformanceResult
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func saveHistoricalData(history []PerformanceResult, outputDir string) error {
	historyFile := filepath.Join(outputDir, "performance-history.json")
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0o644)
}

func generateTrendAnalysis(history []PerformanceResult) TrendAnalysis {
	if len(history) < 2 {
		return TrendAnalysis{
			Status: "insufficient_data",
		}
	}

	latest := history[len(history)-1]
	previous := history[len(history)-2]

	// Calculate trends (positive = improvement, negative = degradation)
	responseTimeTrend := ((previous.AvgResponseTime - latest.AvgResponseTime) / previous.AvgResponseTime) * 100
	errorRateTrend := ((previous.ErrorRate - latest.ErrorRate) / previous.ErrorRate) * 100
	throughputTrend := ((latest.RequestsPerSecond - previous.RequestsPerSecond) / previous.RequestsPerSecond) * 100

	status := "stable"
	if math.Abs(responseTimeTrend) > 10 || math.Abs(errorRateTrend) > 5 {
		status = "significant_change"
	}

	return TrendAnalysis{
		ResponseTimeChange: math.Round(responseTimeTrend*100) / 100,
		ErrorRateChange:    math.Round(errorRateTrend*100) / 100,
		ThroughputChange:   math.Round(throughputTrend*100) / 100,
		Status:             status,
	}
}

func generatePerformanceDashboard(history []PerformanceResult, outputDir string) error {
	if len(history) == 0 {
		return fmt.Errorf("no historical data available for dashboard generation")
	}

	latest := history[len(history)-1]
	trend := generateTrendAnalysis(history)

	// Convert history to JSON for JavaScript
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return err
	}

	dashboardTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cryptoutil Performance Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric-card { background: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .metric-value { font-size: 2em; font-weight: bold; color: #007bff; }
        .metric-label { color: #666; margin-top: 5px; }
        .trend { font-size: 0.9em; margin-top: 10px; }
        .trend.positive { color: #28a745; }
        .trend.negative { color: #dc3545; }
        .trend.neutral { color: #6c757d; }
        .chart-container { margin: 30px 0; height: 400px; }
        .status-indicator {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.9em;
            font-weight: bold;
        }
        .status-good { background: #d4edda; color: #155724; }
        .status-warning { background: #fff3cd; color: #856404; }
        .status-error { background: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Cryptoutil Performance Dashboard</h1>
            <p>Last updated: {{.LastUpdated}}</p>
            <div class="status-indicator {{.StatusClass}}">
                {{.StatusText}}
            </div>
        </div>

        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-value">{{.AvgResponseTime}}ms</div>
                <div class="metric-label">Avg Response Time</div>
                <div class="trend {{.ResponseTimeTrendClass}}">
                    {{.ResponseTimeTrend}}
                </div>
            </div>

            <div class="metric-card">
                <div class="metric-value">{{.RequestsPerSecond}}/s</div>
                <div class="metric-label">Requests/Second</div>
                <div class="trend {{.ThroughputTrendClass}}">
                    {{.ThroughputTrend}}
                </div>
            </div>

            <div class="metric-card">
                <div class="metric-value">{{.ErrorRate}}%</div>
                <div class="metric-label">Error Rate</div>
                <div class="trend {{.ErrorRateTrendClass}}">
                    {{.ErrorRateTrend}}
                </div>
            </div>

            <div class="metric-card">
                <div class="metric-value">{{.TotalRequests}}</div>
                <div class="metric-label">Total Requests</div>
                <div class="metric-label">Test Duration: {{.TestDuration}}s</div>
            </div>
        </div>

        <div class="chart-container">
            <canvas id="performanceChart"></canvas>
        </div>
    </div>

    <script>
        const ctx = document.getElementById('performanceChart').getContext('2d');
        const historyData = {{.HistoryJSON}};

        const labels = historyData.map(d => new Date(d.timestamp).toLocaleDateString());
        const avgResponseTimes = historyData.map(d => d.avgResponseTime);
        const errorRates = historyData.map(d => d.errorRate * 100);
        const throughputs = historyData.map(d => d.requestsPerSecond);

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Avg Response Time (ms)',
                    data: avgResponseTimes,
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    yAxisID: 'y',
                }, {
                    label: 'Error Rate (%)',
                    data: errorRates,
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    yAxisID: 'y1',
                }, {
                    label: 'Requests/sec',
                    data: throughputs,
                    borderColor: 'rgb(54, 162, 235)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    yAxisID: 'y1',
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false,
                },
                stacked: false,
                scales: {
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                            display: true,
                            text: 'Response Time (ms)'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'Rate (%) / Throughput (req/s)'
                        },
                        grid: {
                            drawOnChartArea: false,
                        },
                    }
                }
            }
        });
    </script>
</body>
</html>`

	// Prepare template data
	statusClass := "status-good"
	statusText := "✅ Healthy"
	if latest.ErrorRate >= 0.1 {
		statusClass = "status-error"
		statusText = "❌ Critical"
	} else if latest.ErrorRate >= 0.05 {
		statusClass = "status-warning"
		statusText = "⚠️ Warning"
	}

	responseTimeTrendClass := "neutral"
	responseTimeTrend := "→ No change"
	if trend.ResponseTimeChange > 0 {
		responseTimeTrendClass = "positive"
		responseTimeTrend = fmt.Sprintf("↗ +%.1f%%", trend.ResponseTimeChange)
	} else if trend.ResponseTimeChange < 0 {
		responseTimeTrendClass = "negative"
		responseTimeTrend = fmt.Sprintf("↘ %.1f%%", trend.ResponseTimeChange)
	}

	throughputTrendClass := "neutral"
	throughputTrend := "→ No change"
	if trend.ThroughputChange > 0 {
		throughputTrendClass = "positive"
		throughputTrend = fmt.Sprintf("↗ +%.1f%%", trend.ThroughputChange)
	} else if trend.ThroughputChange < 0 {
		throughputTrendClass = "negative"
		throughputTrend = fmt.Sprintf("↘ %.1f%%", trend.ThroughputChange)
	}

	errorRateTrendClass := "neutral"
	errorRateTrend := "→ No change"
	if trend.ErrorRateChange > 0 {
		errorRateTrendClass = "positive"
		errorRateTrend = fmt.Sprintf("↗ +%.1f%%", trend.ErrorRateChange)
	} else if trend.ErrorRateChange < 0 {
		errorRateTrendClass = "negative"
		errorRateTrend = fmt.Sprintf("↘ %.1f%%", trend.ErrorRateChange)
	}

	templateData := struct {
		LastUpdated            string
		StatusClass            string
		StatusText             string
		AvgResponseTime        string
		ResponseTimeTrendClass string
		ResponseTimeTrend      string
		RequestsPerSecond      string
		ThroughputTrendClass   string
		ThroughputTrend        string
		ErrorRate              string
		ErrorRateTrendClass    string
		ErrorRateTrend         string
		TotalRequests          float64
		TestDuration           string
		HistoryJSON            template.JS
	}{
		LastUpdated:            time.Now().Format("2006-01-02 15:04:05"),
		StatusClass:            statusClass,
		StatusText:             statusText,
		AvgResponseTime:        fmt.Sprintf("%.0f", math.Round(latest.AvgResponseTime)),
		ResponseTimeTrendClass: responseTimeTrendClass,
		ResponseTimeTrend:      responseTimeTrend,
		RequestsPerSecond:      fmt.Sprintf("%.1f", math.Round(latest.RequestsPerSecond*10)/10),
		ThroughputTrendClass:   throughputTrendClass,
		ThroughputTrend:        throughputTrend,
		ErrorRate:              fmt.Sprintf("%.2f", math.Round(latest.ErrorRate*10000)/100),
		ErrorRateTrendClass:    errorRateTrendClass,
		ErrorRateTrend:         errorRateTrend,
		TotalRequests:          latest.TotalRequests,
		TestDuration:           fmt.Sprintf("%.1f", math.Round(latest.Duration/1000*10)/10),
		HistoryJSON:            template.JS(historyJSON),
	}

	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return err
	}

	outputFile := filepath.Join(outputDir, "performance-dashboard.html")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, templateData); err != nil {
		return err
	}

	fmt.Printf("📊 Performance dashboard generated: %s\n", outputFile)
	return nil
}

func generateSummaryReport(history []PerformanceResult, trend TrendAnalysis, outputDir string) error {
	if len(history) == 0 {
		return fmt.Errorf("no historical data available for summary report")
	}

	latest := history[len(history)-1]
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	summaryFile := filepath.Join(outputDir, fmt.Sprintf("performance-summary-%s.md", timestamp))

	summaryTemplate := `# Performance Test Summary
**Date:** {{.Date}}
**Test File:** {{.TestFile}}

## Latest Results
- **Average Response Time:** {{.AvgResponseTime}}ms
- **95th Percentile:** {{.P95ResponseTime}}ms
- **99th Percentile:** {{.P99ResponseTime}}ms
- **Requests/Second:** {{.Throughput}}
- **Error Rate:** {{.ErrorRate}}%
- **Total Requests:** {{.TotalRequests}}

## Performance Trends
- **Response Time:** {{.ResponseTimeChange}}% change
- **Error Rate:** {{.ErrorRateChange}}% change
- **Throughput:** {{.ThroughputChange}}% change
- **Status:** {{.TrendStatus}}

## Threshold Compliance
- {{.ResponseTimeCompliance}} Response Time (P95 < 500ms): {{.ResponseTimeStatus}}
- {{.ErrorRateCompliance}} Error Rate (< 10%): {{.ErrorRateStatus}}

## Historical Data Points: {{.HistoryCount}}
`

	responseTimeCompliance := "✅"
	responseTimeStatus := "PASS"
	if latest.P95ResponseTime >= 500 {
		responseTimeCompliance = "❌"
		responseTimeStatus = "FAIL"
	}

	errorRateCompliance := "✅"
	errorRateStatus := "PASS"
	if latest.ErrorRate >= 0.1 {
		errorRateCompliance = "❌"
		errorRateStatus = "FAIL"
	}

	templateData := struct {
		Date                   string
		TestFile               string
		AvgResponseTime        string
		P95ResponseTime        string
		P99ResponseTime        string
		Throughput             string
		ErrorRate              string
		TotalRequests          float64
		ResponseTimeChange     float64
		ErrorRateChange        float64
		ThroughputChange       float64
		TrendStatus            string
		ResponseTimeCompliance string
		ResponseTimeStatus     string
		ErrorRateCompliance    string
		ErrorRateStatus        string
		HistoryCount           int
	}{
		Date:                   time.Now().Format("2006-01-02 15:04:05"),
		TestFile:               latest.FileName,
		AvgResponseTime:        fmt.Sprintf("%.2f", math.Round(latest.AvgResponseTime*100)/100),
		P95ResponseTime:        fmt.Sprintf("%.2f", math.Round(latest.P95ResponseTime*100)/100),
		P99ResponseTime:        fmt.Sprintf("%.2f", math.Round(latest.P99ResponseTime*100)/100),
		Throughput:             fmt.Sprintf("%.2f", math.Round(latest.RequestsPerSecond*100)/100),
		ErrorRate:              fmt.Sprintf("%.2f", math.Round(latest.ErrorRate*10000)/100),
		TotalRequests:          latest.TotalRequests,
		ResponseTimeChange:     trend.ResponseTimeChange,
		ErrorRateChange:        trend.ErrorRateChange,
		ThroughputChange:       trend.ThroughputChange,
		TrendStatus:            trend.Status,
		ResponseTimeCompliance: responseTimeCompliance,
		ResponseTimeStatus:     responseTimeStatus,
		ErrorRateCompliance:    errorRateCompliance,
		ErrorRateStatus:        errorRateStatus,
		HistoryCount:           len(history),
	}

	tmpl, err := template.New("summary").Parse(summaryTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(summaryFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, templateData); err != nil {
		return err
	}

	fmt.Printf("📋 Summary report generated: %s\n", summaryFile)
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
