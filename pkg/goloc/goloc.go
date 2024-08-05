package goloc

import (
	"fmt"
	"path/filepath"

	"github.com/SonarSource-Demos/sonar-golc/pkg/analyzer"
	"github.com/SonarSource-Demos/sonar-golc/pkg/filesystem"
	"github.com/SonarSource-Demos/sonar-golc/pkg/getter"
	"github.com/SonarSource-Demos/sonar-golc/pkg/gogit"
	"github.com/SonarSource-Demos/sonar-golc/pkg/goloc/language"
	"github.com/SonarSource-Demos/sonar-golc/pkg/reporter"
	"github.com/SonarSource-Demos/sonar-golc/pkg/reporter/json"
	"github.com/SonarSource-Demos/sonar-golc/pkg/reporter/prompt"
	"github.com/SonarSource-Demos/sonar-golc/pkg/scanner"
	"github.com/SonarSource-Demos/sonar-golc/pkg/sorter"
	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
)

type Params struct {
	Path              string
	ByFile            bool
	ExcludePaths      []string
	ExcludeExtensions []string
	IncludeExtensions []string
	OrderByLang       bool
	OrderByFile       bool
	OrderByCode       bool
	OrderByLine       bool
	OrderByBlank      bool
	OrderByComment    bool
	Order             string
	OutputName        string
	OutputPath        string
	ReportFormats     []string
	Branch            string
	Token             string
}

type GCloc struct {
	params    Params
	analyzer  *analyzer.Analyzer
	scanner   *scanner.Scanner
	sorter    sorter.Sorter
	reporters []reporter.Reporter
	Repopath  string
}

func NewGCloc(params Params, languages language.Languages) (*GCloc, error) {
	var path string
	var err error
	loggers := utils.NewLogger()

	if len(params.Branch) != 0 {
		path, err = gogit.Getrepos(params.Path, params.Branch, params.Token)
		if err != nil {
			return nil, err
			//fmt.Println(err)
		}
	} else {
		path, err = getter.Getter(params.Path)
		if err != nil {
			return nil, err
		}

		lastPart := filepath.Base(path)
		if lastPart != "" {
			params.OutputName = fmt.Sprintf("%s%s", params.OutputName, lastPart)
			//fmt.Println("OutputName:", params.OutputName)
		} else {
			//fmt.Println("OutputName:", path)
			//fmt.Println("\n❌ Failed to create OutputName")
			loggers.Errorf("❌ Failed to create OutputName")
		}

		/*	lastSlashIndex := strings.LastIndex(path, "/")
			if lastSlashIndex != -1 {
				lastPart := path[lastSlashIndex+1:]
				params.OutputName = fmt.Sprintf("%s%s", params.OutputName, lastPart)
			} else {
				return nil, fmt.Errorf("\n❌ Failed to created OutputName")

			}*/
	}
	excludePaths, err := filesystem.GetExcludePaths(path, params.ExcludePaths)
	if err != nil {
		return nil, err
	}

	analyzer := analyzer.NewAnalyzer(
		path,
		excludePaths,
		utils.ConvertToMap(params.ExcludeExtensions),
		utils.ConvertToMap(params.IncludeExtensions),
		getExtensionsMap(languages),
	)

	scanner := scanner.NewScanner(languages)

	sorter := getSorter(params.ByFile, params.Order)

	reporters := getReporters(params.ReportFormats, params.OutputName, params.OutputPath)

	return &GCloc{
		params:    params,
		analyzer:  analyzer,
		scanner:   scanner,
		sorter:    sorter,
		reporters: reporters,
		Repopath:  path,
	}, nil
}

func (gc *GCloc) Run() error {
	files, err := gc.analyzer.MatchingFiles()
	if err != nil {
		return err
	}

	scanResult, err := gc.scanner.Scan(files)
	if err != nil {
		return err
	}

	summary := gc.scanner.Summary(scanResult)

	sortedSummary := gc.sortSummary(summary)

	return gc.generateReports(sortedSummary)
}

func (gc *GCloc) ChangeLanguages(languages language.Languages) {
	extensions := getExtensionsMap(languages)
	gc.scanner.SupportedLanguages = languages
	gc.analyzer.SupportedExtensions = extensions
}

func (gc *GCloc) sortSummary(summary *scanner.Summary) *sorter.SortedSummary {
	params := gc.params

	if params.OrderByCode {
		return gc.sorter.OrderByCodeLines(summary)
	}

	if params.OrderByLang {
		return gc.sorter.OrderByLanguage(summary)
	}

	if params.OrderByLine {
		return gc.sorter.OrderByLines(summary)
	}

	if params.OrderByComment {
		return gc.sorter.OrderByComments(summary)
	}

	if params.OrderByBlank {
		return gc.sorter.OrderByBlankLines(summary)
	}

	if params.OrderByFile {
		if languageSorter, ok := gc.sorter.(sorter.LanguageSorter); ok {
			return languageSorter.OrderByFiles(summary)
		}
	}

	return gc.sorter.OrderByCodeLines(summary)
}

func (gc *GCloc) generateReports(sortedSummary *sorter.SortedSummary) error {
	if gc.params.ByFile {
		for _, reporter := range gc.reporters {
			if err := reporter.GenerateReportByFile(sortedSummary); err != nil {
				return err
			}
		}
		return nil
	}

	for _, reporter := range gc.reporters {
		if err := reporter.GenerateReportByLanguage(sortedSummary); err != nil {
			return err
		}
	}

	return nil
}

func getExtensionsMap(languages language.Languages) map[string]string {
	extensions := map[string]string{}

	for language, languageInfo := range languages {
		for _, extension := range languageInfo.Extensions {
			extensions[extension] = language
		}
	}

	return extensions
}

func getSorter(byFile bool, order string) sorter.Sorter {
	if byFile {
		return sorter.NewFileSorter(order)
	}

	return sorter.NewLanguageSorter(order)
}

func getReporters(reportFormats []string, outputName, outputPath string) []reporter.Reporter {
	var reporters []reporter.Reporter

	for _, format := range reportFormats {
		switch format {
		case "prompt":
			reporters = append(reporters, prompt.PromptReporter{})
		case "json":
			reporters = append(reporters, json.JsonReporter{
				OutputName: outputName,
				OutputPath: outputPath,
			})

		default:
			fmt.Printf("%s report format not supported\n", format)
		}
	}

	return reporters
}
