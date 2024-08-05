package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
	"github.com/jung-kurt/gofpdf"
	chart "github.com/wcharczuk/go-chart/v2"
)

type Globalinfo struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
	DevOpsPlatform         string `json:"DevOpsPlatform"`
	NumberRepos            int    `json:"NumberRepos"`
}

type LanguageData struct {
	Language   string  `json:"Language"`
	CodeLines  int     `json:"CodeLines"`
	Percentage float64 `json:"Percentage"`
	CodeLinesF string  `json:"CodeLinesF"`
}

type PageData struct {
	Languages    []LanguageData
	GlobalReport Globalinfo
}

type FileData struct {
	Results []LanguageData1 `json:"Results"`
}

type LanguageData1 struct {
	Language  string `json:"Language"`
	CodeLines int    `json:"CodeLines"`
}

func startServer(port int) {
	fmt.Printf("✅ Server started on http://localhost:%d\n", port)
	fmt.Println("✅ please type < Ctrl+C> to stop the server")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func formatCodeLines(numLines float64) string {
	if numLines >= 1000000 {
		return fmt.Sprintf("%.2fM", numLines/1000000)
	} else if numLines >= 1000 {
		return fmt.Sprintf("%.2fK", numLines/1000)
	} else {
		return fmt.Sprintf("%.0f", numLines)
	}
}

func (l *LanguageData) FormatCodeLines() {
	l.CodeLinesF = utils.FormatCodeLines(float64(l.CodeLines))
}

func isPortOpen(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func main() {
	//var pageData PageData
	directory := "Results"
	var unit string = "%"

	ligneDeCodeParLangage := make(map[string]int)

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
		fmt.Println("❌ Error reading files :", err)
		return
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
		fmt.Println("❌ Error creating output JSON file :", err)
		return
	}
	outputFile := "Results/code_lines_by_language.json"
	err = os.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		fmt.Println("❌ Error writing to output JSON file :", err)
		return
	}

	fmt.Println("✅ Results analysis recorded in", outputFile)

	// Reading data from the GlobalReport JSON file
	data0, err := os.ReadFile("Results/GlobalReport.json")
	if err != nil {
		fmt.Println("❌ Error reading GlobalReport.json file", http.StatusInternalServerError)
		return
	}

	// JSON data decoding
	var Ginfo Globalinfo

	err = json.Unmarshal(data0, &Ginfo)
	if err != nil {
		fmt.Println("❌ Error decoding JSON GlobalReport.json file", http.StatusInternalServerError)
		return
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
		fmt.Println("❌ Error decoding JSON data", http.StatusInternalServerError)
		return
	}

	// Calculating percentages
	total := 0
	for _, lang := range languages {
		total += lang.CodeLines
	}
	for i := range languages {
		languages[i].Percentage = float64(languages[i].CodeLines) / float64(total) * 100
		languages[i].FormatCodeLines()
	}

	// Create the pie chart without labels
	values := []chart.Value{}
	for _, lang := range languages {
		values = append(values, chart.Value{
			Label: lang.Language,
			Value: float64(lang.CodeLines),
		})
	}

	pieChart := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: values,
	}

	// Save the chart as an image
	imagePath := "Results/chart.png"
	f, err := os.Create(imagePath)
	if err != nil {
		fmt.Println("❌ Error creating chart image file:", err)
		return
	}
	defer f.Close()
	err = pieChart.Render(chart.PNG, f)
	if err != nil {
		fmt.Println("❌ Error rendering chart to image:", err)
		return
	}

	// Create a PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.Image("imgs/bg2.jpg", 0, 0, 210, 297, false, "", 0, "")
	logoPath := "imgs/Logo.png"
	pdf.Image(logoPath, 10, 10, 50, 0, false, "", 0, "")
	pdf.Ln(10)
	pdf.Ln(10)
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(255, 255, 255)
	pdf.Cell(0, 10, "Results")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(51, 153, 255)
	pdf.CellFormat(100, 10, Org, "0", 1, "", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(102, 178, 255)
	pdf.CellFormat(100, 10, Tloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepos, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepoloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, NBrepos, "0", 1, "", true, 0, "")
	pdf.Ln(10)

	pdf.AddPage()
	pdf.ImageOptions("imgs/bg2.jpg", 0, 0, 210, 297, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	pdf.Image(logoPath, 10, 10, 50, 0, false, "", 0, "")
	pdf.Ln(10)
	pdf.Ln(10)
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(255, 255, 255)
	pdf.Cell(0, 10, "Statistics and Analysis")
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 10, "Programming languages in %")

	pdf.SetFont("Arial", "B", 10)
	pdf.Ln(10)

	// Create the legend with colors and labels
	for i, language := range languages {
		color := chart.GetDefaultColor(i)
		pdf.SetFillColor(int(color.R), int(color.G), int(color.B))
		pdf.CellFormat(10, 10, "", "1", 0, "", true, 0, "")
		pdf.CellFormat(90, 10, fmt.Sprintf("%s : %.2f %s", language.Language, language.Percentage, unit), "0", 1, "", false, 0, "")
	}

	pdf.Ln(10)

	// Add the chart image to the PDF
	pdf.Image(imagePath, 10, 100, 190, 0, false, "", 0, "")

	// Save the PDF
	err = pdf.OutputFileAndClose("Results/report.pdf")
	if err != nil {
		fmt.Println("❌ Error saving PDF:", err)
		return
	}

	fmt.Println("✅ PDF report generated: Results/report.pdf")
}
