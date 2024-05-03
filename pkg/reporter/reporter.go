package reporter

import "github.com/emmanuel-colussi-sonarsource/sonar-golc/pkg/sorter"

type Reporter interface {
	GenerateReportByLanguage(summary *sorter.SortedSummary) error
	GenerateReportByFile(summary *sorter.SortedSummary) error
}
