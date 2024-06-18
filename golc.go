package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SonarSource-Demos/sonar-golc/assets"
	"github.com/SonarSource-Demos/sonar-golc/pkg/goloc"
	"github.com/briandowns/spinner"

	getbibucket "github.com/SonarSource-Demos/sonar-golc/pkg/devops/getbitbucket/v2"
	getbibucketdc "github.com/SonarSource-Demos/sonar-golc/pkg/devops/getbitbucketdc"
	"github.com/SonarSource-Demos/sonar-golc/pkg/devops/getgithub"
	"github.com/SonarSource-Demos/sonar-golc/pkg/devops/getgitlab"
	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
)

type OrganizationData struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
	DevOpsPlatform         string `json:"DevOpsPlatform"`
	NumberRepos            int    `json:"NumberRepos"`
}

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path"`
}

type Project struct {
	KEY    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  Links  `json:"links"`
}

type Links struct {
	Self []SelfLink `json:"self"`
}

type SelfLink struct {
	Href string `json:"href"`
}

type Config struct {
	Platforms map[string]interface{} `json:"platforms"`
}

type Report struct {
	TotalFiles      int `json:",omitempty"`
	TotalLines      int
	TotalBlankLines int
	TotalComments   int
	TotalCodeLines  int
	Results         interface{}
}

type Result struct {
	TotalFiles      int           `json:"TotalFiles"`
	TotalLines      int           `json:"TotalLines"`
	TotalBlankLines int           `json:"TotalBlankLines"`
	TotalComments   int           `json:"TotalComments"`
	TotalCodeLines  int           `json:"TotalCodeLines"`
	Results         []LanguageRes `json:"Results"`
}

type LanguageRes struct {
	Language   string `json:"Language"`
	Files      int    `json:"Files"`
	Lines      int    `json:"Lines"`
	BlankLines int    `json:"BlankLines"`
	Comments   int    `json:"Comments"`
	CodeLines  int    `json:"CodeLines"`
}

type RepoParams struct {
	ProjectKey string
	Namespace  string
	RepoSlug   string
	MainBranch string
	PathToScan string
}

const errorMessageRepo = "\n❌ Error Analyse Repositories: "
const errorMessageDi = "❌ Error deleting Repository Directory: %v\n"
const errorMessageAnalyse = "❌ No Analysis performed...\n"

var logFile *os.File

func OpenLogFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = file
	return nil
}

// CloseLogFile closes the log file.
func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogError writes an error message to the log file along with additional context information.
func LogError(message string, err error) {
	if logFile != nil {
		logFile.WriteString(fmt.Sprintf("[%s] ERROR: %s - %v\n", time.Now().Format(time.RFC3339), message, err))
	}
}

// Check Exclusion File Exist
func getFileNameIfExists(filePath string) string {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			//The file does not exist
			return "0"
		} else {
			// Check file
			fmt.Printf("❌ Error check file exclusion: %v\n", err)
			return "0"
		}
	} else {
		return filePath
	}
}

// Load Config File
func LoadConfig(filename string) (Config, error) {
	var config Config

	// Lire le contenu du fichier de configuration
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("❌ failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("❌ failed to parse config JSON: %w", err)
	}

	return config, nil
}

// Parse Result Files in JSON Format
func parseJSONFile(filePath, reponame string) int {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Error reading file:", err)
	}

	var report Report
	err = json.Unmarshal(file, &report)
	if err != nil {
		fmt.Println("❌ Error parsing JSON:", err)
	}

	return report.TotalCodeLines
}

// Create a Bakup File for Result directory
func createBackup(sourceDir, pwd string) error {
	backupDir := filepath.Join(pwd, "Saves")
	backupFilePath := generateBackupFilePath(sourceDir, backupDir)

	if err := createBackupDirectory(backupDir); err != nil {
		return err
	}

	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return fmt.Errorf("error creating backup file: %s", err)
	}
	defer backupFile.Close()

	zipWriter := zip.NewWriter(backupFile)
	defer zipWriter.Close()

	if err := addFilesToBackup(sourceDir, zipWriter); err != nil {
		return err
	}

	fmt.Println("✅ Backup created successfully:", backupFilePath)
	return nil
}

// Generate Backup File Name
func generateBackupFilePath(sourceDir, backupDir string) string {
	backupFileName := fmt.Sprintf("%s_%s.zip", filepath.Base(sourceDir), time.Now().Format("2006-01-02_15-04-05"))
	return filepath.Join(backupDir, backupFileName)
}

// Create a backup Directory
func createBackupDirectory(backupDir string) error {
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("error creating backup directory: %s", err)
		}
	}
	return nil
}

// Add Files in backup
func addFilesToBackup(sourceDir string, zipWriter *zip.Writer) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == sourceDir {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if err := addFileToZip(path, relPath, info, zipWriter); err != nil {
			return err
		}
		return nil
	})
}

func addFileToZip(filePath, relPath string, fileInfo os.FileInfo, zipWriter *zip.Writer) error {
	zipFile, err := zipWriter.Create(relPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generic function to analyze repositories
func AnalyseReposList(DestinationResult string, platformConfig map[string]interface{}, repolist interface{}, analyseRepoFunc func(project interface{}, DestinationResult string, platformConfig map[string]interface{}, spin *spinner.Spinner, results chan int, count *int)) (cpt int) {
	fmt.Print("\n🔎 Analysis of Repos ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Color("green", "bold")
	messageF := ""
	spin.FinalMSG = messageF

	// Create a channel to receive results
	results := make(chan int)
	count := 1

	if platformConfig["Multithreading"].(bool) {
		if len(repolist.([]interface{})) > int(platformConfig["NumberWorkerRepos"].(float64)) {
			// Launch goroutines in batches of X
			X := int(platformConfig["Workers"].(float64))
			batches := len(repolist.([]interface{})) / X
			remainder := len(repolist.([]interface{})) % X
			for i := 0; i < batches; i++ {
				for j := i * X; j < (i+1)*X; j++ {
					go analyseRepoFunc(repolist.([]interface{})[j], DestinationResult, platformConfig, spin, results, &count)
				}
				waitForWorkers(X, results)
			}
			// Launch remaining goroutines
			for i := batches * X; i < batches*X+remainder; i++ {
				go analyseRepoFunc(repolist.([]interface{})[i], DestinationResult, platformConfig, spin, results, &count)
			}
			waitForWorkers(remainder, results)
		} else {
			// Launch goroutines for each repo
			for _, project := range repolist.([]interface{}) {
				go analyseRepoFunc(project, DestinationResult, platformConfig, spin, results, &count)
			}
			waitForWorkers(len(repolist.([]interface{})), results)
		}
	} else {
		// Without multithreading
		for _, project := range repolist.([]interface{}) {
			// Execute the analysis synchronously
			analyseRepoFunc(project, DestinationResult, platformConfig, spin, results, &count)
		}
	}

	return len(repolist.([]interface{}))
}

// Analysis functions for different repository types

// Analysis functions for Bitbucket Cloud
func analyseBitCRepo(project interface{}, DestinationResult string, platformConfig map[string]interface{}, spin *spinner.Spinner, results chan int, count *int) {
	p := project.(getbibucket.ProjectBranch)
	params := RepoParams{
		ProjectKey: p.ProjectKey,
		Namespace:  "",
		RepoSlug:   p.RepoSlug,
		MainBranch: p.MainBranch,
		PathToScan: fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", platformConfig["Protocol"].(string), platformConfig["AccessToken"].(string), platformConfig["Baseapi"].(string), platformConfig["Workspace"].(string), p.RepoSlug),
	}
	performRepoAnalysis(params, DestinationResult, spin, results, count)
}

// Analysis functions for Bitbucket DC
func analyseBitSRVRepo(project interface{}, DestinationResult string, platformConfig map[string]interface{}, trimmedURL string, spin *spinner.Spinner, results chan int, count *int) {
	p := project.(getbibucketdc.ProjectBranch)
	params := RepoParams{
		ProjectKey: p.ProjectKey,
		Namespace:  "",
		RepoSlug:   p.RepoSlug,
		MainBranch: p.MainBranch,
		PathToScan: fmt.Sprintf("%s://%s:%s@%sscm/%s/%s.git", platformConfig["Protocol"].(string), platformConfig["Users"].(string), platformConfig["AccessToken"].(string), trimmedURL, p.ProjectKey, p.RepoSlug),
	}
	performRepoAnalysis(params, DestinationResult, spin, results, count)
}

// Analysis functions for GitHub
func analyseGithubRepo(project interface{}, DestinationResult string, platformConfig map[string]interface{}, spin *spinner.Spinner, results chan int, count *int) {
	p := project.(getgithub.ProjectBranch)
	params := RepoParams{
		ProjectKey: p.Org,
		Namespace:  "",
		RepoSlug:   p.RepoSlug,
		MainBranch: p.MainBranch,
		PathToScan: fmt.Sprintf("%s://%s:x-oauth-basic@%s/%s/%s.git", platformConfig["Protocol"].(string), platformConfig["AccessToken"].(string), platformConfig["Baseapi"].(string), p.Org, p.RepoSlug),
	}
	performRepoAnalysis(params, DestinationResult, spin, results, count)
}

// Analysis functions for GitLab
func analyseGitlabRepo(project interface{}, DestinationResult string, platformConfig map[string]interface{}, spin *spinner.Spinner, results chan int, count *int) {
	p := project.(getgitlab.ProjectBranch)
	params := RepoParams{
		ProjectKey: p.Org,
		Namespace:  p.Namespace,
		RepoSlug:   p.RepoSlug,
		MainBranch: p.MainBranch,
		PathToScan: fmt.Sprintf("%s://gitlab-ci-token:%s@%s/%s.git", platformConfig["Protocol"].(string), platformConfig["AccessToken"].(string), "gitlab.com", p.Namespace),
	}
	performRepoAnalysis(params, DestinationResult, spin, results, count)
}

// Perform repository analysis (common logic)
func performRepoAnalysis(params RepoParams, DestinationResult string, spin *spinner.Spinner, results chan int, count *int) {
	var outputFileName = ""
	if len(params.Namespace) > 0 {
		outputFileName = fmt.Sprintf("Result_%s_%s", params.Namespace, params.MainBranch)
	} else {
		outputFileName = fmt.Sprintf("Result_%s_%s_%s", params.ProjectKey, params.RepoSlug, params.MainBranch)
	}
	golocParams := goloc.Params{
		Path:              params.PathToScan,
		ByFile:            false,
		ExcludePaths:      []string{},
		ExcludeExtensions: []string{},
		IncludeExtensions: []string{},
		OrderByLang:       false,
		OrderByFile:       false,
		OrderByCode:       false,
		OrderByLine:       false,
		OrderByBlank:      false,
		OrderByComment:    false,
		Order:             "DESC",
		OutputName:        outputFileName,
		OutputPath:        DestinationResult,
		ReportFormats:     []string{"json"},
		Branch:            params.MainBranch,
	}
	MessB := fmt.Sprintf("   Extracting files from repo : %s ", params.RepoSlug)
	spin.Suffix = MessB
	spin.Start()

	gc, err := goloc.NewGCloc(golocParams, assets.Languages)
	if err != nil {
		fmt.Println(errorMessageRepo, err)
		return
	}

	gc.Run()
	*count++

	// Remove Repository Directory
	err1 := os.RemoveAll(gc.Repopath)
	if err1 != nil {
		fmt.Printf(errorMessageDi, err1)
	}

	spin.Stop()
	fmt.Printf("\r✅ %d The repository <%s> has been analyzed\n", *count, params.RepoSlug)

	// Send result through channel
	results <- 1
}

// Wait for all goroutines to complete
func waitForWorkers(numWorkers int, results chan int) {
	for i := 0; i < numWorkers; i++ {
		fmt.Printf("\r Waiting for workers...\n")
		<-results
	}
}

// Specific analysis functions calling the generic one

// Analysis function call for BitBucket Cloud
func AnalyseReposListBitC(DestinationResult string, platformConfig map[string]interface{}, repolist []getbibucket.ProjectBranch) (cpt int) {
	repoInterfaces := make([]interface{}, len(repolist))
	for i, v := range repolist {
		repoInterfaces[i] = v
	}
	return AnalyseReposList(DestinationResult, platformConfig, repoInterfaces, analyseBitCRepo)
}

// Analysis function call for BitBucket DC
func AnalyseReposListBitSRV(DestinationResult string, platformConfig map[string]interface{}, repolist []getbibucketdc.ProjectBranch) (cpt int) {
	URLcut := platformConfig["Protocol"].(string) + "://"
	trimmedURL := strings.TrimPrefix(platformConfig["Url"].(string), URLcut)
	repoInterfaces := make([]interface{}, len(repolist))
	for i, v := range repolist {
		repoInterfaces[i] = v
	}
	return AnalyseReposList(DestinationResult, platformConfig, repoInterfaces, func(project interface{}, DestinationResult string, platformConfig map[string]interface{}, spin *spinner.Spinner, results chan int, count *int) {
		analyseBitSRVRepo(project, DestinationResult, platformConfig, trimmedURL, spin, results, count)
	})
}

// Analysis function call for GitHub
func AnalyseReposListGithub(DestinationResult string, platformConfig map[string]interface{}, repolist []getgithub.ProjectBranch) (cpt int) {
	repoInterfaces := make([]interface{}, len(repolist))
	for i, v := range repolist {
		repoInterfaces[i] = v
	}
	return AnalyseReposList(DestinationResult, platformConfig, repoInterfaces, analyseGithubRepo)
}

// Analysis function call for Gitlab
func AnalyseReposListGitlab(DestinationResult string, platformConfig map[string]interface{}, repolist []getgitlab.ProjectBranch) (cpt int) {
	repoInterfaces := make([]interface{}, len(repolist))
	for i, v := range repolist {
		repoInterfaces[i] = v
	}
	return AnalyseReposList(DestinationResult, platformConfig, repoInterfaces, analyseGitlabRepo)
}

/* ---------------- Analyse Directory ---------------- */

func AnalyseReposListFile(Listdirectorie, fileexclusionEX []string) {

	fmt.Print("\n🔎 Analysis of Directories ...\n")

	var wg sync.WaitGroup
	wg.Add(len(Listdirectorie))

	for _, Listdirectories := range Listdirectorie {
		go func(dir string) {
			defer wg.Done()

			spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin.Color("green", "bold")
			messageF := ""
			spin.FinalMSG = messageF

			outputFileName := "Result_"
			count := 1

			params := goloc.Params{
				Path:              dir,
				ByFile:            false,
				ExcludePaths:      fileexclusionEX,
				ExcludeExtensions: []string{},
				IncludeExtensions: []string{},
				OrderByLang:       false,
				OrderByFile:       false,
				OrderByCode:       false,
				OrderByLine:       false,
				OrderByBlank:      false,
				OrderByComment:    false,
				Order:             "DESC",
				OutputName:        outputFileName,
				OutputPath:        "Results",
				ReportFormats:     []string{"json"},
				Branch:            "",
				Token:             "",
			}

			gc, err := goloc.NewGCloc(params, assets.Languages)
			if err != nil {
				fmt.Println(errorMessageRepo, err)
				return
			}

			gc.Run()
			spin.Stop()
			fmt.Printf("\r\t✅ %d The directory <%s> has been analyzed\n", count, dir)
			count++
		}(Listdirectories)
	}

	wg.Wait()
}

/* ---------------- End Analyse Directory ---------------- */

func AnalyseRun(params goloc.Params, reponame string) {
	gc, err := goloc.NewGCloc(params, assets.Languages)
	if err != nil {
		fmt.Println(errorMessageRepo, err)
		os.Exit(1)
	}

	gc.Run()
}

func AnalyseRepo(DestinationResult string, Users string, AccessToken string, DevOps string, Organization string, reponame string) (cpt int) {

	//pathToScan := fmt.Sprintf("git::https://%s@%s.com/%s/%s", AccessToken, DevOps, Organization, reponame)
	pathToScan := fmt.Sprintf("https://%s:%s@%s.com/%s/%s", Users, AccessToken, DevOps, Organization, reponame)

	outputFileName := fmt.Sprintf("Result_%s", reponame)
	params := goloc.Params{
		Path:              pathToScan,
		ByFile:            false,
		ExcludePaths:      []string{},
		ExcludeExtensions: []string{},
		IncludeExtensions: []string{},
		OrderByLang:       false,
		OrderByFile:       false,
		OrderByCode:       false,
		OrderByLine:       false,
		OrderByBlank:      false,
		OrderByComment:    false,
		Order:             "DESC",
		OutputName:        outputFileName,
		OutputPath:        DestinationResult,
		ReportFormats:     []string{"json"},
	}
	gc, err := goloc.NewGCloc(params, assets.Languages)
	if err != nil {
		fmt.Println(errorMessageRepo, err)
		os.Exit(1)
	}

	gc.Run()
	cpt++

	// Remove Repository Directory
	err1 := os.RemoveAll(gc.Repopath)
	if err != nil {
		fmt.Printf(errorMessageDi, err1)
		return
	}

	return cpt
}

// Function Read LoadFile for list of directories
func ReadLines(filename string) ([]string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func displayLanguages() {
	fmt.Printf("%-18s | %-78s | %-15s | %s\n", "Language", "Extensions", "Single Comments", "Multi Line Comments")
	fmt.Println("-------------------+--------------------------------------------------------------------------------+-----------------+--------------------")

	for lang, config := range assets.Languages {
		extensions := strings.Join(config.Extensions, ", ") // Concatenate extensions with comma separator

		singleComments := strings.Join(config.LineComments, ", ") // Concatenate single comments with comma separator

		multiLineComments := ""
		for _, comments := range config.MultiLineComments {
			for _, comment := range comments {
				multiLineComments += comment + " "
			}
		}

		fmt.Printf("%-18s | %-78s | %-15s | %s\n", lang, extensions, singleComments, multiLineComments)
	}
}

func main() {

	var maxTotalCodeLines int
	var maxProject, maxRepo string
	var NumberRepos int
	var startTime time.Time
	var ListDirectory []string
	var ListExclusion []string
	var message3, message4, message5 string

	// Test command line Flags

	devopsFlag := flag.String("devops", "", "Specify the DevOps platform")
	fastFlag := flag.Bool("fast", false, "Enable fast mode (only for Github)")
	helpFlag := flag.Bool("help", false, "Show help message")
	languagesFlag := flag.Bool("languages", false, "Show all supported languages")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: golc -devops [OPTIONS]")
		fmt.Println("Options:  <BitBucketSRV>||<BitBucket>||<Github>||<Gitlab>||<Azure>||<File>")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *languagesFlag {
		displayLanguages()
		os.Exit(0) // Exit after displaying languages
	}

	if *devopsFlag == "" {
		fmt.Println("\n❌ Please specify the DevOps platform using the -devops flag : <BitBucketSRV>||<BitBucket>||<Github>||<Gitlab>||<Azure>||<File>")
		fmt.Println("✅ Example for BitBucket server : golc -devops BitBucketSRV")
		os.Exit(1)
	}

	AppConfig, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("\n❌ Failed to load config: %s", err)
		os.Exit(1)
	}

	// Temporary function for future functionality
	if *devopsFlag == "Azure" {
		fmt.Println("❗️ Functionality coming soon...")
		os.Exit(0)
	}

	platformConfig, ok := AppConfig.Platforms[*devopsFlag].(map[string]interface{})
	if !ok {
		fmt.Printf("\n❌ Configuration for DevOps platform '%s' not found\n", *devopsFlag)
		fmt.Println("✅ the -devops flag is : <BitBucketSRV>||<BitBucket>||<Github>||<Gitlab>||<Azure>||<File>")
		os.Exit(1)
	}

	fmt.Printf("\n✅ Using configuration for DevOps platform '%s'\n", *devopsFlag)

	// Test whether to delete the Results directory and save it before deleting.

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	}
	DestinationResult := pwd + "/Results"

	_, err = os.Stat(DestinationResult)
	if err == nil {

		fmt.Printf("❗️ Directory <'%s'> already exists. Do you want to delete it? (y/n): ", DestinationResult)
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {

			fmt.Printf("❗️ Do you want to create a backup of the directory before deleting? (y/n): ")
			var backupResponse string
			fmt.Scanln(&backupResponse)

			if backupResponse == "y" || backupResponse == "Y" {
				// Créer la sauvegarde ZIP
				err := createBackup(DestinationResult, pwd)
				if err != nil {
					fmt.Printf("❌ Error creating backup: %s\n", err)
					os.Exit(1)
				}
			}

			err := os.RemoveAll(DestinationResult)
			if err != nil {
				fmt.Printf("❌ Error deleting directory: %s\n", err)
				os.Exit(1)
			}
			if err := os.MkdirAll(DestinationResult, os.ModePerm); err != nil {
				panic(err)
			}
			ConfigDirectory := DestinationResult + "/config"
			if err := os.MkdirAll(ConfigDirectory, os.ModePerm); err != nil {
				panic(err)
			}
		} else {
			os.Exit(1)
		}

	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(DestinationResult, os.ModePerm); err != nil {
			panic(err)
		}
		ConfigDirectory := DestinationResult + "/config"
		if err := os.MkdirAll(ConfigDirectory, os.ModePerm); err != nil {
			panic(err)
		}

	}
	fmt.Printf("\n")

	// Create Global Report File

	GlobalReport := DestinationResult + "/GlobalReport.txt"
	file, err := os.Create(GlobalReport)
	if err != nil {
		fmt.Println("❌ Error creating file:", err)
		return
	}
	defer file.Close()

	// Open Logs

	err = OpenLogFile("Results/Logs.log")
	if err != nil {
		fmt.Println("❌ Error opening log file:", err)
		return
	}
	defer CloseLogFile()

	// Select DevOps Platform

	switch devops := platformConfig["DevOps"].(string); devops {
	case "github":

		var fileexclusion = ".cloc_github_ignore"
		fileexclusionEX := getFileNameIfExists(fileexclusion)
		var fast bool

		startTime = time.Now()

		if *fastFlag {
			fmt.Println("🚀 Fast mode enabled for Github")
			fast = true
			err := getgithub.FastAnalys(platformConfig, fileexclusionEX)

			if err != nil {
				fmt.Printf("❌ Quick scan Analysis : '%s'", err)
				os.Exit(0)
			}
		} else {
			fast = false

			repositories, err := getgithub.GetRepoGithubList(platformConfig, fileexclusionEX, fast)
			if err != nil {
				fmt.Printf("Error Get Info Repositories in organization '%s' : '%s'", platformConfig["Organization"].(string), err)
				return
			}

			if len(repositories) == 0 {
				fmt.Printf(errorMessageAnalyse)
				os.Exit(1)

			} else {

				NumberRepos = AnalyseReposListGithub(DestinationResult, platformConfig, repositories)

			}
		}

	case "gitlab":

		var fileexclusion = ".cloc_gitlab_ignore"
		fileexclusionEX := getFileNameIfExists(fileexclusion)

		startTime = time.Now()

		gitproject, err := getgitlab.GetRepoGitLabList(platformConfig, fileexclusionEX)
		if err != nil {
			fmt.Printf("Error Get Info Repositories in organization '%s' : '%s'", platformConfig["Organization"].(string), err)
			return
		}

		if len(gitproject) == 0 {
			fmt.Printf(errorMessageAnalyse)
			os.Exit(1)

		} else {
			//os.Exit(1)
			NumberRepos = AnalyseReposListGitlab(DestinationResult, platformConfig, gitproject)

		}

	case "bitbucket_dc":

		var fileexclusion = platformConfig["FileExclusion"].(string)
		fileexclusionEX := getFileNameIfExists(fileexclusion)

		startTime = time.Now()
		projects, err := getbibucketdc.GetProjectBitbucketList(platformConfig, fileexclusionEX)
		if err != nil {
			fmt.Printf("❌ Error Get Info Projects in Bitbucket server '%s' : ", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Printf(errorMessageAnalyse)
			os.Exit(1)

		} else {

			// Run scanning repositories
			NumberRepos = AnalyseReposListBitSRV(DestinationResult, platformConfig, projects)
		}

	case "bitbucket":
		var fileexclusion = platformConfig["FileExclusion"].(string)
		fileexclusionEX := getFileNameIfExists(fileexclusion)

		startTime = time.Now()

		projects1, err := getbibucket.GetProjectBitbucketListCloud(platformConfig, fileexclusionEX)

		if err != nil {
			fmt.Printf("❌ Error Get Info Project(s) in Bitbucket cloud '%v' ", err)
			return
		}
		if len(projects1) == 0 {
			fmt.Printf(errorMessageAnalyse)
			os.Exit(1)

		} else {

			// Run scanning repositories
			NumberRepos = AnalyseReposListBitC(DestinationResult, platformConfig, projects1)
		}

	case "file":

		fileexclusionEX := getFileNameIfExists(platformConfig["FileExclusion"].(string))
		fileload := getFileNameIfExists(platformConfig["FileLoad"].(string))

		if fileexclusionEX != "0" {
			ListExclusion, err = ReadLines(fileexclusionEX)
			if err != nil {
				fmt.Println("❌ Error reading file <.cloc_file_ignore>:", err)
				os.Exit(1)
			}
		} else {
			ListExclusion = make([]string, 0)

		}

		if fileload != "0" {
			ListDirectory, err = ReadLines(fileload)
			if err != nil {
				fmt.Println("❌ Error reading file <.cloc_file_file>:", err)
				os.Exit(1)
			}
			if len(ListDirectory) == 0 {
				ListDirectory = append(ListDirectory, platformConfig["Directory"].(string))
			}
		} else {
			if len(platformConfig["Directory"].(string)) == 0 {
				fmt.Println("❌ No analysis possible, no directory, specified file or specified loading file")
				os.Exit(1)
			} else {
				ListDirectory = append(ListDirectory, platformConfig["Directory"].(string))
			}
		}
		startTime = time.Now()
		AnalyseReposListFile(ListDirectory, ListExclusion)
	}

	// Begin of report file analysis
	fmt.Print("\n🔎 Analyse Report ...\n")
	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Suffix = " Analyse Report..."
	spin.Color("green", "bold")
	spin.Start()

	// List files in the directory
	files, err := os.ReadDir(DestinationResult)
	if err != nil {
		fmt.Println("❌ Error listing files:", err)
		os.Exit(1)
	}

	// Initialize the sum of TotalCodeLines
	totalCodeLinesSum := 0

	// Analyse All file
	for _, file := range files {
		// Check if the file is a JSON file
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Read contents of JSON file
			filePath := filepath.Join(DestinationResult, file.Name())
			jsonData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("\n❌ Error reading file %s: %v\n", file.Name(), err)
				continue
			}

			// Parse JSON content into a Result structure
			var result Result
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				fmt.Printf("\n❌ Error parsing JSON contents of file %s: %v\n", file.Name(), err)
				continue
			}

			totalCodeLinesSum += result.TotalCodeLines

			// Check if this repo has a higher TotalCodeLines than the current maximum
			if result.TotalCodeLines > maxTotalCodeLines {
				maxTotalCodeLines = result.TotalCodeLines
				// Extract project and repo name from file name
				parts := strings.Split(strings.TrimSuffix(file.Name(), ".json"), "_")
				if platformConfig["DevOps"].(string) != "file" {
					maxProject = parts[1]
					maxRepo = parts[2]
				} else {
					maxProject = ""
					maxRepo = parts[1]
					NumberRepos++
				}
			}
		}

	}
	maxTotalCodeLines1 := utils.FormatCodeLines(float64(maxTotalCodeLines))
	totalCodeLinesSum1 := utils.FormatCodeLines(float64(totalCodeLinesSum))

	// Global Result file
	data := OrganizationData{
		Organization:           platformConfig["Organization"].(string),
		TotalLinesOfCode:       totalCodeLinesSum1,
		LargestRepository:      maxRepo,
		LinesOfCodeLargestRepo: maxTotalCodeLines1,
		DevOpsPlatform:         platformConfig["DevOps"].(string),
		NumberRepos:            NumberRepos,
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("\n❌ Error during JSON encoding in Gobal Report:", err)
		return
	}
	// Created Global Result json file
	file1, err := os.Create("Results/GlobalReport.json")
	if err != nil {
		fmt.Println("\n❌ Error during file creation Gobal Report:", err)
		return
	}
	defer file.Close()

	_, err = file1.Write(jsonData)
	if err != nil {
		fmt.Println("\n❌ Error writing to file:", err)
		return
	}

	spin.Stop()

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if platformConfig["DevOps"].(string) != "file" {
		message0 := fmt.Sprintf("\n✅ Number of Repository analyzed in Organization <%s> is %d \n", platformConfig["Organization"].(string), NumberRepos)
		message1 := fmt.Sprintf("✅ The repository with the largest line of code is in project <%s> the repo name is <%s> with <%s> lines of code\n", maxProject, maxRepo, maxTotalCodeLines1)
		message2 := fmt.Sprintf("✅ The total sum of lines of code in Organization <%s> is : %s Lines of Code\n", platformConfig["Organization"].(string), totalCodeLinesSum1)
		message4 = fmt.Sprintf("\n✅ Time elapsed : %02d:%02d:%02d\n", hours, minutes, seconds)
		message3 = message0 + message1 + message2
		message5 = message3 + message4

	} else {
		message0 := fmt.Sprintf("\n✅ Number of Directory analyzed in Organization <%s> is %d \n", platformConfig["Organization"].(string), NumberRepos)
		message2 := fmt.Sprintf("✅ The total sum of lines of code in Organization <%s> is : %s Lines of Code\n", platformConfig["Organization"].(string), totalCodeLinesSum1)
		message4 = fmt.Sprintf("\n✅ Time elapsed : %02d:%02d:%02d\n", hours, minutes, seconds)
		message3 = message0 + message2
		message5 = message3 + message4

	}

	fmt.Println(message3)
	fmt.Println("\n✅ Reports are located in the <'Results'> directory")
	fmt.Println(message4)

	// Write message in Gobal Report File
	_, err = file.WriteString(message5)
	if err != nil {
		fmt.Println("\n❌ Error writing to file:", err)
		return
	}

	fmt.Println("\nℹ️  To generate and visualize results on a web interface, follow these steps: ")
	fmt.Println("\t✅ run : ResultsAll")

}
