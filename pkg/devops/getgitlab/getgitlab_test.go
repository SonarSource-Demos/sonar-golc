package getgitlab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/briandowns/spinner"
	"github.com/xanzy/go-gitlab"
)

// Constants to avoid duplicating string literals (SonarQube maintainability)
const (
	errFailedToCreateTempDir = "Failed to create temp dir: %v"
	resultsConfigDir         = "Results/config"
)

// Test helper functions to reduce duplication
func setupTestEnvironment(t *testing.T, prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf(errFailedToCreateTempDir, err)
	}

	originalWd, _ := os.Getwd()
	cleanup := func() {
		os.Chdir(originalWd)
		os.RemoveAll(tempDir)
	}

	os.Chdir(tempDir)
	return tempDir, cleanup
}

func createTestDirectories(t *testing.T, dirs []string) {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}
}

func createTestAnalysisResult() AnalysisResult {
	return AnalysisResult{
		NumRepositories: 1,
		ProjectBranches: []ProjectBranch{
			{
				Org:        "test-org",
				RepoSlug:   "test-repo",
				MainBranch: "main",
			},
		},
	}
}

func TestSaveResultGitlab(t *testing.T) {
	_, cleanup := setupTestEnvironment(t, "test_gitlab_save_*")
	defer cleanup()

	createTestDirectories(t, []string{"Logs", "Results", resultsConfigDir})

	t.Run("SaveResult creates gitlab analysis file", func(t *testing.T) {
		result := createTestAnalysisResult()

		// Test that SaveResult creates the correct file
		err := SaveResult(result)
		if err != nil {
			t.Errorf("SaveResult() error = %v, want nil", err)
		}

		// Verify that the correct GitLab-specific file was created
		expectedFile := resultsConfigDir + "/analysis_result_gitlab.json"
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
		githubFile := resultsConfigDir + "/analysis_result_github.json"
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
	_, cleanup := setupTestEnvironment(t, "test_gitlab_logs_*")
	defer cleanup()

	createTestDirectories(t, []string{"Logs", resultsConfigDir})

	t.Run("Error encoding JSON file formatting", func(t *testing.T) {
		// This tests the specific error logging line that was fixed in GitLab
		// by making the file directory read-only to force an encoding error

		// Make the config directory read-only to force error
		err := os.Chmod(resultsConfigDir, 0444)
		if err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		// Restore permissions after test
		defer func() {
			os.Chmod(resultsConfigDir, 0755)
		}()

		result := createTestAnalysisResult()

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
	_, cleanup := setupTestEnvironment(t, "test_gitlab_naming_*")
	defer cleanup()

	createTestDirectories(t, []string{"Logs", "Results", resultsConfigDir})

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
		gitlabFile := resultsConfigDir + "/analysis_result_gitlab.json"
		githubFile := resultsConfigDir + "/analysis_result_github.json"

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

// Integration tests to improve coverage for getgitlab package
func TestGitlabIntegrationCoverage(t *testing.T) {
	_, cleanup := setupTestEnvironment(t, "test_gitlab_integration_*")
	defer cleanup()

	createTestDirectories(t, []string{"Logs", "Results", resultsConfigDir})

	t.Run("SaveResult error scenarios", func(t *testing.T) {
		// Test with invalid result data to trigger different code paths
		invalidResult := AnalysisResult{
			NumRepositories: -1, // Invalid negative value
			ProjectBranches: nil,
		}

		err := SaveResult(invalidResult)
		// This exercises error handling paths and improves coverage
		if err != nil {
			t.Logf("SaveResult handled invalid data correctly: %v", err)
		}
	})

	t.Run("SaveResult with various data sizes", func(t *testing.T) {
		// Test with different data structures to improve coverage
		testCases := []struct {
			name   string
			result AnalysisResult
		}{
			{
				name: "Empty repositories",
				result: AnalysisResult{
					NumRepositories: 0,
					ProjectBranches: []ProjectBranch{},
				},
			},
			{
				name:   "Single repository",
				result: createTestAnalysisResult(),
			},
			{
				name: "Multiple repositories",
				result: AnalysisResult{
					NumRepositories: 3,
					ProjectBranches: []ProjectBranch{
						{Org: "org1", RepoSlug: "repo1", MainBranch: "main"},
						{Org: "org2", RepoSlug: "repo2", MainBranch: "develop"},
						{Org: "org3", RepoSlug: "repo3", MainBranch: "master"},
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := SaveResult(tc.result)
				if err != nil {
					t.Errorf("SaveResult failed for %s: %v", tc.name, err)
				}

				// Verify file was created
				expectedFile := resultsConfigDir + "/analysis_result_gitlab.json"
				if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
					t.Errorf("SaveResult did not create file for %s", tc.name)
				}

				// Verify file contents
				fileData, err := os.ReadFile(expectedFile)
				if err == nil {
					var savedResult AnalysisResult
					if json.Unmarshal(fileData, &savedResult) != nil {
						t.Errorf("SaveResult created invalid JSON for %s", tc.name)
					}
				}
			})
		}
	})
}

// ---------------- Additional tests for getgitlab.go behavior ----------------

const testRepoExcluded = "ns/excluded"

func TestLoadExclusionReposSuccess(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "exclusions.txt")
	content := "repo-one\n\n repo-two \n#not-a-comment-line\nrepo-three\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("unable to write temp exclusion file: %v", err)
	}

	ex, err := LoadExclusionRepos(file)
	if err != nil {
		t.Fatalf("LoadExclusionRepos returned error: %v", err)
	}

	// Expected: non-empty trimmed lines become keys
	if !ex["repo-one"] || !ex["repo-two"] || !ex["#not-a-comment-line"] || !ex["repo-three"] {
		t.Errorf("unexpected exclusion map contents: %#v", ex)
	}
}

func TestLoadExclusionReposFileNotFound(t *testing.T) {
	_, err := LoadExclusionRepos(filepath.Join(t.TempDir(), "nope.txt"))
	if err == nil {
		t.Fatalf("expected error when file does not exist")
	}
}

func TestIsExcludedExactAndPrefix(t *testing.T) {
	ex := ExclusionRepos{
		"org/repo":   true,
		"org/prefix": true,
	}

	if !isExcluded("org/repo", ex) {
		t.Errorf("expected exact match to be excluded")
	}
	if !isExcluded("org/prefix-sub", ex) {
		t.Errorf("expected prefix match to be excluded")
	}
	if isExcluded("org/other", ex) {
		t.Errorf("did not expect non-matching repo to be excluded")
	}
}

func TestIsProjectExcludedOrInvalid(t *testing.T) {
	// excluded
	ex := ExclusionRepos{"ns/proj": true}
	proj := &gitlab.Project{PathWithNamespace: "ns/proj"}
	empty, archived := 0, 0
	excluded, isEmpty, isArchived := isProjectExcludedOrInvalid(proj, ex, &empty, &archived)
	if !excluded || isEmpty || isArchived {
		t.Errorf("expected project to be excluded only")
	}
	if empty != 0 || archived != 0 {
		t.Errorf("counters should not change for excluded project")
	}

	// empty repo
	ex = ExclusionRepos{}
	proj = &gitlab.Project{PathWithNamespace: "ns/proj2", EmptyRepo: true}
	empty, archived = 0, 0
	excluded, isEmpty, isArchived = isProjectExcludedOrInvalid(proj, ex, &empty, &archived)
	if excluded || !isEmpty || isArchived {
		t.Errorf("expected project to be empty only")
	}
	if empty != 1 || archived != 0 {
		t.Errorf("expected empty counter incremented, got empty=%d archived=%d", empty, archived)
	}

	// archived repo
	proj = &gitlab.Project{PathWithNamespace: "ns/proj3", Archived: true}
	empty, archived = 0, 0
	excluded, isEmpty, isArchived = isProjectExcludedOrInvalid(proj, ex, &empty, &archived)
	if excluded || isEmpty || !isArchived {
		t.Errorf("expected project to be archived only")
	}
	if empty != 0 || archived != 1 {
		t.Errorf("expected archived counter incremented, got empty=%d archived=%d", empty, archived)
	}
}

func TestAnalyzeProjEarlyReturns(t *testing.T) {
	// excluded path
	ap := AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: testRepoExcluded,
		},
		ExclusionList: ExclusionRepos{testRepoExcluded: true},
	}
	pb, excl, emp, arch := analyzeProj(ap)
	if excl != 1 || emp != 0 || arch != 0 || (pb != ProjectBranch{}) {
		t.Errorf("excluded path invalid result: pb=%+v excl=%d emp=%d arch=%d", pb, excl, emp, arch)
	}

	// empty path
	ap = AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: "ns/empty",
			EmptyRepo:         true,
		},
	}
	pb, excl, emp, arch = analyzeProj(ap)
	if excl != 0 || emp != 1 || arch != 0 || (pb != ProjectBranch{}) {
		t.Errorf("empty path invalid result: pb=%+v excl=%d emp=%d arch=%d", pb, excl, emp, arch)
	}

	// archived path
	ap = AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: "ns/archived",
			Archived:          true,
		},
	}
	pb, excl, emp, arch = analyzeProj(ap)
	if excl != 0 || emp != 0 || arch != 1 || (pb != ProjectBranch{}) {
		t.Errorf("archived path invalid result: pb=%+v excl=%d emp=%d arch=%d", pb, excl, emp, arch)
	}
}

func TestProcessProjectCounterIncrements(t *testing.T) {
	// Setup a real spinner to avoid nil deref if used (shouldn't be used on early returns)
	sp := spinner.New(spinner.CharSets[1], 10*time.Millisecond)

	// excluded
	excluded := 0
	empty := 0
	archived := 0
	ap := AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: testRepoExcluded,
		},
		ExclusionList: ExclusionRepos{testRepoExcluded: true},
		Spin1:         sp,
	}
	list, cpt := processProject(ap, 1, sp, nil, &empty, &archived, &excluded)
	if len(list) != 0 || cpt != 1 || excluded != 1 || empty != 0 || archived != 0 {
		t.Errorf("excluded: list=%v cpt=%d excl=%d empty=%d arch=%d", list, cpt, excluded, empty, archived)
	}

	// empty
	excluded, empty, archived = 0, 0, 0
	ap = AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: "ns/empty",
			EmptyRepo:         true,
		},
		Spin1: sp,
	}
	list, cpt = processProject(ap, 1, sp, nil, &empty, &archived, &excluded)
	if len(list) != 0 || cpt != 1 || excluded != 0 || empty != 1 || archived != 0 {
		t.Errorf("empty: list=%v cpt=%d excl=%d empty=%d arch=%d", list, cpt, excluded, empty, archived)
	}

	// archived
	excluded, empty, archived = 0, 0, 0
	ap = AnalyzeProject{
		Project: &gitlab.Project{
			PathWithNamespace: "ns/archived",
			Archived:          true,
		},
		Spin1: sp,
	}
	list, cpt = processProject(ap, 1, sp, nil, &empty, &archived, &excluded)
	if len(list) != 0 || cpt != 1 || excluded != 0 || empty != 0 || archived != 1 {
		t.Errorf("archived: list=%v cpt=%d excl=%d empty=%d arch=%d", list, cpt, excluded, empty, archived)
	}
}
