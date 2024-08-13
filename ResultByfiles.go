package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/briandowns/spinner"
)

type FileResult struct {
	File       string `json:"File"`
	Lines      int    `json:"Lines"`
	BlankLines int    `json:"BlankLines"`
	Comments   int    `json:"Comments"`
	CodeLines  int    `json:"CodeLines"`
}

type RepositoryReport struct {
	TotalLines      int          `json:"TotalLines"`
	TotalBlankLines int          `json:"TotalBlankLines"`
	TotalComments   int          `json:"TotalComments"`
	TotalCodeLines  int          `json:"TotalCodeLines"`
	Results         []FileResult `json:"Results"`
}

func main() {
	// Set the directory where the JSON files are located
	dir := "Results"

	// Set the reports directory
	outputDir := "Results/reports"

	// Create the reports directory if it does not exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	// Read all files in directory
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a list to store reports from JSON files
	var allReportsHTML string
	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Suffix = " Generation of html report files..."
	spin.Color("green", "bold")
	spin.Start()


	// Browse files and process those that start with "Result_"
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "Result_") {
			// Lire le contenu du fichier JSON
			filePath := dir + "/" + file.Name()
			bytes, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Parse JSON
			var report RepositoryReport
			err = json.Unmarshal(bytes, &report)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Extract information from repository and branch
			repoName, branchName := extractRepoAndBranch(file.Name())

			// Generate HTML report for this repository and branch
			reportHTML := generateHTMLReport(report, repoName, branchName)
			allReportsHTML += reportHTML + `<div class="page-break"></div>` // Add to full report with new page

			// Write the HTML report to an individual file in the Results directory
			outputFileName := fmt.Sprintf("report_%s_%s.html", repoName, branchName)
			outputFilePath := filepath.Join(outputDir, outputFileName)
			err = os.WriteFile(outputFilePath, []byte(reportHTML), 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	spin.Stop()
	fmt.Println("\n✅ Results analysis recorded in : Results/reports\n")

	spin.Suffix = " Created Global report File..."
	spin.Color("green", "bold")
	spin.Start()
	

	// Generate the full HTML report
	completeReportHTML := generateCompleteReportHTML(allReportsHTML)
	completeReportPath := filepath.Join(outputDir, "complete_report.html")
	err = os.WriteFile(completeReportPath, []byte(completeReportHTML), 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Convert the full report to PDF
	err = convertHTMLToPDF(completeReportPath, filepath.Join(outputDir, "complete_report.pdf"))
	if err != nil {
		fmt.Println(err)
		return
	}

	spin.Stop()
	fmt.Println("✅ Result analysis recorded in : Results/reports/complete_report.pdf & Results/reports/complete_report.html \n")
}

// Function to extract repository and branch name
func extractRepoAndBranch(fileName string) (string, string) {
	// Remove the "Result_" prefix and the ".json" extension
	baseName := strings.TrimPrefix(fileName, "Result_")
	baseName = strings.TrimSuffix(baseName, ".json")

	// Split the string to get the repository and branch nam
	parts := strings.Split(baseName, "_")
	if len(parts) >= 3 {
		return parts[1], parts[2]
	}
	return parts[0], "No" // Default values ​​in case of error
}

// Function to extract the relative path of the complete file
func extractRelativeFilePath(filePath string) string {

	parts := strings.Split(filePath, "/")

	if len(parts) > 7 {

		return strings.Join(parts[7:], "/")
	}

	return filePath
}

// Function to generate HTML for a repository report
func generateHTMLReport(report RepositoryReport, repoName, branchName string) string {
	reportHTML := fmt.Sprintf(`
<div class="container"><img src="../../imgs/Logob.png" alt="" /></div>	
<h2>Repository Files Details : %s for the branch : %s</h2>
<ul>
	<li>TotalLines: %d</li>
	<li>TotalBlankLines: %d</li>
	<li>TotalComments: %d</li>
	<li>TotalCodeLines: %d</li>
</ul>
<table border="1">
<tr>
	<th>File</th>
	<th>Lines</th>
	<th>BlankLines</th>
	<th>Comments</th>
	<th>CodeLines</th>
</tr>`, repoName, branchName, report.TotalLines, report.TotalBlankLines, report.TotalComments, report.TotalCodeLines)

	for _, result := range report.Results {
		fileName := extractRelativeFilePath(result.File)
		reportHTML += fmt.Sprintf(`<tr>
	<td>%s</td>
	<td>%d</td>
	<td>%d</td>
	<td>%d</td>
	<td>%d</td>
</tr>`, fileName, result.Lines, result.BlankLines, result.Comments, result.CodeLines)
	}

	reportHTML += `</table>`

	return reportHTML
}

// Function to generate the complete HTML
func generateCompleteReportHTML(allReportsHTML string) string {
	return fmt.Sprintf(`
<html>
<head>
    <title>Full Analysis Report</title>
	 
    <style>
        .page-break {
            page-break-before: always;
        }
    </style>
</head>
<body>
	
<h1>Full Analysis Report</h1>

%s
</div>
</body>
</html>`, allReportsHTML)
}

// Function to convert HTML file to PDF
func convertHTMLToPDF(inputPath, outputPath string) error {
	// Déterminer le système d'exploitation et l'architecture
	osType := runtime.GOOS
	archType := runtime.GOARCH

	// Build the path to the wkhtmltopdf binary
	var wkhtmlPath string
	switch osType {
	case "linux":
		if archType == "amd64" {
			wkhtmlPath = "Tools/linux/amd64/wkhtmltopdf"
		} else if archType == "arm64" {
			wkhtmlPath = "Tools/linux/arm64/wkhtmltopdf"
			
			libPath := filepath.Join("Tools", "linux", "arm64", "lib")


			// Get the old LD_LIBRARY_PATH
			oldLDLibraryPath := os.Getenv("LD_LIBRARY_PATH")

			newLDLibraryPath := libPath
    		if oldLDLibraryPath != "" {
       			 newLDLibraryPath = libPath + ":" + oldLDLibraryPath
   			 }

			 fmt.Println("LD_LIBRARY_PATH", newLDLibraryPath)
			// Sets LD_LIBRARY_PATH
			err := os.Setenv("LD_LIBRARY_PATH", newLDLibraryPath)
			if err != nil {
				return fmt.Errorf("❌ could not set LD_LIBRARY_PATH: %v", err)
			}
		
			// Sets LD_PRELOAD to load specific libraries first (optional)
			err = os.Setenv("LD_PRELOAD", filepath.Join(libPath, "libssl.so.1.1")+":"+filepath.Join(libPath, "libcrypto.so.1.1"))
			if err != nil {
				return fmt.Errorf("❌ could not set LD_PRELOAD: %v", err)
			}

		}
	case "darwin": // macOS
		if archType == "amd64" {
			wkhtmlPath = "Tools/darwin/amd64/wkhtmltopdf"
		} else if archType == "arm64" {
			wkhtmlPath = "Tools/darwin/arm64/wkhtmltopdf"
		}
	case "windows":
		if archType == "amd64" {
			wkhtmlPath = "Tools/windows/amd64/wkhtmltopdf.exe"
		} else if archType == "arm64" {
			wkhtmlPath = "Tools/windows/arm64/wkhtmltopdf.exe"
		}
	default:
		return fmt.Errorf("❌ unsupported OS: %s", osType)
	}

	absWkhtmlPath, err := filepath.Abs(wkhtmlPath)
	if err != nil {
		return fmt.Errorf("❌ could not get absolute path for wkhtmltopdf: %v", err)
	}

	// Check if the binary exists
	if _, err := os.Stat(absWkhtmlPath); os.IsNotExist(err) {
		return fmt.Errorf("❌ wkhtmltopdf binary not found at %s", absWkhtmlPath)
	}

	

	// Set Path for wkhtmltopdf binary
	err = os.Setenv("WKHTMLTOPDF_PATH", absWkhtmlPath)
	if err != nil {
		return fmt.Errorf("❌ could not set WKHTMLTOPDF_PATH: %v", err)
	}
	wkhtmltopdf.SetPath(wkhtmlPath)

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		fmt.Println("❌ Error not fund")
		return err
	}

	// Add the HTML file as input
	page := wkhtmltopdf.NewPage(inputPath)
	page.EnableLocalFileAccess.Set(true)

	// Add the page to the PDF generator
	pdfg.AddPage(page)

	// Create PDF
	if err := pdfg.Create(); err != nil {
		return err
	}

	// Save to output file
	err = pdfg.WriteFile(outputPath)
	return err
}
