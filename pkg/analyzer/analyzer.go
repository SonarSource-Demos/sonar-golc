package analyzer

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type Analyzer struct {
	SupportedExtensions   map[string]string
	path                  string
	excludePaths          []string
	excludePathSegments   map[string]bool // segment names that exclude a path if any path component matches
	excludeExtensions     map[string]bool
	includeExtensions     map[string]bool
}

type FileMetadata struct {
	FilePath  string
	Extension string
	Language  string
}

func NewAnalyzer(
	path string,
	excludePaths []string,
	excludePathSegments []string,
	excludeExtensions map[string]bool,
	includeExtensions map[string]bool,
	extensions map[string]string,
) *Analyzer {
	segmentSet := make(map[string]bool)
	for _, s := range excludePathSegments {
		segmentSet[s] = true
	}
	return &Analyzer{
		SupportedExtensions: extensions,
		path:                path,
		excludePaths:        excludePaths,
		excludePathSegments: segmentSet,
		excludeExtensions:   excludeExtensions,
		includeExtensions:   includeExtensions,
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

	// Exclude if any path segment (directory or file name) matches ExcludePathSegments
	if len(a.excludePathSegments) > 0 {
		dir := filepath.Dir(path)
		for _, segment := range strings.Split(filepath.ToSlash(dir), "/") {
			if segment != "" && a.excludePathSegments[segment] {
				return false
			}
		}
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
