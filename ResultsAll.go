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

func startServer(port int) {
	fmt.Printf("✅ Server started on http://localhost:%d\n", port)
	fmt.Println("✅ please type < Ctrl+C> to stop the server")
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
	// Création du fichier zip
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Création d'un nouvel archive zip
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Parcours du répertoire source
	return filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// On construit le chemin relatif pour le zip
		relativePath, err := filepath.Rel(filepath.Dir(source), file)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			// Ajouter le répertoire au zip
			_, err := zipWriter.Create(relativePath + "/")
			return err
		}

		// Ouvrir le fichier à zipper
		fileToZip, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		// Créer une entrée dans le zip
		writer, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		// Copier le contenu du fichier dans l'entrée zip
		_, err = io.Copy(writer, fileToZip)
		return err
	})
}

func zipResults(w http.ResponseWriter, r *http.Request) {
	resultsDir := "./Results"
	target := "Results.zip"

	err := ZipDirectory(resultsDir, target)
	if err != nil {
		http.Error(w, "error creating zip file", http.StatusInternalServerError)
	}

	// Configurer les en-têtes pour le téléchargement
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=Results.zip")

	http.ServeFile(w, r, "Results.zip")
}

func main() {
	var pageData PageData
	//var unit string = "%"

	// Map to hold lines of code by language
	ligneDeCodeParLangage := make(map[string]int)

	// Reading data from the code_lines_by_language.json file
	inputFileData, err := os.ReadFile("Results/code_lines_by_language.json")
	if err != nil {
		fmt.Println("❌ Error reading code_lines_by_language.json file", err)
		return
	}

	var results []LanguageData
	err = json.Unmarshal(inputFileData, &results)
	if err != nil {
		fmt.Println("❌ Error decoding JSON code_lines_by_language.json file", err)
		return
	}

	// Summarize results by language
	for _, result := range results {
		language := result.Language
		codeLines := result.CodeLines
		ligneDeCodeParLangage[language] += codeLines
	}

	// Prepare the data for the HTML template
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

	// Calculate percentages
	for i := range languages {
		languages[i].Percentage = float64(languages[i].CodeLines) / float64(totalLines) * 100
	}

	// Reading data from the GlobalReport JSON file
	data0, err := os.ReadFile("Results/GlobalReport.json")
	if err != nil {
		fmt.Println("❌ Error reading GlobalReport.json file", err)
		return
	}

	// JSON data decoding
	var Ginfo Globalinfo
	err = json.Unmarshal(data0, &Ginfo)
	if err != nil {
		fmt.Println("❌ Error decoding JSON GlobalReport.json file", err)
		return
	}

	pageData = PageData{
		Languages:    languages,
		GlobalReport: Ginfo,
	}

	// Load HTML template
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Run Template
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

	fmt.Println("Would you like to launch web visualization? (Y/N)")
	var launchWeb string
	fmt.Scanln(&launchWeb)

	if launchWeb == "Y" || launchWeb == "y" {
		fmt.Println("✅ Launching web visualization...")
		http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))

		// Start the server
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
	} else {
		fmt.Println("Exiting...")
		os.Exit(0)
	}
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
      
    </style>
   
    <script src="/dist/vendors/chartjs/chart.js"></script>
</head>
<body>
<main class="main" id="top">
      <nav class="navbar navbar-expand-lg fixed-top navbar-dark" data-navbar-on-scroll="data-navbar-on-scroll">
       <div class="container">
                <a class="navbar-brand" href="index.html"><img src="dist/img/Logo.png" alt="" /></a>
                <div class="collapse navbar-collapse" id="navbarSupportedContent">
                    <ul class="navbar-nav ms-auto mt-2 mt-lg-0">
                        <li class="nav-item">
                            <button class="nav-link btn btn-link btn-primary" onclick="window.location.href='/download'" target="downloads" title="Download Results Files">Download Results</button>
                        </li>
                       
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
