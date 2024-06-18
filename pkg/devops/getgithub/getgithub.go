package getgithub

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/colussim/GoLC/assets"
	"github.com/google/go-github/v62/github"
)

type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
}
type PlatformConfig struct {
	Organization string
	URL          string
}

type SummaryStats struct {
	LargestRepo       string
	LargestRepoBranch string
	NbRepos           int
	EmptyRepo         int
	TotalExclude      int
	TotalArchiv       int
	TotalBranches     int
}

// RepositoryMap represents a map of repositories to ignore
type ExclusionRepos map[string]bool

type ParamsReposGithub struct {
	Repos         []*github.Repository
	URL           string
	BaseAPI       string
	Apiver        string
	AccessToken   string
	Organization  string
	NBRepos       int
	ExclusionList ExclusionRepos
	Spin          *spinner.Spinner
	Branch        string
	Period        int
	Stats         bool
	DefaultB      bool
}
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"full_name"`
	SizeR         int64  `json:"size"`
	Language      string `json:"language"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	LOC           map[string]int
}

type ProjectBranch struct {
	Org         string
	RepoSlug    string
	MainBranch  string
	LargestSize int64
}

type AnalysisResult struct {
	NumRepositories int
	ProjectBranches []ProjectBranch
}

type TreeItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
	Size int    `json:"size,omitempty"`
}

type TreeResponse struct {
	Sha       string     `json:"sha"`
	Url       string     `json:"url"`
	Tree      []TreeItem `json:"tree"`
	Truncated bool       `json:"truncated"`
}

type Branch struct {
	Name      string     `json:"name"`
	Commit    CommitInfo `json:"commit"`
	Protected bool       `json:"protected"`
}

type CommitInfo struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

type BranchInfoEvents struct {
	Name      string
	Pushes    int
	Commits   int
	Additions int
	Deletions int
}

type Lastanalyse struct {
	LastRepos  string
	LastBranch string
}

type RepoBranch struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	Branches []*github.Branch `json:"branches"`
}

type LanguageInfo1 struct {
	Language  string
	CodeLines int
}

const PrefixMsg = "Get Repo(s)..."
const MessageApiRate = "❗️ Rate limit exceeded. Waiting for rate limit reset..."
const ApiHeader1 = "application/vnd.github.v3+json"
const ErrorMesssage1 = "❌ Error saving repositories in file Results/config/analysis_repos_github.json: %v\n"

// Load repository ignore map from file
func loadExclusionRepos1(filename string) (ExclusionRepos, error) {
	ignoreMap := make(ExclusionRepos)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoName := strings.TrimSpace(scanner.Text())
		if repoName != "" {
			ignoreMap[repoName] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreMap, nil
}

// Check if a repository should be ignored
func shouldIgnore(repoName string, ignoreMap ExclusionRepos) bool {
	_, ignored := ignoreMap[repoName]
	return ignored
}

func SaveResult(result AnalysisResult) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_result_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the result and write it to the file
	if err := encoder.Encode(result); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_result_github.json> :", err)
		return err
	}

	fmt.Println("✅ Result saved successfully!")
	return nil
}

func SaveBranch(branch RepoBranch) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_branch_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Branch file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(branch); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_branch_github.json> :", err)
		return err
	}

	//	fmt.Println("✅ Branch saved successfully!")
	return nil
}

func SaveCommit(repos []*github.RepositoryCommit) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_commit_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Repos file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(repos); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_commit_github.json> :", err)
		return err
	}

	//fmt.Println("✅ Commits saved successfully!")
	return nil
}
func SaveRepos(repos []*github.Repository) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_repos_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Repos file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(repos); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_repos_github.json> :", err)
		return err
	}

	fmt.Println("✅ \r Repos saved successfully!")
	return nil
}

func SaveLast(last Lastanalyse) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_last_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Last file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(last); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_last_github.json> :", err)
		return err
	}

	fmt.Println("✅ Last saved successfully!")
	return nil
}

func GetReposGithub(parms ParamsReposGithub, ctx context.Context, client *github.Client) ([]ProjectBranch, int, int, int, int, int) {
	var TotalBranches, notAnalyzedCount, emptyRepo, cpt, cptarchiv int
	var importantBranches []ProjectBranch
	cpt = 1

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Color("green", "bold")

	message4 := "Repo(s)"
	fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

	for _, repo := range parms.Repos {
		repoName := *repo.Name
		if repo.GetArchived() {
			cptarchiv++
			continue
		}
		if len(parms.ExclusionList) != 0 && shouldIgnore(repoName, parms.ExclusionList) {
			fmt.Printf("\t   ✅ Skipping analysis for repository '%s' as per ignore list.\n", repoName)
			notAnalyzedCount++
			continue
		}
		isEmpty, err := reposIfEmpty(ctx, client, repoName, parms.Organization)
		if err != nil {
			fmt.Print(err.Error())
			continue
		}
		if !isEmpty {
			largestRepoBranch, repoBranches := analyzeRepoBranches(parms, ctx, client, repo, cpt, spin1)
			importantBranches = append(importantBranches, ProjectBranch{
				Org:         parms.Organization,
				RepoSlug:    repoName,
				MainBranch:  largestRepoBranch,
				LargestSize: int64(len(repoBranches)),
			})
			TotalBranches += len(repoBranches)
		} else {
			emptyRepo++
		}
		cpt++
	}

	result := AnalysisResult{
		NumRepositories: parms.NBRepos,
		ProjectBranches: importantBranches,
	}
	if err := SaveResult(result); err != nil {
		fmt.Println("❌ Error Save Result of Analysis :", err)
		os.Exit(1)
	}

	return importantBranches, emptyRepo, parms.NBRepos, TotalBranches, notAnalyzedCount, cptarchiv
}

func analyzeRepoBranches(parms ParamsReposGithub, ctx context.Context, client *github.Client, repo *github.Repository, cpt int, spin1 *spinner.Spinner) (string, []*github.Branch) {
	var branches []*github.Branch
	var allEvents []*github.Event
	var branchPushes map[string]*BranchInfoEvents

	opt := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	messageB := fmt.Sprintf("\t   Analysis top branch(es) in repository <%s> ...", *repo.Name)
	spin1.Prefix = messageB
	spin1.Start()

	var largestRepoBranch string
	var err error
	var nbrbranche int

	if parms.DefaultB {
		// If DefaultBranch is true, retrieve the default branch of the repository
		branch, _, _ := client.Repositories.GetBranch(ctx, parms.Organization, *repo.Name, *repo.DefaultBranch, 0)
		branches = append(branches, branch)
		largestRepoBranch = *repo.DefaultBranch
		nbrbranche = 1

	} else if len(parms.Branch) != 0 {
		// If branch name is provided in params, try to get information about the specified branch
		branch, _, err := client.Repositories.GetBranch(ctx, parms.Organization, *repo.Name, parms.Branch, 0)
		if err == nil {
			// If branch exists, use it
			largestRepoBranch = parms.Branch
			branches = append(branches, branch)
			nbrbranche = len(branches)
		} else {
			// If branch does not exist, use default branch
			branches, err = getAllBranches(ctx, client, *repo.Name, parms.Organization, opt)
			if err != nil {
				fmt.Printf("❌ Error when retrieving branches for repo %v: %v\n", *repo.Name, err)
				spin1.Stop()
				return "", nil
			}
			largestRepoBranch = determineLargestBranch(parms, repo, branchPushes)
			nbrbranche = len(branches)
		}
	} else {
		// If DefaultBranch is false and branch name is not provided, get all branches
		branches, err = getAllBranches(ctx, client, *repo.Name, parms.Organization, opt)
		if err != nil {
			fmt.Printf("❌ Error when retrieving branches for repo %v: %v\n", *repo.Name, err)
			spin1.Stop()
			return "", nil
		}
		largestRepoBranch = determineLargestBranch(parms, repo, branchPushes)
		nbrbranche = len(branches)
	}

	allEvents, err = getAllEvents(ctx, client, *repo.Name, parms.Organization)
	if err != nil {
		fmt.Println("❌ Error fetching repository events:", err)
		spin1.Stop()
		return "", nil
	}

	branchPushes = countBranchPushes(allEvents, parms.Period)
	analyzeBranches(ctx, client, parms, *repo.Name, branchPushes)

	spin1.Stop()

	fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d - largest Branch: %s \n", cpt, *repo.Name, nbrbranche, largestRepoBranch)

	return largestRepoBranch, branches
}

func getAllBranches(ctx context.Context, client *github.Client, repoName, organization string, opt *github.BranchListOptions) ([]*github.Branch, error) {
	var branches []*github.Branch
	for {
		branchPage, resp, err := client.Repositories.ListBranches(ctx, organization, repoName, opt)
		if err != nil {
			if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
				fmt.Println(MessageApiRate)
				time.Sleep(rateLimitErr.GetRetryAfter())
				continue
			}
			return nil, err
		}
		branches = append(branches, branchPage...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return branches, nil
}

func getAllEvents(ctx context.Context, client *github.Client, repoName, organization string) ([]*github.Event, error) {
	var allEvents []*github.Event
	opt := &github.ListOptions{PerPage: 100}
	for {
		events, resp, err := client.Activity.ListRepositoryEvents(ctx, organization, repoName, opt)
		if err != nil {
			if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
				fmt.Println(MessageApiRate)
				time.Sleep(rateLimitErr.GetRetryAfter())
				continue
			}
			return nil, err
		}
		allEvents = append(allEvents, events...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allEvents, nil
}

func countBranchPushes(events []*github.Event, period int) map[string]*BranchInfoEvents {
	branchPushes := make(map[string]*BranchInfoEvents)
	oneMonthAgo := time.Now().AddDate(0, period, 0)
	for _, event := range events {
		if event.CreatedAt != nil && event.CreatedAt.After(oneMonthAgo) {
			switch event.GetType() {
			case "PushEvent":
				payload, err := event.ParsePayload()
				if err != nil {
					fmt.Println("❌ Error parsing payload:", err)
					continue
				}
				pushEvent, ok := payload.(*github.PushEvent)
				if ok {
					branch := pushEvent.GetRef()
					if len(branch) > 11 && branch[:11] == "refs/heads/" {
						branchName := branch[11:]
						if _, exists := branchPushes[branchName]; !exists {
							branchPushes[branchName] = &BranchInfoEvents{Name: branchName}
						}
						branchPushes[branchName].Pushes++
					}
				}
			}
		}
	}
	return branchPushes
}

func analyzeBranches(ctx context.Context, client *github.Client, parms ParamsReposGithub, repoName string, branchPushes map[string]*BranchInfoEvents) {
	oneMonthAgo := time.Now().AddDate(0, parms.Period, 0)
	for _, info := range branchPushes {
		if parms.Stats {
			analyzeWithStats(ctx, client, parms.Organization, repoName, oneMonthAgo, info)
		} else {
			analyzeWithoutStats(ctx, client, parms.Organization, repoName, oneMonthAgo, info)
		}
	}
}

func analyzeWithStats(ctx context.Context, client *github.Client, organization, repoName string, oneMonthAgo time.Time, info *BranchInfoEvents) {
	contributorsStats, _, err := client.Repositories.ListContributorsStats(ctx, organization, repoName)
	if err != nil {
		if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
			fmt.Println(MessageApiRate)
			time.Sleep(rateLimitErr.GetRetryAfter())
			contributorsStats, _, err = client.Repositories.ListContributorsStats(ctx, organization, repoName)
		}
		if err != nil {
			fmt.Printf("❌ Error fetching contributors stats: %v\n", err)
			return
		}
	}

	for _, contributorStats := range contributorsStats {
		for _, week := range contributorStats.Weeks {
			if week.Week.After(oneMonthAgo) {
				info.Additions += *week.Additions
				info.Deletions += *week.Deletions
				info.Commits += *week.Commits
			}
		}
	}
}

func analyzeWithoutStats(ctx context.Context, client *github.Client, organization, repoName string, oneMonthAgo time.Time, info *BranchInfoEvents) {
	opt := &github.CommitsListOptions{
		SHA:         info.Name,
		Since:       oneMonthAgo,
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := client.Repositories.ListCommits(ctx, organization, repoName, opt)
		if err != nil {
			if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
				fmt.Println(MessageApiRate)
				time.Sleep(rateLimitErr.GetRetryAfter())
				commits, resp, err = client.Repositories.ListCommits(ctx, organization, repoName, opt)
			}
			if err != nil {
				fmt.Printf("Error fetching commits for branch %s: %v\n", info.Name, err)
				break
			}
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	info.Commits = len(allCommits)
}

func determineLargestBranch(parms ParamsReposGithub, repo *github.Repository, branchPushes map[string]*BranchInfoEvents) string {
	var largestRepoBranch string
	if len(branchPushes) > 0 {
		branchList := make([]*BranchInfoEvents, 0, len(branchPushes))
		for _, info := range branchPushes {
			branchList = append(branchList, info)
		}
		sort.Slice(branchList, func(i, j int) bool {
			if parms.Stats {
				if branchList[i].Commits == branchList[j].Commits {
					return (branchList[i].Additions + branchList[i].Deletions) > (branchList[j].Additions + branchList[j].Deletions)
				}
				return branchList[i].Commits > branchList[j].Commits
			} else {
				return branchList[i].Commits > branchList[j].Commits
			}
		})
		largestRepoBranch = branchList[0].Name
	} else {
		largestRepoBranch = *repo.DefaultBranch
	}
	return largestRepoBranch
}

// Get Infos for all Repositories in Organization

func GetRepoGithubList(platformConfig map[string]interface{}, exclusionfile string, fast bool) ([]ProjectBranch, error) {
	//var largestRepoSize int64
	var totalSize int64
	var totalExclude, totalArchiv, emptyRepo, TotalBranches, nbRepos int
	var largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var repositories []*github.Repository
	var exclusionList ExclusionRepos
	var err1 error

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")
	spin.Start()

	exclusionList, err1 = loadExclusionFile(exclusionfile, spin)
	if err1 != nil {
		return nil, err1
	}

	ctx, client := initializeGithubClient(platformConfig)

	if len(platformConfig["Repos"].(string)) == 0 {
		repositories, err1 = fetchAllRepositories(ctx, client, platformConfig["Organization"].(string), opt)
	} else {
		repositories, err1 = fetchSingleRepository(ctx, client, platformConfig)
	}

	if err1 != nil {
		return importantBranches, nil
	}

	params := getCommonParams(platformConfig, repositories, exclusionList, spin)
	sortRepositoriesByUpdatedAt(repositories)

	if err := SaveRepos(repositories); err != nil {
		fmt.Printf(ErrorMesssage1, err)
	}

	importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv = GetReposGithub(params, ctx, client)

	largestRepoBranch, largesRepo = findLargestRepository(importantBranches, &totalSize)

	config := PlatformConfig{
		Organization: platformConfig["Organization"].(string),
		URL:          platformConfig["Url"].(string),
	}

	stats := SummaryStats{
		LargestRepo:       largesRepo,
		LargestRepoBranch: largestRepoBranch,
		NbRepos:           nbRepos,
		EmptyRepo:         emptyRepo,
		TotalExclude:      totalExclude,
		TotalArchiv:       totalArchiv,
		TotalBranches:     TotalBranches,
	}

	printSummary(config, stats)

	return importantBranches, nil
}

func loadExclusionFile(exclusionfile string, spin *spinner.Spinner) (ExclusionRepos, error) {
	var exclusionList ExclusionRepos
	var err error

	if exclusionfile == "0" {
		exclusionList = make(map[string]bool)
	} else {
		exclusionList, err = loadExclusionRepos1(exclusionfile)
		if err != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exclusionfile, err)
			spin.Stop()
			return nil, err
		}
	}
	return exclusionList, nil
}

func initializeGithubClient(platformConfig map[string]interface{}) (context.Context, *github.Client) {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(platformConfig["AccessToken"].(string))
	return ctx, client
}

func fetchAllRepositories(ctx context.Context, client *github.Client, organization string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, error) {
	var repositories []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, organization, opt)
		if err != nil {
			fmt.Printf("❌ Error fetching repositories: %v\n", err)
			return nil, err
		}
		repositories = append(repositories, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repositories, nil
}

func fetchSingleRepository(ctx context.Context, client *github.Client, platformConfig map[string]interface{}) ([]*github.Repository, error) {
	repos, _, err := client.Repositories.Get(ctx, platformConfig["Organization"].(string), platformConfig["Repos"].(string))
	if err != nil {
		fmt.Printf("❌ Error fetching repository: %v\n", err)
		return nil, err
	}
	return []*github.Repository{repos}, nil
}

func getCommonParams(platformConfig map[string]interface{}, repositories []*github.Repository, exclusionList ExclusionRepos, spin *spinner.Spinner) ParamsReposGithub {
	return ParamsReposGithub{
		Repos:         repositories,
		URL:           platformConfig["Url"].(string),
		BaseAPI:       platformConfig["Baseapi"].(string),
		Apiver:        platformConfig["Apiver"].(string),
		AccessToken:   platformConfig["AccessToken"].(string),
		Organization:  platformConfig["Organization"].(string),
		NBRepos:       len(repositories),
		ExclusionList: exclusionList,
		Spin:          spin,
		Branch:        platformConfig["Branch"].(string),
		Period:        int(platformConfig["Period"].(float64)),
		Stats:         platformConfig["Stats"].(bool),
		DefaultB:      platformConfig["DefaultBranch"].(bool),
	}
}

func findLargestRepository(importantBranches []ProjectBranch, totalSize *int64) (string, string) {
	var largestRepoSize int64
	var largestRepoBranch, largesRepo string

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug
		}
		*totalSize += branch.LargestSize
	}
	//return largestRepoSize, largestRepoBranch, largesRepo
	return largestRepoBranch, largesRepo
}

func printSummary(config PlatformConfig, stats SummaryStats) {
	fmt.Printf("\n✅ The largest Repository is <%s> in the organization <%s> with the branch <%s> \n", stats.LargestRepo, config.Organization, stats.LargestRepoBranch)
	fmt.Printf("\r✅ Total Repositories that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", stats.NbRepos-stats.EmptyRepo-stats.TotalExclude-stats.TotalArchiv, stats.EmptyRepo, stats.TotalExclude, stats.TotalArchiv)
	fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", stats.TotalBranches)
}

// func FastAnalys(url, baseapi, apiver, accessToken, organization, exlusionfile, repos, branchmain string, period int) error {
func FastAnalys(platformConfig map[string]interface{}, exlusionfile string) error {

	var totalExclude int
	var totalArchiv int
	var repositories []*github.Repository
	var exclusionList ExclusionRepos
	var err1 error
	var emptyRepo int
	nbRepos := 0
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	} // Number Object by page in API Request

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")
	spin.Start()

	// Test if exclusion file exist
	if exlusionfile == "0" {
		exclusionList = make(map[string]bool)

	} else {
		exclusionList, err1 = loadExclusionRepos1(exlusionfile)
		if err1 != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exlusionfile, err1)
			spin.Stop()
			//return nil, err1
		}

	}

	if len(platformConfig["Repos"].(string)) == 0 {

		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(platformConfig["AccessToken"].(string))

		// Get all Repositories in Organization
		for {
			repos, resp, err := client.Repositories.ListByOrg(ctx, platformConfig["Organization"].(string), opt)

			if err != nil {
				fmt.Printf("❌ Error fetching repositories: %v\n", err)
				//return importantBranches, nil
			}

			repositories = append(repositories, repos...)

			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage

		}

		parms := ParamsReposGithub{
			Repos:         repositories,
			URL:           platformConfig["Url"].(string),
			BaseAPI:       platformConfig["Baseapi"].(string),
			Apiver:        platformConfig["Apiver"].(string),
			AccessToken:   platformConfig["AccessToken"].(string),
			Organization:  platformConfig["Organization"].(string),
			NBRepos:       len(repositories),
			ExclusionList: exclusionList,
			Spin:          spin,
			Branch:        platformConfig["Branch"].(string),
			Period:        int(platformConfig["Period"].(float64)),
			Stats:         platformConfig["Stats"].(bool),
		}

		sortRepositoriesByUpdatedAt(repositories)

		// Save List of Repos
		err := SaveRepos(repositories)
		if err != nil {
			fmt.Printf(ErrorMesssage1, err)
		}

		nbRepos, emptyRepo, totalExclude, totalArchiv, err = GetGithubLanguages(parms, ctx, client, int(platformConfig["Factor"].(float64)))
		if err != nil {
			return err
		}

	} else {

		var reposSlice []*github.Repository
		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(platformConfig["AccessToken"].(string))

		repos1, _, err := client.Repositories.Get(ctx, platformConfig["Organization"].(string), platformConfig["Repos"].(string))
		if err != nil {
			fmt.Printf("❌ Error fetching repository: %v\n", err)

		}

		reposSlice = append(reposSlice, repos1)
		parms := ParamsReposGithub{
			Repos:         reposSlice,
			URL:           platformConfig["Url"].(string),
			BaseAPI:       platformConfig["Baseapi"].(string),
			Apiver:        platformConfig["Apiver"].(string),
			AccessToken:   platformConfig["AccessToken"].(string),
			Organization:  platformConfig["Organization"].(string),
			NBRepos:       len(repositories),
			ExclusionList: exclusionList,
			Spin:          spin,
			Branch:        platformConfig["Branch"].(string),
			Period:        int(platformConfig["Period"].(float64)),
			Stats:         platformConfig["Stats"].(bool),
		}
		nbRepos, emptyRepo, totalExclude, totalArchiv, err = GetGithubLanguages(parms, ctx, client, int(platformConfig["Factor"].(float64)))
		if err != nil {
			return err
		}

	}

	fmt.Printf("\r✅ Total Repositories that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", nbRepos-emptyRepo-totalExclude-totalArchiv, emptyRepo, totalExclude, totalArchiv)
	return nil
}

func GetGithubLanguages(parms ParamsReposGithub, ctx context.Context, client *github.Client, factor int) (int, int, int, int, error) {

	cptarchiv := 0        // Counter archiv repos
	notAnalyzedCount := 0 // Counter Number of repositories excluded
	emptyRepo := 0        // Counter Number of repositories empty
	parms.Spin.Stop()
	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Color("green", "bold")

	message4 := "Repo(s)"
	fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

	for _, repo := range parms.Repos {

		repoName := *repo.Name

		// Test if repo is archived
		if repo.GetArchived() {
			cptarchiv++
			continue
		}

		// Test is repo is excluded
		if len(parms.ExclusionList) != 0 {
			if shouldIgnore(repoName, parms.ExclusionList) {
				fmt.Printf("\t   ✅ Skipping analysis for repository '%s' as per ignore list.\n", repoName)
				notAnalyzedCount++ // Increment the counter for repositories analyzed
				continue
			}
		}
		// Next Step : Test is Repository is empty
		isEmpty, err := reposIfEmpty(ctx, client, repoName, parms.Organization)
		if err != nil {
			fmt.Print(err.Error())
			continue

		}
		if !isEmpty {
			ctx := context.Background()
			client := github.NewClient(nil).WithAuthToken(parms.AccessToken)

			totalFiles := 0
			totalLines := 0
			totalBlankLines := 0
			totalComments := 0
			totalCodeLines := 0
			results := make([]map[string]interface{}, 0)
			supportedLanguages := assets.Languages

			languages, _, err := client.Repositories.ListLanguages(ctx, parms.Organization, repoName)
			if err != nil {
				mess := fmt.Sprintf("\r❌ failed to fetch languages. Status code: %v\n", err)
				return 0, 0, 0, 0, fmt.Errorf(mess)
			}

			for lang, lines := range languages {
				if _, ok := supportedLanguages[lang]; ok {
					totalLines += lines / factor
					totalCodeLines += lines / factor
					result := map[string]interface{}{
						"Language":   lang,
						"Files":      1, // Assuming each language file is counted as 1
						"Lines":      lines / factor,
						"BlankLines": 0, // Placeholder for now
						"Comments":   0, // Placeholder for now
						"CodeLines":  lines / factor,
					}
					results = append(results, result)
				}
			}

			output := map[string]interface{}{
				"TotalFiles":      totalFiles,
				"TotalLines":      totalLines,
				"TotalBlankLines": totalBlankLines,
				"TotalComments":   totalComments,
				"TotalCodeLines":  totalCodeLines,
				"Results":         results,
			}

			// Marshal the output to JSON
			jsonData, err := json.MarshalIndent(output, "", "    ")
			if err != nil {
				mess := fmt.Sprintf("\r❌ Error marshaling JSON: %v\n", err)
				return 0, 0, 0, 0, fmt.Errorf(mess)
			}

			// Write JSON data to file
			Resultfile := fmt.Sprintf("Results/Result_%s_%s.json", parms.Organization, repoName)
			file, err := os.Create(Resultfile)
			if err != nil {
				mess := fmt.Sprintf("\r❌ Error creating file: %v\n", err)
				return 0, 0, 0, 0, fmt.Errorf(mess)
			}
			defer file.Close()

			_, err = file.Write(jsonData)
			if err != nil {
				mess := fmt.Sprintf("\r❌ Error writing JSON to file: %v\n", err)
				return 0, 0, 0, 0, fmt.Errorf(mess)
			}

			fmt.Println("\t  ✅  JSON data written to :", Resultfile)

		} else {
			emptyRepo++
		}
	}

	return parms.NBRepos, emptyRepo, notAnalyzedCount, cptarchiv, nil
}

func reposIfEmpty(ctx context.Context, client *github.Client, repoName, org string) (bool, error) {
	// Get the number of commits in the repository
	commits, _, err := client.Repositories.ListCommits(ctx, org, repoName, nil)
	if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
		fmt.Println(MessageApiRate)
		waitTime := rateLimitErr.GetRetryAfter()
		// Sleep until the rate limit resets
		time.Sleep(waitTime)
	}
	if err != nil {
		// If an error occurred, inspect the response body
		var githubError *github.ErrorResponse
		if errors.As(err, &githubError) {
			if githubError.Message == "Git Repository is empty." {
				return true, nil
			} else {
				return true, fmt.Errorf("\n❌ Failed to check repository <%s> is empty - : %v", repoName, err)
			}
		} else {
			return true, fmt.Errorf("\n❌ Failed to check repository <%s> is empty - : %v", repoName, err)
		}
	}

	// Test if the repository is empty
	isEmpty := len(commits) == 0
	if isEmpty {
		return true, nil
	} else {
		return false, nil
	}
}

func sortRepositoriesByUpdatedAt(repos []*github.Repository) {
	sort.Slice(repos, func(i, j int) bool {
		timeI := repos[i].GetUpdatedAt().Time
		timeJ := repos[j].GetUpdatedAt().Time
		return timeI.After(timeJ)
	})
}

func GithubAllBranches(url, AccessToken, apiver string) ([]Branch, error) {

	client := http.Client{}
	var branches []Branch

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", ApiHeader1)
		req.Header.Set("Authorization", "token "+AccessToken)
		req.Header.Set("X-GitHub-Api-Version", apiver)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("\n❌ Failed to list branches. Status code: %d", resp.StatusCode)
		}

		var branchList []Branch
		err = json.NewDecoder(resp.Body).Decode(&branchList)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branchList...)

		nextPageURL := getNextPage(resp.Header)
		if nextPageURL == "" {
			break
		}
		url = nextPageURL
	}

	return branches, nil
}

// manage pagination
func getNextPage(header http.Header) string {
	linkHeader := header.Get("Link")
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) == 2 && strings.Contains(parts[1], `rel="next"`) {
			return strings.Trim(parts[0], "<>")
		}
	}

	return ""
}
