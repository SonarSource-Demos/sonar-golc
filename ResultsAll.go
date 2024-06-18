package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
	"github.com/jung-kurt/gofpdf"
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
	var pageData PageData
	directory := "Results"
	var unit string = "%"

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

	/*--------------------------------------------------------------------------------*/

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

	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(51, 153, 255)
	pdf.CellFormat(100, 10, Org, "0", 1, "", true, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(102, 178, 255)
	pdf.CellFormat(100, 10, Tloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepos, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, Lrepoloc, "0", 1, "", true, 0, "")
	pdf.CellFormat(100, 10, NBrepos, "0", 1, "", true, 0, "")
	pdf.Ln(10) // Aller à la ligne

	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(51, 153, 255)
	pdf.CellFormat(100, 10, "Languages :", "0", 1, "", true, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(102, 178, 255)

	// Calculating percentages
	total := 0
	for _, lang := range languages {
		total += lang.CodeLines
	}
	for i := range languages {
		languages[i].Percentage = float64(languages[i].CodeLines) / float64(total) * 100
		languages[i].FormatCodeLines()
		pdflang := fmt.Sprintf("%s : %.2f %s - %s LOC", languages[i].Language, languages[i].Percentage, unit, languages[i].CodeLinesF)
		pdf.CellFormat(100, 10, pdflang, "0", 1, "", true, 0, "")
	}

	// Load HTML template
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	pageData = PageData{
		Languages:    languages,
		GlobalReport: Ginfo,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Run Template
		err = tmpl.Execute(w, pageData)
		if err != nil {
			http.Error(w, "❌ Error executing HTML template", http.StatusInternalServerError)
			return
		}
	})

	// Create a PDF

	err = pdf.OutputFileAndClose("Results/GlobalReport.pdf")
	if err != nil {
		fmt.Println("❌ Error saving PDF file:", err)
		os.Exit(1)
	}

	fmt.Println("✅ PDF generated successfully!")

	fmt.Println("Would you like to launch web visualization? (Y/N)")
	var launchWeb string
	fmt.Scanln(&launchWeb)

	if launchWeb == "Y" || launchWeb == "y" {
		fmt.Println("✅ Launching web visualization...")

		// Start HTTP server

		http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))

		if isPortOpen(8080) {
			fmt.Println("❗️ Port 8080 is already in use.")
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("✅ Please enter the port you wish to use : ")
			portStr, _ := reader.ReadString('\n')
			portStr = strings.TrimSpace(portStr)
			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Println("❌ Invalid port...")
				os.Exit(1)
			}
			if isPortOpen(port) {
				fmt.Printf("❌ Port %d is already in use...\n", port)
				os.Exit(1)
			} else {

				startServer(port)
			}

		} else {

			fmt.Print("❗️ Do you want to use the default port 8080? (Y/n):")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)

			if strings.ToLower(answer) == "n" {

				fmt.Print("✅ Please enter the port you wish to use : ")
				portStr, _ := reader.ReadString('\n')
				portStr = strings.TrimSpace(portStr)
				port, err := strconv.Atoi(portStr)
				if err != nil {
					fmt.Println("❌ Invalid port. Use of default port 8080...")
					port = 8080
					startServer(port)
				} else {
					if isPortOpen(port) {
						fmt.Printf("❌ Port %d is already in use...\n", port)
						os.Exit(1)
					} else {

						startServer(port)
					}
				}

			} else {

				startServer(8080)
			}
		}

	} else {
		// Si l'utilisateur répond autre chose que "Y" ou "y", quitter le programme
		fmt.Println("Exiting...")
		os.Exit(0)
	}
	//select {}

}

// HTML template
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en-US" dir="ltr">

  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Result Go LOC</title>

    <link href="https://fonts.googleapis.com/css2?family=Manrope:wght@200;300;400;500;600;700&amp;display=swap" rel="stylesheet">
    <link href="/dist/css/theme.min.css" rel="stylesheet" type="text/css" />
    <link href="/dist/vendors/fontawesome/css/all.min.css" rel="stylesheet" type="text/css" />
    
  </head>
    <style>
       
        .chart-container {
            flex: 1;
        }
        .percentage-container {
            flex: 1;
            padding-left: 20px;
        }
      
    </style>
    <script src="/dist/vendors/chartjs/chart.js"></script>
</head>
<body>
<main class="main" id="top">
      <nav class="navbar navbar-expand-lg fixed-top navbar-dark" data-navbar-on-scroll="data-navbar-on-scroll">
        <div class="container"><a class="navbar-brand" href="index.html"><img src="dist/img/Logo.png" alt="" /></a>
         <div class="collapse navbar-collapse" id="navbarSupportedContent">
            <ul class="navbar-nav ms-auto mt-2 mt-lg-0">
            </ul>
          </div>
        </div>
      </nav>
      <div class="bg-dark"><img class="img-fluid position-absolute end-0" src="dist/img/bg.png" alt="" />
  
     

    <section>

      <div class="container">
        <div class="row align-items-center py-lg-8 py-6" style="margin-top: -5%">
          <div class="col-lg-6 text-center text-lg-start">
            <h1 class="text-white fs-5 fs-xl-6">Results</h1>     
              <div class="card text-white bg-primary mb-4" style="max-width: 23rem;">
                <h5 class="card-header text-white" style="padding: 1rem 1rem;"> <i class="fas fa-chart-line"></i> Organization: {{.GlobalReport.Organization}}

                {{if eq .GlobalReport.DevOpsPlatform "bitbucket_dc"}}
                    <i class="fab fa-bitbucket"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "bitbucket"}}
                    <i class="fab fa-bitbucket"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "github"}}
                     <i class="fab fa-github"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "gitlab"}}
                    <i class="fab fa-gitlab"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "azure"}}
                    <i class="fab fa-microsoft"></i>
                {{else}}
                    <i class="fas fa-folder"></i>
                {{end}}

                </h5>

                 <div class="card-body" style="padding: 1rem 1rem;">
                   <p class="card-text"><i class="fas fa-code-branch"></i> Total lines Of code : {{.GlobalReport.TotalLinesOfCode}}</p>
                   <p class="card-text"><i class="fas fa-folder"></i> Largest Repository : {{.GlobalReport.LargestRepository}}</p>
                   <p class="card-text"><i class="fas fa-code-branch"></i> Lines of code largest Repository : {{.GlobalReport.LinesOfCodeLargestRepo}}</p>
				   <p class="card-text"><i class="fas fa-code-branch"></i> Number of Repositories analyzed : {{.GlobalReport.NumberRepos}}</p>
                 </div>
               </div>
               <div class="chart-container">
                <canvas id="camembertChart" width="400" height="400" ></canvas>
               </div>
          </div>
          <div class="col-lg-6  mt-3 mt-lg-0">
            <div class="card text-white bg-primary mb-4" style="max-width: 20rem;">
                <h5 class="card-header text-white" style="padding: 1rem 1rem;"><i class="fas fa-code"></i> Languages</h5>
                <div class="card-body text-white" style="padding: 1rem 1rem;">
                    <ul>
                    {{range .Languages}}
                        <li>{{.Language}}: {{printf "%.2f" .Percentage}}% - {{.CodeLinesF}} LOC</li>
                    {{end}}
                    </ul>
                </div>    
            </div>
          </div>
          
         
        </div>
        <div class="swiper">
            
        </div>
     </div>
    </section>

 
</main>

    <script src="/dist/vendors/chartjs/chart.js"></script>
    <script> 

    function formatTooltipLabel(tooltipItem, data) {
        var label =tooltipItem || '';
        var value = data;
        
        var unit = "";
    
        if (value >= 1000000) {
            unit = "M";
            value = (value / 1000000).toFixed(2) + unit;
        } else if (value >= 1000) {
            unit = "K";
            value = (value / 1000).toFixed(2) + unit;
        }
    
        return label + ': ' + value;
    }
    
    function commarize(min) {
        min = min || 1e3;
        // Alter numbers larger than 1k
        if (this >= min) {
          var units = ["k", "M", "B", "T"];
      
          var order = Math.floor(Math.log(this) / Math.log(1000));
      
          var unitname = units[(order - 1)];
          var num = Math.floor(this / 1000 ** order);
      
          // output number remainder + unitname
          return num + unitname
        }
      
        // return formatted original number
        return this.toLocaleString()
      }
      
    
    

        var ctx = document.getElementById('camembertChart').getContext('2d');
        var camembertChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
               labels: [{{range .Languages}}"{{.Language}}",{{end}}],
            
                datasets: [{
                    label: 'LOC ',
                    data: [{{range .Languages}}{{.CodeLines}},{{end}}],
                    backgroundColor: [
                        'rgba(255, 99, 132, 0.5)',
                        'rgba(54, 162, 235, 0.5)',
                        'rgba(255, 206, 86, 0.5)',
                        'rgba(75, 192, 192, 0.5)',
                        'rgba(153, 102, 255, 0.5)',
                        'rgba(255, 159, 64, 0.5)'
                    ],
                    borderColor: [
                        'rgba(255, 99, 132, 1)',
                        'rgba(54, 162, 235, 1)',
                        'rgba(255, 206, 86, 1)',
                        'rgba(75, 192, 192, 1)',
                        'rgba(153, 102, 255, 1)',
                        'rgba(255, 159, 64, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: false,
             
                legend: {
                    display: false
                },
                plugins: {
                    legend: {
                        labels: {
                            color: 'white' 
                        }
                    }, 
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                              // let value1:=context.dataset.data[context.dataIndex] ;
                            //  alert(context.dataset.data[context.dataIndex]);
                              //  alert(context.dataset.data);
                                return formatTooltipLabel(context.dataset.label, context.dataset.data[context.dataIndex]);
                            
                            }
                             
                        }
                    }
                }
                
            }
        });
    </script>
</body>
</html>

`
