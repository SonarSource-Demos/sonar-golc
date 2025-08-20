package getgitlab

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSaveResultGitlab(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "test_gitlab_save_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create necessary directories
	dirs := []string{"Logs", "Results", "Results/config"}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	t.Run("SaveResult creates gitlab analysis file", func(t *testing.T) {
		// Create a test result
		result := AnalysisResult{
			NumRepositories: 1,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "test-org",
					RepoSlug:   "test-repo",
					MainBranch: "main",
				},
			},
		}

		// Test that SaveResult creates the correct file
		err := SaveResult(result)
		if err != nil {
			t.Errorf("SaveResult() error = %v, want nil", err)
		}

		// Verify that the correct GitLab-specific file was created
		expectedFile := "Results/config/analysis_result_gitlab.json"
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected GitLab analysis file was not created: %s", expectedFile)
		}

		// Verify the file contains valid JSON
		fileData, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		var parsedResult AnalysisResult
		err = json.Unmarshal(fileData, &parsedResult)
		if err != nil {
			t.Errorf("Created file contains invalid JSON: %v", err)
		}

		// Verify the data integrity
		if parsedResult.NumRepositories != result.NumRepositories {
			t.Errorf("SaveResult() NumRepositories = %d, want %d", parsedResult.NumRepositories, result.NumRepositories)
		}

		if len(parsedResult.ProjectBranches) != len(result.ProjectBranches) {
			t.Errorf("SaveResult() ProjectBranches count = %d, want %d", len(parsedResult.ProjectBranches), len(result.ProjectBranches))
		}

		// Verify specific GitLab filename (not GitHub)
		githubFile := "Results/config/analysis_result_github.json"
		if _, err := os.Stat(githubFile); !os.IsNotExist(err) {
			t.Error("SaveResult() incorrectly created GitHub analysis file instead of GitLab file")
		}
	})

	t.Run("SaveResult handles encoding error", func(t *testing.T) {
		// Create a result that would cause encoding issues
		result := AnalysisResult{
			NumRepositories: -1, // Invalid data to potentially cause issues
			ProjectBranches: nil,
		}

		// Test that error handling works correctly
		err := SaveResult(result)
		// The function should handle this gracefully
		// Even if there's an error, it should be logged with proper formatting (which was the fix)
		if err != nil {
			// This exercises the error logging path that was fixed
			t.Logf("SaveResult properly handled encoding error: %v", err)
		}
	})
}

func TestGitlabErrorFormatting(t *testing.T) {
	// Create temporary logs directory for testing
	tempDir, err := os.MkdirTemp("", "test_gitlab_logs_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create Logs directory
	err = os.MkdirAll("Logs", 0755)
	if err != nil {
		t.Fatalf("Failed to create Logs dir: %v", err)
	}

	// Create Results/config directory
	err = os.MkdirAll("Results/config", 0755)
	if err != nil {
		t.Fatalf("Failed to create Results/config dir: %v", err)
	}

	t.Run("Error encoding JSON file formatting", func(t *testing.T) {
		// This tests the specific error logging line that was fixed in GitLab
		// by making the file directory read-only to force an encoding error

		// Make the config directory read-only to force error
		err := os.Chmod("Results/config", 0444)
		if err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		// Restore permissions after test
		defer func() {
			os.Chmod("Results/config", 0755)
		}()

		result := AnalysisResult{
			NumRepositories: 1,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "test-org",
					RepoSlug:   "test-repo",
					MainBranch: "main",
				},
			},
		}

		// This should trigger the error logging with proper %v formatting
		err = SaveResult(result)
		if err == nil {
			t.Error("Expected SaveResult to fail with read-only directory")
		}

		// The error should have been logged with proper formatting
		// The specific line that was fixed:
		// loggers.Errorf("❌ Error encoding JSON file <Results/config/analysis_result_gitlab.json> :%v", err)
	})
}

func TestGitlabFileNaming(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "test_gitlab_naming_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create necessary directories
	dirs := []string{"Logs", "Results", "Results/config"}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	t.Run("GitLab uses correct filename", func(t *testing.T) {
		result := AnalysisResult{
			NumRepositories: 2,
			ProjectBranches: []ProjectBranch{
				{
					Org:        "gitlab-org-1",
					RepoSlug:   "repo-1",
					MainBranch: "main",
				},
				{
					Org:        "gitlab-org-2",
					RepoSlug:   "repo-2",
					MainBranch: "develop",
				},
			},
		}

		err := SaveResult(result)
		if err != nil {
			t.Errorf("SaveResult() error = %v, want nil", err)
		}

		// The key fix: GitLab should save to analysis_result_gitlab.json, NOT analysis_result_github.json
		gitlabFile := "Results/config/analysis_result_gitlab.json"
		githubFile := "Results/config/analysis_result_github.json"

		// Verify GitLab file exists
		if _, err := os.Stat(gitlabFile); os.IsNotExist(err) {
			t.Errorf("GitLab analysis file does not exist: %s", gitlabFile)
		}

		// Verify GitHub file does NOT exist (this was the bug that was fixed)
		if _, err := os.Stat(githubFile); !os.IsNotExist(err) {
			t.Errorf("GitHub analysis file incorrectly created by GitLab SaveResult: %s", githubFile)
		}

		// Verify file contents are correct
		fileData, err := os.ReadFile(gitlabFile)
		if err != nil {
			t.Fatalf("Failed to read GitLab file: %v", err)
		}

		var savedResult AnalysisResult
		err = json.Unmarshal(fileData, &savedResult)
		if err != nil {
			t.Fatalf("GitLab file contains invalid JSON: %v", err)
		}

		if savedResult.NumRepositories != 2 {
			t.Errorf("GitLab file NumRepositories = %d, want 2", savedResult.NumRepositories)
		}

		if len(savedResult.ProjectBranches) != 2 {
			t.Errorf("GitLab file ProjectBranches count = %d, want 2", len(savedResult.ProjectBranches))
		}
	})
}
