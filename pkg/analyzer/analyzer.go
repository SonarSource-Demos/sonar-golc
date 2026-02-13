package analyzer

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// TestExclusion holds SonarQube-style rules to exclude test code from being counted.
// When non-nil and any rule is non-empty, files matching any rule are excluded.
type TestExclusion struct {
	// FileNamePrefixes: exclude if the filename (base) starts with any of these (e.g. "test").
	FileNamePrefixes []string
	// FileNameContains: exclude if the filename contains any of these substrings (e.g. "test.", "tests.").
	FileNameContains []string
	// DirNames: exclude if any directory in the file path has this exact name (e.g. "doc", "docs", "test", "tests", "mock", "mocks").
	DirNames []string
	// DirNameSuffixes: exclude if any directory in the path has a name ending with any of these (e.g. "test", "tests").
	DirNameSuffixes []string
}

type Analyzer struct {
	SupportedExtensions map[string]string
	path                string
	excludePaths        []string
	excludeExtensions   map[string]bool
	includeExtensions   map[string]bool
	testExclusion       *TestExclusion
}

type FileMetadata struct {
	FilePath  string
	Extension string
	Language  string
}

func NewAnalyzer(
	path string,
	excludePaths []string,
	excludeExtensions map[string]bool,
	includeExtensions map[string]bool,
	extensions map[string]string,
	testExclusion *TestExclusion,
) *Analyzer {
	return &Analyzer{
		SupportedExtensions: extensions,
		path:                path,
		excludePaths:        excludePaths,
		excludeExtensions:   excludeExtensions,
		includeExtensions:   includeExtensions,
		testExclusion:       testExclusion,
	}
}

func (a *Analyzer) MatchingFiles() ([]FileMetadata, error) {
	var files []FileMetadata

	err := filepath.Walk(a.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileExtension := a.getFileExtension(path)
		if a.canAdd(path, fileExtension) {
			fm := FileMetadata{
				FilePath:  path,
				Extension: fileExtension,
				Language:  a.SupportedExtensions[fileExtension],
			}
			files = append(files, fm)
		}

		return nil
	})

	return files, err
}

func (a *Analyzer) getFileExtension(path string) string {
	extension := filepath.Ext(path)

	if extension == "" {
		extension = filepath.Base(path)
	}

	return extension
}

func (a *Analyzer) canAdd(path string, extension string) bool {
	for _, pathToExclude := range a.excludePaths {
		if strings.HasPrefix(path, pathToExclude) {
			return false
		}
	}

	// SonarQube-style test code exclusion (configurable via config.json TestExclusion)
	if a.testExclusion != nil && a.isTestCode(path) {
		return false
	}

	if len(a.includeExtensions) > 0 {
		_, ok := a.includeExtensions[a.getFileExtension(path)]
		return ok
	}

	if _, ok := a.excludeExtensions[a.getFileExtension(path)]; ok {
		return false
	}

	_, ok := a.SupportedExtensions[extension]
	return ok
}

// isTestCode returns true if the file path matches any SonarQube-style test exclusion rule.
// Matching is case-insensitive so that e.g. "Test", "Tests", "test", "tests" all match.
func (a *Analyzer) isTestCode(path string) bool {
	t := a.testExclusion
	base := filepath.Base(path)
	baseLower := strings.ToLower(base)
	dir := filepath.Dir(path)
	segments := strings.Split(filepath.ToSlash(dir), "/")

	// Filename starts with any of FileNamePrefixes (e.g. "test")
	for _, p := range t.FileNamePrefixes {
		if p != "" && strings.HasPrefix(baseLower, strings.ToLower(p)) {
			return true
		}
	}

	// Filename contains any of FileNameContains (e.g. "test.", "tests.")
	for _, sub := range t.FileNameContains {
		if sub != "" && strings.Contains(baseLower, strings.ToLower(sub)) {
			return true
		}
	}

	// Any directory in path has exact name in DirNames (e.g. doc, docs, test, tests, mock, mocks)
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		for _, d := range t.DirNames {
			if d != "" && strings.EqualFold(segment, d) {
				return true
			}
		}
	}

	// Any directory in path has name ending with DirNameSuffixes (e.g. "test", "tests")
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		segLower := strings.ToLower(segment)
		for _, suffix := range t.DirNameSuffixes {
			if suffix != "" && strings.HasSuffix(segLower, strings.ToLower(suffix)) {
				return true
			}
		}
	}

	return false
}
