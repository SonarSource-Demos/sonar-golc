package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		expected  string
	}{
		{
			name:      "Text shorter than max length",
			text:      "short",
			maxLength: 10,
			expected:  "short",
		},
		{
			name:      "Text longer than max length",
			text:      "this is a very long text",
			maxLength: 10,
			expected:  "this is...",
		},
		{
			name:      "Text exactly max length",
			text:      "exact",
			maxLength: 5,
			expected:  "exact",
		},
		{
			name:      "Empty text",
			text:      "",
			maxLength: 5,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLength)
			if result != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestIsMainBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		expected   bool
	}{
		{"main branch", "main", true},
		{"master branch", "master", true},
		{"develop branch", "develop", true},
		{"development branch", "development", true},
		{"default branch", "default", true},
		{"feature branch", "feature/test", false},
		{"random branch", "some-branch", false},
		{"empty branch", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMainBranch(tt.branchName)
			if result != tt.expected {
				t.Errorf("isMainBranch(%q) = %v, want %v", tt.branchName, result, tt.expected)
			}
		})
	}
}

func TestGetFirstPartForPlatform(t *testing.T) {
	branch := ProjectBranch{
		Org:        "test-org",
		ProjectKey: "test-project",
		RepoSlug:   "test-repo",
	}

	tests := []struct {
		name     string
		platform string
		expected string
	}{
		{"Azure platform", "azure", "test-project"},
		{"BitBucket DC platform", "bitbucketdc", "test-project"},
		{"BitBucket platform", "bitbucket", "test-project"},
		{"GitLab platform", "gitlab", "test-org"},
		{"GitHub platform", "github", "test-org"},
		{"Unknown platform", "unknown", "test-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstPartForPlatform(tt.platform, branch, "test-repo")
			if result != tt.expected {
				t.Errorf("getFirstPartForPlatform(%q, branch, %q) = %q, want %q", tt.platform, "test-repo", result, tt.expected)
			}
		})
	}
}

func TestCalculateTotals(t *testing.T) {
	repositories := []RepositoryData{
		{
			Lines:      100,
			BlankLines: 10,
			Comments:   20,
			CodeLines:  70,
		},
		{
			Lines:      200,
			BlankLines: 15,
			Comments:   25,
			CodeLines:  160,
		},
	}

	totalLines, totalBlankLines, totalComments, totalCodeLines := calculateTotals(repositories)

	expectedTotalLines := 300
	expectedTotalBlankLines := 25
	expectedTotalComments := 45
	expectedTotalCodeLines := 230

	if totalLines != expectedTotalLines {
		t.Errorf("calculateTotals() totalLines = %d, want %d", totalLines, expectedTotalLines)
	}
	if totalBlankLines != expectedTotalBlankLines {
		t.Errorf("calculateTotals() totalBlankLines = %d, want %d", totalBlankLines, expectedTotalBlankLines)
	}
	if totalComments != expectedTotalComments {
		t.Errorf("calculateTotals() totalComments = %d, want %d", totalComments, expectedTotalComments)
	}
	if totalCodeLines != expectedTotalCodeLines {
		t.Errorf("calculateTotals() totalCodeLines = %d, want %d", totalCodeLines, expectedTotalCodeLines)
	}
}

func TestCreateReportFilePaths(t *testing.T) {
	directory := "/test/directory"

	csvPath, jsonPath, pdfPath := createReportFilePaths(directory)

	expectedCsvPath := "/test/directory/byfile-report/csv-report"
	expectedJsonPath := "/test/directory/byfile-report"
	expectedPdfPath := "/test/directory/byfile-report/pdf-report"

	if csvPath != expectedCsvPath {
		t.Errorf("createReportFilePaths() csvPath = %q, want %q", csvPath, expectedCsvPath)
	}
	if jsonPath != expectedJsonPath {
		t.Errorf("createReportFilePaths() jsonPath = %q, want %q", jsonPath, expectedJsonPath)
	}
	if pdfPath != expectedPdfPath {
		t.Errorf("createReportFilePaths() pdfPath = %q, want %q", pdfPath, expectedPdfPath)
	}
}

func TestDetectPlatformAndReadAnalysis(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_analysis_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Results/config directory
	configDir := "Results/config"
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Test case 1: GitHub analysis file exists
	t.Run("GitHub platform detected", func(t *testing.T) {
		// Create a test GitHub analysis file
		testData := AnalysisResult{
			NumRepositories: 1,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "test-org",
					RepoSlug:   "test-repo",
					MainBranch: "main",
				},
			},
		}

		jsonData, _ := json.Marshal(testData)
		githubFile := filepath.Join(configDir, "analysis_result_github.json")
		err = os.WriteFile(githubFile, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		platform, data, err := detectPlatformAndReadAnalysis()
		if err != nil {
			t.Errorf("detectPlatformAndReadAnalysis() error = %v, want nil", err)
		}
		if platform != "github" {
			t.Errorf("detectPlatformAndReadAnalysis() platform = %q, want %q", platform, "github")
		}
		if len(data) == 0 {
			t.Error("detectPlatformAndReadAnalysis() returned empty data")
		}

		// Clean up
		os.Remove(githubFile)
	})

	// Test case 2: No analysis files exist
	t.Run("No platform files exist", func(t *testing.T) {
		_, _, err := detectPlatformAndReadAnalysis()
		if err == nil {
			t.Error("detectPlatformAndReadAnalysis() error = nil, want error when no files exist")
		}
	})
}

func TestGenerateRepositoryCSVReport(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_csv_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data
	summary := &RepositorySummaryReport{
		TotalRepositories: 2,
		TotalLines:        300,
		TotalBlankLines:   25,
		TotalComments:     45,
		TotalCodeLines:    230,
		Repositories: []RepositoryData{
			{
				Number:     1,
				Repository: "repo1",
				Branch:     "main",
				Lines:      100,
				BlankLines: 10,
				Comments:   20,
				CodeLines:  70,
			},
			{
				Number:     2,
				Repository: "repo2",
				Branch:     "master",
				Lines:      200,
				BlankLines: 15,
				Comments:   25,
				CodeLines:  160,
			},
		},
	}

	// Test CSV generation
	err = generateRepositoryCSVReport(summary, tempDir)
	if err != nil {
		t.Errorf("generateRepositoryCSVReport() error = %v, want nil", err)
	}

	// Verify file was created
	csvFile := filepath.Join(tempDir, "repository_summary.csv")
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		t.Error("generateRepositoryCSVReport() did not create CSV file")
	}

	// Read and verify file contents
	content, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read generated CSV file: %v", err)
	}

	// Check that the file contains expected data
	contentStr := string(content)
	expectedStrings := []string{
		"#,Repository,Branch,Lines,Blank Lines,Comments,Code Lines",
		"1,repo1,main,100,10,20,70",
		"2,repo2,master,200,15,25,160",
		"TOTAL,2 repositories,,300,25,45,230",
	}

	for _, expected := range expectedStrings {
		if !containsLine(contentStr, expected) {
			t.Errorf("generateRepositoryCSVReport() CSV does not contain expected line: %s", expected)
		}
	}
}

func TestGenerateRepositoryJSONReport(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_json_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data
	summary := &RepositorySummaryReport{
		TotalRepositories: 1,
		TotalLines:        100,
		TotalBlankLines:   10,
		TotalComments:     20,
		TotalCodeLines:    70,
		Repositories: []RepositoryData{
			{
				Number:     1,
				Repository: "test-repo",
				Branch:     "main",
				Lines:      100,
				BlankLines: 10,
				Comments:   20,
				CodeLines:  70,
			},
		},
	}

	// Test JSON generation
	err = generateRepositoryJSONReport(summary, tempDir)
	if err != nil {
		t.Errorf("generateRepositoryJSONReport() error = %v, want nil", err)
	}

	// Verify file was created
	jsonFile := filepath.Join(tempDir, "repository_summary.json")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Error("generateRepositoryJSONReport() did not create JSON file")
	}

	// Read and verify file contents
	content, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to read generated JSON file: %v", err)
	}

	// Parse JSON to verify structure
	var parsedSummary RepositorySummaryReport
	err = json.Unmarshal(content, &parsedSummary)
	if err != nil {
		t.Errorf("generateRepositoryJSONReport() generated invalid JSON: %v", err)
	}

	// Verify data integrity
	if parsedSummary.TotalRepositories != summary.TotalRepositories {
		t.Errorf("generateRepositoryJSONReport() TotalRepositories = %d, want %d", parsedSummary.TotalRepositories, summary.TotalRepositories)
	}
	if len(parsedSummary.Repositories) != len(summary.Repositories) {
		t.Errorf("generateRepositoryJSONReport() repository count = %d, want %d", len(parsedSummary.Repositories), len(summary.Repositories))
	}
}

func TestGenerateRepositoryPDFReport(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_pdf_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data
	summary := &RepositorySummaryReport{
		TotalRepositories: 1,
		TotalLines:        100,
		TotalBlankLines:   10,
		TotalComments:     20,
		TotalCodeLines:    70,
		TotalLinesF:       "100",
		TotalBlankLinesF:  "10",
		TotalCommentsF:    "20",
		TotalCodeLinesF:   "70",
		Repositories: []RepositoryData{
			{
				Number:      1,
				Repository:  "test-repo",
				Branch:      "main",
				Lines:       100,
				BlankLines:  10,
				Comments:    20,
				CodeLines:   70,
				LinesF:      "100",
				BlankLinesF: "10",
				CommentsF:   "20",
				CodeLinesF:  "70",
			},
		},
	}

	// Test PDF generation
	err = generateRepositoryPDFReport(summary, tempDir)
	if err != nil {
		t.Errorf("generateRepositoryPDFReport() error = %v, want nil", err)
	}

	// Verify file was created
	pdfFile := filepath.Join(tempDir, "repository_summary.pdf")
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Error("generateRepositoryPDFReport() did not create PDF file")
	}

	// Verify file is not empty
	info, err := os.Stat(pdfFile)
	if err != nil {
		t.Fatalf("Failed to stat generated PDF file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("generateRepositoryPDFReport() created empty PDF file")
	}
}

func TestGenerateRepositorySummaryReports_NoAnalysisFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_no_analysis_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Logs directory to prevent logger errors
	err = os.MkdirAll("Logs", 0755)
	if err != nil {
		t.Fatalf("Failed to create Logs dir: %v", err)
	}

	// Test with no analysis files (should skip gracefully)
	err = GenerateRepositorySummaryReports(tempDir)
	if err != nil {
		t.Errorf("GenerateRepositorySummaryReports() error = %v, want nil (should skip gracefully)", err)
	}
}

func TestGenerateRepositorySummaryReports_WithAnalysisFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_with_analysis_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create necessary directory structure
	dirs := []string{
		"Logs",
		"Results/config",
		"Results/byfile-report",
		"byfile-report/csv-report",
		"byfile-report/pdf-report",
	}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	// Create analysis result file
	analysisData := AnalysisResult{
		NumRepositories: 1,
		ProjectBranches: []ProjectBranch{
			{
				Org:        "test-org",
				RepoSlug:   "test-repo",
				MainBranch: "main",
			},
		},
	}
	analysisJSON, _ := json.Marshal(analysisData)
	err = os.WriteFile("Results/config/analysis_result_github.json", analysisJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to create analysis file: %v", err)
	}

	// Create byfile report
	byfileData := map[string]interface{}{
		"TotalLines":      100,
		"TotalBlankLines": 10,
		"TotalComments":   20,
		"TotalCodeLines":  70,
	}
	byfileJSON, _ := json.Marshal(byfileData)
	err = os.WriteFile("Results/byfile-report/Result_test-org_test-repo_main_byfile.json", byfileJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to create byfile report: %v", err)
	}

	// Test with analysis files and byfile reports
	err = GenerateRepositorySummaryReports(tempDir)
	if err != nil {
		t.Errorf("GenerateRepositorySummaryReports() error = %v, want nil", err)
	}

	// Verify that reports were created
	csvFile := filepath.Join(tempDir, "byfile-report/csv-report/repository_summary.csv")
	jsonFile := filepath.Join(tempDir, "byfile-report/repository_summary.json")
	pdfFile := filepath.Join(tempDir, "byfile-report/pdf-report/repository_summary.pdf")

	files := []string{csvFile, jsonFile, pdfFile}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected report file was not created: %s", file)
		}
	}
}

func TestGenerateReportWithErrorHandling(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_error_handling_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory to ensure logger works
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Logs directory
	err = os.MkdirAll("Logs", 0755)
	if err != nil {
		t.Fatalf("Failed to create Logs dir: %v", err)
	}

	t.Run("Success case", func(t *testing.T) {
		generateReportWithErrorHandling("Test", "/tmp/test.txt", func() error {
			return nil
		})
		// Test passes if no panic occurs
	})

	t.Run("Error case", func(t *testing.T) {
		generateReportWithErrorHandling("Test", "/tmp/test.txt", func() error {
			return os.ErrNotExist
		})
		// Test passes if no panic occurs
	})
}

func TestGetRepositoryData_EmptyAnalysis(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_empty_analysis_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Results/config directory
	err = os.MkdirAll("Results/config", 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create analysis file with empty repositories
	analysisData := AnalysisResult{
		NumRepositories: 0,
		ProjectBranches: []ProjectBranch{},
	}
	analysisJSON, _ := json.Marshal(analysisData)
	err = os.WriteFile("Results/config/analysis_result_github.json", analysisJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to create analysis file: %v", err)
	}

	// Test with empty analysis
	repositories, err := getRepositoryData()
	if err != nil {
		t.Errorf("getRepositoryData() error = %v, want nil", err)
	}
	if len(repositories) != 0 {
		t.Errorf("getRepositoryData() returned %d repositories, want 0", len(repositories))
	}
}

// Helper function to check if a string contains a specific line
func containsLine(content, line string) bool {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

func TestGetRepositoryData_ComplexScenarios(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_complex_scenarios_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create necessary directory structure
	dirs := []string{
		"Results/config",
		"Results/byfile-report",
	}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	t.Run("Invalid JSON handling", func(t *testing.T) {
		// Create invalid JSON file
		err = os.WriteFile("Results/config/analysis_result_github.json", []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid JSON file: %v", err)
		}

		// Test that invalid JSON is handled
		repositories, err := getRepositoryData()
		if err == nil {
			t.Error("getRepositoryData() error = nil, want error for invalid JSON")
		}
		if repositories != nil {
			t.Error("getRepositoryData() returned repositories for invalid JSON")
		}

		// Clean up
		os.Remove("Results/config/analysis_result_github.json")
	})

	t.Run("Invalid byfile JSON handling", func(t *testing.T) {
		// Create valid analysis file
		analysisData := AnalysisResult{
			NumRepositories: 1,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "test-org",
					RepoSlug:   "invalid-byfile-repo",
					MainBranch: "main",
				},
			},
		}
		analysisJSON, _ := json.Marshal(analysisData)
		err = os.WriteFile("Results/config/analysis_result_github.json", analysisJSON, 0644)
		if err != nil {
			t.Fatalf("Failed to create analysis file: %v", err)
		}

		// Create invalid byfile JSON
		err = os.WriteFile("Results/byfile-report/Result_test-org_invalid-byfile-repo_main_byfile.json", []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid byfile JSON: %v", err)
		}

		// Test that invalid byfile JSON is skipped
		repositories, err := getRepositoryData()
		if err != nil {
			t.Errorf("getRepositoryData() error = %v, want nil", err)
		}
		// Should return empty since byfile JSON is invalid
		if len(repositories) != 0 {
			t.Errorf("getRepositoryData() returned %d repositories, want 0 (invalid byfile JSON)", len(repositories))
		}
	})
}

func TestCreatePDFTableHeader(t *testing.T) {
	t.Run("PDF table header creation", func(t *testing.T) {
		// This tests the helper function that was extracted during refactoring
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()

		// Test that the function doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("createPDFTableHeader panicked: %v", r)
			}
		}()

		createPDFTableHeader(pdf, "Test Header")

		// Basic validation that content was added
		if pdf.PageCount() == 0 {
			t.Error("createPDFTableHeader did not add content to PDF")
		}
	})
}

func TestCreateRepositoryPDFRow(t *testing.T) {
	t.Run("PDF row creation", func(t *testing.T) {
		// This tests the helper function that was extracted during refactoring
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()

		repo := RepositoryData{
			Number:      1,
			Repository:  "very-long-repository-name-that-should-be-truncated",
			Branch:      "very-long-branch-name-that-should-be-truncated",
			Lines:       1000,
			BlankLines:  100,
			Comments:    200,
			CodeLines:   700,
			LinesF:      "1.0K",
			BlankLinesF: "100",
			CommentsF:   "200",
			CodeLinesF:  "700",
		}

		// Test that the function doesn't panic and handles long names
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("createRepositoryPDFRow panicked: %v", r)
			}
		}()

		createRepositoryPDFRow(pdf, repo, true)
		createRepositoryPDFRow(pdf, repo, false) // Test alternating colors

		if pdf.PageCount() == 0 {
			t.Error("createRepositoryPDFRow did not add content to PDF")
		}
	})
}

func TestAdvancedPlatformDetection(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_platform_detection_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Results/config directory
	err = os.MkdirAll("Results/config", 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	t.Run("BitBucket DC platform with different filename", func(t *testing.T) {
		// Test the special case for BitBucket DC
		analysisData := AnalysisResult{
			NumRepositories: 1,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "bitbucketdc-org",
					ProjectKey: "BBDC",
					RepoSlug:   "bbdc-repo",
					MainBranch: "main",
				},
			},
		}
		analysisJSON, _ := json.Marshal(analysisData)
		err = os.WriteFile("Results/config/analysis_repos_bitbucketdc.json", analysisJSON, 0644)
		if err != nil {
			t.Fatalf("Failed to create BitBucket DC file: %v", err)
		}

		platform, data, err := detectPlatformAndReadAnalysis()
		if err != nil {
			t.Errorf("detectPlatformAndReadAnalysis() error = %v, want nil", err)
		}
		if platform != "bitbucketdc" {
			t.Errorf("detectPlatformAndReadAnalysis() platform = %q, want %q", platform, "bitbucketdc")
		}
		if len(data) == 0 {
			t.Error("detectPlatformAndReadAnalysis() returned empty data for BitBucket DC")
		}

		// Clean up
		os.Remove("Results/config/analysis_repos_bitbucketdc.json")
	})

	t.Run("All supported platforms", func(t *testing.T) {
		// Test all platforms defined in the detection map
		platforms := map[string]string{
			"github":      "analysis_result_github.json",
			"azure":       "analysis_result_azure.json",
			"bitbucket":   "analysis_result_bitbucket.json",
			"gitlab":      "analysis_result_gitlab.json",
			"bitbucketdc": "analysis_repos_bitbucketdc.json",
		}

		for platform, filename := range platforms {
			t.Run("Platform "+platform, func(t *testing.T) {
				// Clean up previous files
				os.RemoveAll("Results/config")
				os.MkdirAll("Results/config", 0755)

				analysisData := AnalysisResult{
					NumRepositories: 1,
					ProjectBranches: []ProjectBranch{
						{
							Org:        platform + "-org",
							ProjectKey: strings.ToUpper(platform),
							RepoSlug:   platform + "-repo",
							MainBranch: "main",
						},
					},
				}
				analysisJSON, _ := json.Marshal(analysisData)
				err = os.WriteFile("Results/config/"+filename, analysisJSON, 0644)
				if err != nil {
					t.Fatalf("Failed to create %s file: %v", platform, err)
				}

				detectedPlatform, data, err := detectPlatformAndReadAnalysis()
				if err != nil {
					t.Errorf("detectPlatformAndReadAnalysis() for %s error = %v, want nil", platform, err)
				}
				if detectedPlatform != platform {
					t.Errorf("detectPlatformAndReadAnalysis() platform = %q, want %q", detectedPlatform, platform)
				}
				if len(data) == 0 {
					t.Errorf("detectPlatformAndReadAnalysis() returned empty data for %s", platform)
				}
			})
		}
	})
}
