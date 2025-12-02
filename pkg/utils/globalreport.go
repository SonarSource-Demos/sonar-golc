package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type FileData struct {
	Results []LanguageData1 `json:"Results"`
}

type LanguageData1 struct {
	Language  string `json:"Language"`
	CodeLines int    `json:"CodeLines"`
}

type LanguageData struct {
	Language   string  `json:"Language"`
	CodeLines  int     `json:"CodeLines"`
	Percentage float64 `json:"Percentage"`
	CodeLinesF string  `json:"CodeLinesF"`
}

type Globalinfo struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
	DevOpsPlatform         string `json:"DevOpsPlatform"`
	NumberRepos            int    `json:"NumberRepos"`
}

func (l *LanguageData) FormatCodeLines() {
	l.CodeLinesF = FormatCodeLines(float64(l.CodeLines))
}

func getTotalCodeLines(languages []LanguageData) int {
	total := 0
	for _, lang := range languages {
		total += lang.CodeLines
	}
	return total
}

func CreateGlobalReport(directory string) error {

	//directory := "Results"
	loggers := NewLogger()

	totals, err := collectLanguageTotals(directory)
	if err != nil {
		loggers.Errorf("❌ Error reading files : %v", err)
		return err
	}

	// Persist code_lines_by_language.json and keep marshaled bytes for later
	outputData, err := writeLanguageTotalsJSON(totals)
	if err != nil {
		loggers.Errorf("❌ Error creating output JSON file : %v", err)
		return err
	}

	// Reading data from the GlobalReport JSON file
	ginfo, err := readGlobalInfoFromFile("Results/GlobalReport.json")
	if err != nil {
		return err
	}

	// JSON data decoding
	var languages []LanguageData
	err = json.Unmarshal(outputData, &languages)
	if err != nil {
		loggers.Errorf("❌ Error decoding JSON data : %v", err)
		return err
	}

	// Create a PDF
	if err := renderGlobalPDF(languages, ginfo); err != nil {
		return err
	}

	loggers.Infof("✅ Gobal PDF report exported to %s", "Results/GlobalReport.pdf")
	return nil
}

// collectLanguageTotals walks result files and aggregates language totals.
func collectLanguageTotals(directory string) (map[string]int, error) {
	ligneDeCodeParLangage := make(map[string]int)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasPrefix(info.Name(), "Result_") {
			return nil
		}
		// Skip file-level reports to avoid double counting
		if strings.Contains(info.Name(), "_byfile") {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		fileData, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var data FileData
		if err := json.Unmarshal(fileData, &data); err != nil {
			return err
		}
		for _, result := range data.Results {
			if strings.TrimSpace(result.Language) == "" {
				continue
			}
			ligneDeCodeParLangage[result.Language] += result.CodeLines
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ligneDeCodeParLangage, nil
}

// writeLanguageTotalsJSON writes Results/code_lines_by_language.json and returns the serialized bytes.
func writeLanguageTotalsJSON(totals map[string]int) ([]byte, error) {
	loggers := NewLogger()
	var resultats []LanguageData1
	for lang, total := range totals {
		resultats = append(resultats, LanguageData1{
			Language:  lang,
			CodeLines: total,
		})
	}
	outputData, err := json.MarshalIndent(resultats, "", "  ")
	if err != nil {
		return nil, err
	}
	const outputFile = "Results/code_lines_by_language.json"
	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return nil, err
	}
	loggers.Infof("✅ Results analysis recorded in %s", outputFile)
	return outputData, nil
}

// readGlobalInfoFromFile reads Results/GlobalReport.json into Globalinfo.
func readGlobalInfoFromFile(path string) (Globalinfo, error) {
	loggers := NewLogger()
	data, err := os.ReadFile(path)
	if err != nil {
		loggers.Errorf("❌ Error reading GlobalReport.json file : %v", err)
		return Globalinfo{}, err
	}
	var g Globalinfo
	if err := json.Unmarshal(data, &g); err != nil {
		loggers.Errorf("❌ Error decoding JSON GlobalReport.json file : %v", err)
		return Globalinfo{}, err
	}
	return g, nil
}

// renderGlobalPDF generates the GlobalReport.pdf from languages and global info.
func renderGlobalPDF(languages []LanguageData, ginfo Globalinfo) error {
	var unit = "%"
	loggers := NewLogger()
	Org := "Organization : " + ginfo.Organization
	Tloc := "Total lines Of code : " + ginfo.TotalLinesOfCode
	Lrepos := "Largest Repository : " + ginfo.LargestRepository
	Lrepoloc := "Lines of code largest Repository : " + ginfo.LinesOfCodeLargestRepo
	NBrepos := fmt.Sprintf("Number of Repositories analyzed : %d", ginfo.NumberRepos)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	logoPath := "imgs/Logob.png"
	//pdf.Image(logoPath, 10, 10, 50, 0, false, "", 0, "")

	pdf.Ln(10)
	pdf.SetFont("Times", "B", 12)
	pdf.Cell(0, 20, "Global Report")
	pdf.Ln(20)

	pdf.SetFont("Times", "B", 12)
	pdf.SetFillColor(51, 153, 255)
	pdf.CellFormat(100, 10, Org, "0", 1, "", true, 0, "")
	pdf.SetFont("Times", "", 10)
	pdf.SetFillColor(102, 178, 255)
	pdf.CellFormat(100, 10, Tloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepos, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepoloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, NBrepos, "0", 1, "", true, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Times", "B", 12)
	pdf.SetFillColor(51, 153, 255)

	rowCount := 0
	const maxRowsPerPage = 14

	for _, lang := range languages {
		if rowCount%maxRowsPerPage == 0 {
			if rowCount != 0 {
				pdf.AddPage()
			}

			pdf.Image(logoPath, 10, 10, 50, 0, false, "", 0, "")
			pdf.Ln(10)
			pdf.SetFont("Times", "B", 12)
			pdf.SetFillColor(51, 153, 255)
			pdf.CellFormat(100, 10, "Languages :", "0", 1, "", true, 0, "")
			pdf.SetFont("Times", "", 10)
			pdf.SetFillColor(102, 178, 255)
			rowCount = 0

		}

		lang.Percentage = float64(lang.CodeLines) / float64(getTotalCodeLines(languages)) * 100
		lang.CodeLinesF = fmt.Sprintf("%d", lang.CodeLines)
		pdf.SetFont("Times", "", 10)
		pdflang := fmt.Sprintf("%s : %.2f %s - %s LOC", lang.Language, lang.Percentage, unit, lang.CodeLinesF)
		pdf.CellFormat(100, 10, pdflang, "0", 1, "", true, 0, "")
		rowCount++
	}

	if err := pdf.OutputFileAndClose("Results/GlobalReport.pdf"); err != nil {
		loggers.Errorf("❌ Error saving PDF file: %v", err)
		return err
	}
	return nil
}
