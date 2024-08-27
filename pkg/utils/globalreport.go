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
	var unit string = "%"
	loggers := NewLogger()

	ligneDeCodeParLangage := make(map[string]int)

	/*--------------------------------------------------------------------------------*/
	// Results/code_lines_by_language.json file generation

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If the file is not a directory and its name starts with "Result_", then
		if !info.IsDir() && strings.HasPrefix(info.Name(), "Result_") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if filepath.Ext(path) == ".json" {
				// Reading the JSON file
				fileData, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// JSON data decoding
				var data FileData
				err = json.Unmarshal(fileData, &data)
				if err != nil {
					return err
				}

				// Browse results for each file
				for _, result := range data.Results {
					language := result.Language
					codeLines := result.CodeLines
					ligneDeCodeParLangage[language] += codeLines
				}
			}
		}
		return nil
	})
	if err != nil {
		loggers.Errorf("❌ Error reading files : %v", err)
		return err
	}

	// Create output structure
	var resultats []LanguageData1
	for lang, total := range ligneDeCodeParLangage {
		resultats = append(resultats, LanguageData1{
			Language:  lang,
			CodeLines: total,
		})
	}
	// Writing results to a JSON file
	outputData, err := json.MarshalIndent(resultats, "", "  ")
	if err != nil {
		loggers.Errorf("❌ Error creating output JSON file : %v", err)
		return err
	}
	outputFile := "Results/code_lines_by_language.json"
	err = os.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		loggers.Errorf("❌ Error writing to output JSON file : %v", err)
		return err
	}

	loggers.Infof("✅ Results analysis recorded in %s", outputFile)

	// Reading data from the GlobalReport JSON file
	data0, err := os.ReadFile("Results/GlobalReport.json")
	if err != nil {
		loggers.Errorf("❌ Error reading GlobalReport.json file : %v", err)
		return err
	}

	// JSON data decoding
	var Ginfo Globalinfo

	err = json.Unmarshal(data0, &Ginfo)
	if err != nil {
		loggers.Errorf("❌ Error decoding JSON GlobalReport.json file : %v", err)
		return err
	}
	Org := "Organization : " + Ginfo.Organization
	Tloc := "Total lines Of code : " + Ginfo.TotalLinesOfCode
	Lrepos := "Largest Repository : " + Ginfo.LargestRepository
	Lrepoloc := "Lines of code largest Repository : " + Ginfo.LinesOfCodeLargestRepo
	NBrepos := fmt.Sprintf("Number of Repositories analyzed : %d", Ginfo.NumberRepos)

	// JSON data decoding
	var languages []LanguageData
	err = json.Unmarshal(outputData, &languages)
	if err != nil {
		loggers.Errorf("❌ Error decoding JSON data : %v", err)
		return err
	}

	// Create a PDF

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

	err = pdf.OutputFileAndClose("Results/GlobalReport.pdf")
	if err != nil {
		loggers.Errorf("❌ Error saving PDF file: %v", err)
		os.Exit(1)
	}

	loggers.Infof("✅ Gobal PDF report exported to %s", "Results/GlobalReport.pdf")

	return nil

}
