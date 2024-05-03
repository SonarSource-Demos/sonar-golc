package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/colussim/GoLC/pkg/sorter"
)

type JsonReporter struct {
	OutputName string
	OutputPath string
}

type languageResult struct {
	Language   string
	Files      int
	Lines      int
	BlankLines int
	Comments   int
	CodeLines  int
}

type fileResult struct {
	File       string
	Lines      int
	BlankLines int
	Comments   int
	CodeLines  int
}

type report struct {
	TotalFiles      int `json:",omitempty"`
	TotalLines      int
	TotalBlankLines int
	TotalComments   int
	TotalCodeLines  int
	Results         interface{}
}

func (j JsonReporter) GenerateReportByLanguage(summary *sorter.SortedSummary) error {
	jsonReport := &report{
		TotalFiles:      summary.TotalFiles,
		TotalLines:      summary.TotalLines,
		TotalBlankLines: summary.TotalBlankLines,
		TotalComments:   summary.TotalComments,
		TotalCodeLines:  summary.TotalCodeLines,
		Results:         []languageResult{},
	}

	for _, r := range summary.Results {
		jsonReport.Results = append(jsonReport.Results.([]languageResult), languageResult{
			Language:   r.Name,
			Files:      summary.FilesByLanguage[r.Name],
			Lines:      r.Lines,
			BlankLines: r.BlankLines,
			Comments:   r.Comments,
			CodeLines:  r.CodeLines,
		})
	}

	return j.writeJson(jsonReport)
}

func (j JsonReporter) GenerateReportByFile(summary *sorter.SortedSummary) error {
	jsonReport := &report{
		TotalLines:      summary.TotalLines,
		TotalBlankLines: summary.TotalBlankLines,
		TotalComments:   summary.TotalComments,
		TotalCodeLines:  summary.TotalCodeLines,
		Results:         []fileResult{},
	}

	for _, r := range summary.Results {
		jsonReport.Results = append(jsonReport.Results.([]fileResult), fileResult{
			File:       r.Name,
			Lines:      r.Lines,
			BlankLines: r.BlankLines,
			Comments:   r.Comments,
			CodeLines:  r.CodeLines,
		})
	}

	return j.writeJson(jsonReport)
}

func (j JsonReporter) writeJson(jsonReport *report) error {
	file, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return err
	}

	outputName := strings.Replace(j.OutputName, "/", "_", -1)
	if !strings.HasSuffix(outputName, ".json") {
		outputName += ".json"
	}

	path := filepath.Join(j.OutputPath, outputName)
	if err := os.WriteFile(path, file, 0644); err != nil {
		return err
	}

	fmt.Printf("\n\t✅ json report exported to %s\n", path)

	return nil
}
