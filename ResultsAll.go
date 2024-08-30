package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
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

var globalInfo Globalinfo       // Variable pour stocker les infos globales
var languageData []LanguageData // Variable pour stocker les données des langages

func getGlobalInfo() Globalinfo {
	return globalInfo
}

func getLanguageData() []LanguageData {
	return languageData
}

func startServer(port int) {
	fmt.Printf("✅ Server started on http://localhost:%d\n", port)
	fmt.Println("✅ Please type < Ctrl+C > to stop the server")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
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

func ZipDirectory(source string, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(filepath.Dir(source), file)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			_, err := zipWriter.Create(relativePath + "/")
			return err
		}

		fileToZip, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		writer, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, fileToZip)
		return err
	})
}

func zipResults(w http.ResponseWriter, r *http.Request) {
	resultsDir := "./Results"
	target := "Results.zip"

	err := ZipDirectory(resultsDir, target)
	if err != nil {
		http.Error(w, "Error creating zip file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=Results.zip")

	http.ServeFile(w, r, "Results.zip")
}

func main() {
	var pageData PageData

	ligneDeCodeParLangage := make(map[string]int)

	// Reading data from the code_lines_by_language.json file
	inputFileData, err := os.ReadFile("Results/code_lines_by_language.json")
	if err != nil {
		fmt.Println("❌ Error reading code_lines_by_language.json file", err)
		return
	}

	err = json.Unmarshal(inputFileData, &languageData)
	if err != nil {
		fmt.Println("❌ Error decoding JSON code_lines_by_language.json file", err)
		return
	}

	// Summarize results by language
	for _, result := range languageData {
		language := result.Language
		codeLines := result.CodeLines
		ligneDeCodeParLangage[language] += codeLines
	}

	var languages []LanguageData
	totalLines := 0
	for lang, total := range ligneDeCodeParLangage {
		totalLines += total
		languages = append(languages, LanguageData{
			Language:   lang,
			CodeLines:  total,
			CodeLinesF: utils.FormatCodeLines(float64(total)),
		})
	}

	for i := range languages {
		languages[i].Percentage = float64(languages[i].CodeLines) / float64(totalLines) * 100
	}

	data0, err := os.ReadFile("Results/GlobalReport.json")
	if err != nil {
		fmt.Println("❌ Error reading GlobalReport.json file", err)
		return
	}

	err = json.Unmarshal(data0, &globalInfo)
	if err != nil {
		fmt.Println("❌ Error decoding JSON GlobalReport.json file", err)
		return
	}

	pageData = PageData{
		Languages:    languages,
		GlobalReport: globalInfo,
	}

	// Load HTML template
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = tmpl.Execute(w, pageData)
		if err != nil {
			http.Error(w, "❌ Error executing HTML template", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			zipResults(w, r)
			return
		}
		http.Error(w, "❌ Method not allowed", http.StatusMethodNotAllowed)
	})

	// API Endpoint for Language Data
	http.HandleFunc("/api/languages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(languageData)
	})

	// API Endpoint for Global Info
	http.HandleFunc("/api/global-info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalInfo)
	})

	/*fmt.Println("Would you like to launch web visualization? (Y/N)")
	var launchWeb string
	fmt.Scanln(&launchWeb)*/

	//if launchWeb == "Y" || launchWeb == "y" {
	fmt.Println("✅ Launching web visualization...")
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
		startServer(8080)
	}
	/*} else {
		fmt.Println("Exiting...")
		os.Exit(0)
	} */
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
    <style>
        .chart-container {
            flex: 1;
        }
        .percentage-container {
            flex: 1;
            padding-left: 20px;
        }
            .modal {
            display: none; 
            position: fixed; 
            z-index: 1; 
            left: 0;
            top: 0;
            width: 100%; 
            height: 100%; 
            overflow: auto; 
            background-color: rgb(0,0,0);
            background-color: rgba(0,0,0,0.4);
            padding-top: 60px;
        }
        .modal-content {
            background-color: #fefefe;
            margin: 5% auto; 
            padding: 20px;
            border: 1px solid #888;
            width: 80%; 
        }
            .close {
            color: #aaa;
            float: right;
            font-size: 28px;
            font-weight: bold;
        }
        .close:hover,
        .close:focus {
            color: black;
            text-decoration: none;
            cursor: pointer;
        }


      .css-xvw69q {
        background-color: rgb(255, 255, 255);
        border: 1px solid rgb(225, 230, 243);
        padding: 1.5rem;
        border-radius: 0.25rem;
      }
       .sw-flex {
        display: flex !important;
      }
      .sw-items-baseline {
       align-items: baseline !important;
      }
      .sw-mt-4 {
        margin-top: 1rem !important;
      }
        .rule-desc, .markdown {
          line-height: 1.5;
      }
    </style>
    <script src="/dist/vendors/chartjs/chart.js"></script>
    <script src="/dist/vendors/bootstrap/js/bootstrap.bundle.min.js"></script>
  </head>
  <body>
    <main class="main" id="top">
      <nav class="navbar navbar-expand-lg fixed-top navbar-dark" data-navbar-on-scroll="data-navbar-on-scroll">
       <div class="container"><a class="navbar-brand" href="index.html"><img src="dist/img//Logo.png" alt="" /></a>
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation"><i class="fa-solid fa-bars text-white fs-3"></i></button>
          <div class="collapse navbar-collapse" id="navbarSupportedContent">
            <ul class="navbar-nav ms-auto mt-2 mt-lg-0">
              <li class="nav-item"><a class="nav-link active" aria-current="page" title="Download Reports" href="/download" target="downloads">Reports</a></li>
              <li class="nav-item"><a class="nav-link" aria-current="page" title="API REF" href="#" id="apiButton">API</a></li>
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
                <div class="card text-white bg-primary mb-4" style="max-width: 24rem;">
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
                    <p class="card-text"><i class="fas fa-code-branch"></i> Total lines of code : {{.GlobalReport.TotalLinesOfCode}}</p>
                    <p class="card-text"><i class="fas fa-folder"></i> Largest Repository : {{.GlobalReport.LargestRepository}}</p>
                    <p class="card-text"><i class="fas fa-code-branch"></i> Lines of code in largest Repository : {{.GlobalReport.LinesOfCodeLargestRepo}}</p>
                    <p class="card-text"><i class="fas fa-code-branch"></i> Number of Repositories analyzed : {{.GlobalReport.NumberRepos}}</p>
                  </div>
                </div>
                <div class="chart-container">
                  <canvas id="camembertChart" width="400" height="400"></canvas>
                </div>
            </div>
            <div class="col-lg-6 mt-3 mt-lg-0">
              <div class="card text-white bg-primary mb-4" style="max-width: 21rem;">
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
        </div>
      </section>
    </main>

     <!-- Modal -->

      <!-- Modal -->
    <div id="apiModal" class="modal modal-lg" >
     <div class="modal-dialog modal-dialog-centered modal-lg">
      <div class="modal-content">
        <span class="close"><i class="fa fa-times-circle"></i></span>
           <div class="css-xvw69q e1wpxmm14">
                 <header class="sw-flex sw-items-baseline">
                    <h3><i class="fa fa-info-circle"></i> API Information</h3>
                 </header>
                 <div class="sw-mt-4 markdown"><i class="fa fa-link"></i> <strong>GET</strong> /api/languages</div>
                 <div class="sw-mt-4 markdown">return a list of language with number of line of code</div>
                 <div class="accordion" id="accordion1">
                    <div class="accordion-item">
                      <h2 class="accordion-header" id="headingOne">
                        <button class="accordion-button" type="button" data-bs-toggle="collapse" data-bs-target="#collapseOne" aria-expanded="false" aria-controls="collapseOne">
                          <strong>Response Example<strong>
                        </button>
                      </h2>
                    <div id="collapseOne" class="accordion-collapse collapse" aria-labelledby="headingOne" data-bs-parent="#accordion1">
                        <div class="accordion-body">
                        <pre><code>
                          {  
                            "Language":"C#",
                            "CodeLines":17826,
                            "Percentage":0,
                            "CodeLinesF":""
                          }
                        </code></pre>
                        </div>
                    </div>
                  </div>
                   <div class="sw-mt-4 markdown"><i class="fa fa-link"></i> <strong>GET</strong> /api/global-info</div>
                   <div class="sw-mt-4 markdown">Returns the global information for the analysis.</div>

                    <div class="accordion" id="accordion2">
                    <div class="accordion-item">
                      <h2 class="accordion-header" id="headingOne2">
                        <button class="accordion-button" type="button" data-bs-toggle="collapse" data-bs-target="#collapseTwo" aria-expanded="false" aria-controls="collapseTwo">
                          <strong>Response Example<strong>
                        </button>
                      </h2>
                    <div id="collapseTwo" class="accordion-collapse collapse" aria-labelledby="headingOne" data-bs-parent="#accordion2">
                        <div class="accordion-body">
                        <pre><code>
                          {  
                            "Organization":	"SonarSource-Demos"
                            "TotalLinesOfCode":	"7.13M"
                            "LargestRepository":	"opencv"
                            "LinesOfCodeLargestRepo":	"2.34M"
                            "DevOpsPlatform":	"github"
                            "NumberRepos":	4
                          }
                        </code></pre>
                        </div>
                    </div>
                  </div>

                   
              </div>
            </div>
      </div>
      </div>
    </div>


   
    <script src="/dist/vendors/chartjs/chart.js"></script>
    <script>
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
                                return context.label + ': ' + context.raw.toLocaleString() + ' LOC';
                            }
                        }
                    }
                }
            }
        });
        var modal = document.getElementById("apiModal");
        var btn = document.getElementById("apiButton");
        var span = document.getElementsByClassName("close")[0];

        btn.onclick = function() {
            modal.style.display = "block";
        }

        span.onclick = function() {
            modal.style.display = "none";
        }

        window.onclick = function(event) {
            if (event.target == modal) {
                modal.style.display = "none";
            }
        }
    </script>
  </body>
</html>
`
