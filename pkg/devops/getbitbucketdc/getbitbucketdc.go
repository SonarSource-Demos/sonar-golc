package getbibucketdc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/emmanuel-colussi-sonarsource/sonar-golc/pkg/utils"
)

type ProjectBranch struct {
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

type RepositoryData struct {
	Repository  int `json:"repository"`
	Attachments int `json:"attachments"`
}

type AnalysisResult struct {
	NumProjects     int
	NumRepositories int
	ProjectBranches []ProjectBranch
}

type ProjectResponse struct {
	Size          int       `json:"size"`
	Limit         int       `json:"limit"`
	IsLastPage    bool      `json:"isLastPage"`
	Values        []Project `json:"values"`
	Start         int       `json:"start"`
	NextPageStart int       `json:"nextPageStart"`
}

type Project struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type Projectc struct {
	Key         string `json:"key"`
	UUID        string `json:"uuid"`
	IsPrivate   bool   `json:"is_private"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Links       struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type RepoResponse struct {
	Size          int    `json:"size"`
	Limit         int    `json:"limit"`
	IsLastPage    bool   `json:"isLastPage"`
	Values        []Repo `json:"values"`
	Start         int    `json:"start"`
	NextPageStart int    `json:"nextPageStart"`
}

type Repo struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Project struct {
		Key string `json:"key"`
	} `json:"project"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type ProjectRepo struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type ParamsReposDC struct {
	Projects         string
	URL              string
	BaseAPI          string
	APIVersion       string
	AccessToken      string
	BitbucketURLBase string
	ExclusionList    *ExclusionList
	Branch           string
	Spin             *spinner.Spinner
	DefaultB         bool
}

type BranchResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Branch `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}
type Branch struct {
	ID              string `json:"id"`
	Name            string `json:"displayId"`
	Type            string `json:"type"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
	IsDefault       bool   `json:"isDefault"`
}

type BranchesResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Branch `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}
type FileResponse struct {
	Path          Path     `json:"path"`
	Revision      string   `json:"revision"`
	Children      Children `json:"children"`
	Start         int      `json:"start"`
	IsLastPage    bool     `json:"isLastPage"`
	NextPageStart int      `json:"nextPageStart"`
}

type Path struct {
	Components []string `json:"components"`
	Name       string   `json:"name"`
	ToString   string   `json:"toString"`
}

type Children struct {
	Size   int    `json:"size"`
	Limit  int    `json:"limit"`
	Values []File `json:"values"`
}

type File struct {
	Path      Path   `json:"path"`
	ContentID string `json:"contentId"`
	Type      string `json:"type"`
	Size      int    `json:"size"`
}

type ExclusionList struct {
	Projects map[string]bool `json:"projects"`
	Repos    map[string]bool `json:"repos"`
}

type ParamsReposProjectDC struct {
	Projects         []Project
	URL              string
	BaseAPI          string
	APIVersion       string
	AccessToken      string
	BitbucketURLBase string
	NBRepos          int
	ExclusionList    *ExclusionList
	Spin             *spinner.Spinner
	Branch           string
	DefaultB         bool
}

var ErrEmptyRepo = errors.New("repository is empty")

func loadExclusionList(filename string) (*ExclusionList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	exclusionList := &ExclusionList{
		Projects: make(map[string]bool),
		Repos:    make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "/")
		if len(parts) == 1 {
			// Get Projet
			exclusionList.Projects[parts[0]] = true
		} else if len(parts) == 2 {
			// Get Repos
			exclusionList.Repos[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exclusionList, nil
}

func GetReposProject(projects []Project, parms ParamsReposProjectDC, bitbucketURLBase string, nbRepos int, exclusionList *ExclusionList) ([]ProjectBranch, int, int) {
	var importantBranches []ProjectBranch
	emptyRepo := 0
	result := AnalysisResult{}

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = "Get Projects... "
	spin1.Color("green", "bold")

	parms.Spin.Start()
	parms.Spin.FinalMSG = fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(projects))
	parms.Spin.Stop()

	for _, project := range projects {
		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)
		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos", parms.URL, parms.BaseAPI, parms.APIVersion, project.Key)

		repos, err := fetchAllRepos(urlrepos, parms.AccessToken, exclusionList)
		if err != nil {
			fmt.Println("\r❌ Get Repos for each Project:", err)
			continue
		}

		nbRepos += len(repos)
		fmt.Printf("\t  ✅ The number of Repo(s) found is: %d\n", len(repos))

		for _, repo := range repos {
			if err := processRepo(project.Key, repo, parms, bitbucketURLBase, spin1, &importantBranches); err != nil {
				if err == ErrEmptyRepo {
					emptyRepo++
				} else {
					fmt.Printf("❌ Error processing repo %s: %v\n", repo.Name, err)
				}
			}
		}
	}

	result.NumProjects = len(projects)
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	if err := saveAnalysisResult1("Results/config/analysis_repos.json", result); err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, nbRepos, emptyRepo
	}

	return importantBranches, nbRepos, emptyRepo
}

func processRepo(projectKey string, repo Repo, parms ParamsReposProjectDC, bitbucketURLBase string, spin1 *spinner.Spinner, importantBranches *[]ProjectBranch) error {
	isEmpty, err := isRepositoryEmpty(projectKey, repo.Slug, parms.AccessToken, bitbucketURLBase, parms.APIVersion)
	if err != nil {
		return fmt.Errorf("testing if repo is empty %s: %w", repo.Name, err)
	}

	if isEmpty {
		fmt.Println("❌ Repo is empty:", repo.Name)
		return ErrEmptyRepo
	}

	branches, err := getBranches1(projectKey, repo, parms)
	if err != nil {
		return err
	}

	if len(branches) == 0 {
		fmt.Printf("❗️ No branches found for repository %s\n", repo.Slug)
		return nil
	}

	fmt.Printf("\n\t   ✅ Repo: <%s> - Number of branches: %d\n", repo.Name, len(branches))

	largestRepoSize, largestRepoBranch, err := findLargestBranch1(projectKey, repo.Slug, branches, parms, spin1)
	if err != nil {
		return err
	}

	fmt.Printf("\t     ✅ The largest branch of the repo is <%s> of size : %s\n", largestRepoBranch, utils.FormatSize(int64(largestRepoSize)))

	*importantBranches = append(*importantBranches, ProjectBranch{
		ProjectKey:  projectKey,
		RepoSlug:    repo.Slug,
		MainBranch:  largestRepoBranch,
		LargestSize: largestRepoSize,
	})

	return nil
}

func getDefaultBranch(url1, accessToken string) (*Branch, error) {
	var allBranches []Branch
	start := 0

	for {

		url := fmt.Sprintf("%s&start=%d", url1, start)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("❌ failed to get branches: %s", resp.Status)
		}

		var branchesRes BranchesResponse
		if err := json.NewDecoder(resp.Body).Decode(&branchesRes); err != nil {
			return nil, err
		}

		allBranches = append(allBranches, branchesRes.Values...)

		if branchesRes.IsLastPage {
			break
		}

		start = branchesRes.NextPageStart
	}

	for _, branch := range allBranches {
		if branch.IsDefault {
			return &branch, nil
		}
	}

	return nil, fmt.Errorf("❌ default branch not found")
}

func GetRepos(project string, repos []Repo, parms ParamsReposDC, bitbucketURLBase string, exclusionList *ExclusionList) ([]ProjectBranch, int, int) {
	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	var branches []Branch
	emptyRepo := 0
	nbRepos := 1
	result := AnalysisResult{}

	fmt.Printf("\n🟢 Analyse Projet: %s \n", project)

	for _, repo := range repos {
		isEmpty, err := isRepositoryEmpty(project, repo.Slug, parms.AccessToken, bitbucketURLBase, parms.APIVersion)
		if err != nil {
			logAndExit(fmt.Sprintf("❌ Error when testing if repo is empty %s: %v\n", repo.Name, err), parms.Spin)
		}

		if isEmpty {
			fmt.Println("❌ Repo is empty:", repo.Name)
			emptyRepo++
			continue
		}

		branches, err = getBranches(project, repo.Slug, parms)
		if err != nil {
			logAndExit(fmt.Sprintf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err), parms.Spin)
		}

		fmt.Printf("\n\t   ✅ Repo: <%s> - Number of branches: %d\n", repo.Name, len(branches))

		if len(branches) == 0 {
			fmt.Printf("❗️ No branches found for repository %s\n", repo.Slug)
			continue
		}

		largestRepoSize, largestRepoBranch, err = findLargestBranch(project, repo.Slug, branches, parms)
		if err != nil {
			logAndExit(fmt.Sprintf("❌ Error retrieving branch size: %v\n", err), parms.Spin)
		}

		fmt.Printf("\t     ✅ The largest branch of the repo is <%s> of size : %s\n", largestRepoBranch, utils.FormatSize(int64(largestRepoSize)))
		importantBranches = append(importantBranches, ProjectBranch{
			ProjectKey:  project,
			RepoSlug:    repo.Slug,
			MainBranch:  largestRepoBranch,
			LargestSize: largestRepoSize,
		})
	}

	result.NumProjects = 1
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	if err := saveAnalysisResult(result); err != nil {
		logAndExit(fmt.Sprintf("❌ Error creating Analysis file: %v\n", err), parms.Spin)
	}

	return importantBranches, nbRepos, emptyRepo
}

func logAndExit(message string, spin *spinner.Spinner) {
	fmt.Println(message)
	if spin != nil {
		spin.Stop()
	}
	os.Exit(1)
}

func getBranches(project, repoSlug string, parms ParamsReposDC) ([]Branch, error) {
	var branches []Branch
	var err error

	if parms.DefaultB {
		urlbr := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?limit=100&start=", parms.URL, parms.BaseAPI, parms.APIVersion, project, repoSlug)
		defaultBranch, err := getDefaultBranch(urlbr, parms.AccessToken)
		if err != nil {
			return nil, err
		}
		branches = append(branches, *defaultBranch)

	} else if len(parms.Branch) == 0 {
		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", parms.URL, parms.BaseAPI, parms.APIVersion, project, repoSlug)
		branches, err = fetchAllBranches(urlrepos, parms.AccessToken)
	} else {
		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?filterText=%s", parms.URL, parms.BaseAPI, parms.APIVersion, project, repoSlug, parms.Branch)
		branches, err = ifExistBranches(urlrepos, parms.AccessToken)
	}
	return branches, err
}

func getBranches1(projectKey string, repo Repo, parms ParamsReposProjectDC) ([]Branch, error) {
	var branches []Branch
	var err error

	if parms.DefaultB {
		urlbr := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?limit=100&start=", parms.URL, parms.BaseAPI, parms.APIVersion, projectKey, repo.Slug)
		defaultBranch, err := getDefaultBranch(urlbr, parms.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("fetching default branch: %w", err)
		}
		branches = append(branches, *defaultBranch)
	} else if len(parms.Branch) == 0 {
		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", parms.URL, parms.BaseAPI, parms.APIVersion, projectKey, repo.Slug)
		branches, err = fetchAllBranches(urlrepos, parms.AccessToken)
	} else {
		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?filterText=%s", parms.URL, parms.BaseAPI, parms.APIVersion, projectKey, repo.Slug, parms.Branch)
		branches, err = ifExistBranches(urlrepos, parms.AccessToken)
		if err != nil || len(branches) == 0 {
			fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your <config.json> file\n", parms.Branch, repo.Slug)
			return nil, nil
		}
	}

	return branches, err
}

func findLargestBranch(project, repoSlug string, branches []Branch, parms ParamsReposDC) (int, string, error) {
	var largestRepoSize int
	var largestRepoBranch string

	for _, branch := range branches {
		parms.Spin.Prefix = fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
		parms.Spin.Start()

		size, err := fetchBranchSize(project, repoSlug, branch.Name, parms.AccessToken, parms.URL, parms.APIVersion)
		parms.Spin.Stop()
		if err != nil {
			fmt.Println("❌ Error retrieving branch size:", err)
			continue
		}

		if size > largestRepoSize {
			largestRepoSize = size
			largestRepoBranch = branch.Name
		}
	}

	return largestRepoSize, largestRepoBranch, nil
}

func findLargestBranch1(projectKey, repoSlug string, branches []Branch, parms ParamsReposProjectDC, spin1 *spinner.Spinner) (int, string, error) {
	var largestRepoSize int
	var largestRepoBranch string

	for _, branch := range branches {
		spin1.Prefix = fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
		spin1.Start()

		size, err := fetchBranchSize(projectKey, repoSlug, branch.Name, parms.AccessToken, parms.URL, parms.APIVersion)
		spin1.Stop()
		if err != nil {
			return 0, "", fmt.Errorf("retrieving branch size: %w", err)
		}

		if size > largestRepoSize {
			largestRepoSize = size
			largestRepoBranch = branch.Name
		}
	}

	return largestRepoSize, largestRepoBranch, nil
}

func saveAnalysisResult(result AnalysisResult) error {
	file, err := os.Create("Results/config/analysis_repos_bitbucketdc.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(result)
}

func saveAnalysisResult1(filePath string, result AnalysisResult) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(result)
}

func GetProjectBitbucketList(platformConfig map[string]interface{}, exclusionFile string) ([]ProjectBranch, error) {
	var importantBranches []ProjectBranch
	var exclusionList *ExclusionList
	var err error
	var nbRepos int

	bitbucketURLBase := platformConfig["Url"].(string)
	bitbucketURL := fmt.Sprintf("%s%s%s/projects", platformConfig["Url"].(string), platformConfig["Baseapi"].(string), platformConfig["Apiver"].(string))

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")

	// Load Exclusion List
	exclusionList, err = loadOrCreateExclusionList(exclusionFile)
	if err != nil {
		fmt.Printf("\n❌ Error Reading Exclusion File <%s>: %v\n", exclusionFile, err)
		return nil, err
	}

	// Determine the Projects and Repos to Analyze
	projects, repos, err := determineProjectsAndRepos(platformConfig, exclusionList, bitbucketURL, spin)
	if err != nil {
		return nil, err
	}
	// Analyze Projects and Repos

	if len(repos) == 0 {
		parms := ParamsReposProjectDC{
			URL:              platformConfig["Url"].(string),
			BaseAPI:          platformConfig["Baseapi"].(string),
			APIVersion:       platformConfig["Apiver"].(string),
			AccessToken:      platformConfig["AccessToken"].(string),
			BitbucketURLBase: bitbucketURLBase,
			ExclusionList:    exclusionList,
			Spin:             spin,
			Branch:           platformConfig["Branch"].(string),
			DefaultB:         platformConfig["DefaultBranch"].(bool),
		}
		importantBranches, nbRepos, _ = GetReposProject(projects, parms, bitbucketURLBase, nbRepos, exclusionList)
	} else {
		parms := ParamsReposDC{
			Projects:         platformConfig["Project"].(string),
			URL:              platformConfig["Url"].(string),
			BaseAPI:          platformConfig["Baseapi"].(string),
			APIVersion:       platformConfig["Apiver"].(string),
			AccessToken:      platformConfig["AccessToken"].(string),
			BitbucketURLBase: bitbucketURLBase,
			ExclusionList:    exclusionList,
			Branch:           platformConfig["Branch"].(string),
			Spin:             spin,
			DefaultB:         platformConfig["DefaultBranch"].(bool),
		}
		importantBranches, nbRepos, _ = GetRepos(platformConfig["Project"].(string), repos, parms, bitbucketURLBase, exclusionList)

	}

	// Summarize Analysis Results
	return summarizeAnalysisResults(importantBranches, nbRepos), nil
}

func loadOrCreateExclusionList(exclusionFile string) (*ExclusionList, error) {
	if exclusionFile == "0" {
		return &ExclusionList{
			Projects: make(map[string]bool),
			Repos:    make(map[string]bool),
		}, nil
	}
	return loadExclusionList(exclusionFile)
}

func determineProjectsAndRepos(platformConfig map[string]interface{}, exclusionList *ExclusionList, bitbucketURL string, spin *spinner.Spinner) ([]Project, []Repo, error) {
	var projects []Project
	var repos []Repo
	var err error

	project := platformConfig["Project"].(string)
	repo := platformConfig["Repos"].(string)

	if project == "" && repo == "" {
		spin.Start()
		projects, err = fetchAllProjects(bitbucketURL, platformConfig["AccessToken"].(string), exclusionList)
		spin.Stop()
	} else if project != "" && repo == "" {
		if isProjectExcluded1(project, *exclusionList) {
			return nil, nil, fmt.Errorf("project %s is excluded from the analysis", project)
		}
		spin.Start()
		projects, err = fetchOnelProjects(fmt.Sprintf("%s/%s", bitbucketURL, project), platformConfig["AccessToken"].(string), exclusionList)
		spin.Stop()
	} else if project != "" && repo != "" {
		Texclude := project + "/" + repo
		if isProjectAndRepoExcluded(Texclude, *exclusionList) {
			return nil, nil, fmt.Errorf("project %s and repository %s are excluded from the analysis", project, repo)
		}
		spin.Start()
		repos, err = fetchOneRepos(fmt.Sprintf("%s/%s/repos/%s", bitbucketURL, project, repo), platformConfig["AccessToken"].(string), exclusionList)
		spin.Stop()
	} else {
		return nil, nil, fmt.Errorf("project name is empty")
	}

	if err != nil {
		return nil, nil, fmt.Errorf("error fetching projects or repos: %v", err)
	}
	return projects, repos, nil
}

func summarizeAnalysisResults(importantBranches []ProjectBranch, nbRepos int) []ProjectBranch {
	var totalSize, largestRepoSize int
	var largestRepoProject, largestRepoBranch, largestRepo string

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largestRepoProject = branch.ProjectKey
			largestRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}

	totalSizeMB := utils.FormatSize(int64(totalSize))
	largestRepoSizeMB := utils.FormatSize(int64(largestRepoSize))

	fmt.Printf("\n✅ The largest repo is <%s> in the project <%s> with the branch <%s> and a size of %s\n", largestRepo, largestRepoProject, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories: %s\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories analyzed: %d - Find empty : %d\n", nbRepos, len(importantBranches)-nbRepos)

	return importantBranches
}

func ifExistBranches(repoURL, accessToken string) ([]Branch, error) {

	req, err := http.NewRequest("GET", repoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var branchesResp BranchResponse
	if resp.StatusCode == http.StatusOK {

		err = json.NewDecoder(resp.Body).Decode(&branchesResp)
		if err != nil {
			return nil, err
		}

	} else {
		var errorResponse struct {
			Type  string `json:"type"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error from API: %s", errorResponse.Error.Message)
	}

	return branchesResp.Values, nil
}

func fetchAllProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Project, error) {
	var allProjects []Project
	for {
		projectsResp, err := fetchProjects(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		projectResponse := projectsResp.(*ProjectResponse)

		for _, project := range projectResponse.Values {

			if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
				allProjects = append(allProjects, project)
			} else {
				if !isProjectExcluded(exclusionList, project.Key) {
					allProjects = append(allProjects, project)
				}
			}
		}

		if projectResponse.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, projectResponse.NextPageStart)
	}
	return allProjects, nil
}

func fetchOnelProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Project, error) {
	var allProjects []Project

	projectsResp, err := fetchProjects(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	project := projectsResp.(*Project)

	if len(project.Key) == 0 {
		fmt.Println("\n❌ Error Project not exist")
		os.Exit(1)
	}
	if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
		allProjects = append(allProjects, *project)
	} else {
		if !isProjectExcluded(exclusionList, project.Key) {
			allProjects = append(allProjects, *project)
		}
	}

	return allProjects, nil
}

func fetchOneRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Repo, error) {
	var allRepos []Repo

	reposResp, err := fetchRepos(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	repo := reposResp.(*Repo)

	if len(repo.Name) == 0 {
		fmt.Println("\n❌ Error Repo or Project not exist")
		os.Exit(1)
	}

	KEYTEST := repo.Project.Key + "/" + repo.Slug

	if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
		allRepos = append(allRepos, *repo)
	} else {
		if !isRepoExcluded(exclusionList, KEYTEST) {
			allRepos = append(allRepos, *repo)
		}
	}

	return allRepos, nil
}

func fetchProjects(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
	var projectsResp interface{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isProjectResponse {
		projectsResp = &ProjectResponse{}
	} else {
		projectsResp = &Project{}
	}

	err = json.Unmarshal(body, &projectsResp)
	if err != nil {
		return nil, err
	}

	return projectsResp, nil

}

func isProjectAndRepoExcluded(repoName string, exclusionList ExclusionList) bool {

	excluded, repoExcluded := exclusionList.Repos[repoName]
	return repoExcluded && excluded
}

func isProjectExcluded1(projectName string, exclusionList ExclusionList) bool {
	_, found := exclusionList.Projects[projectName]
	return found
}
func isProjectExcluded(exclusionList *ExclusionList, project string) bool {
	_, excluded := exclusionList.Projects[project]
	return excluded
}

func isRepoExcluded(exclusionList *ExclusionList, repo string) bool {
	_, excluded := exclusionList.Repos[repo]
	return excluded
}

func fetchAllRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Repo, error) {
	var allRepos []Repo
	for {
		reposResp, err := fetchRepos(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		ReposResponse := reposResp.(*RepoResponse)
		for _, repo := range ReposResponse.Values {
			KEYTEST := repo.Project.Key + "/" + repo.Slug

			if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
				allRepos = append(allRepos, repo)
			} else {
				if !isRepoExcluded(exclusionList, KEYTEST) {
					allRepos = append(allRepos, repo)
				}
			}

		}

		if ReposResponse.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, ReposResponse.NextPageStart)
	}
	return allRepos, nil
}

func fetchRepos(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
	var reposResp interface{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isProjectResponse {
		reposResp = &RepoResponse{}
	} else {
		reposResp = &Repo{}
	}

	err = json.Unmarshal(body, &reposResp)
	if err != nil {
		return nil, err
	}

	return reposResp, nil

}

func fetchAllBranches(url string, accessToken string) ([]Branch, error) {
	var allBranches []Branch
	for {
		branchesResp, err := fetchBranches(url, accessToken)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branchesResp.Values...)
		if branchesResp.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, branchesResp.NextPageStart)
	}
	return allBranches, nil
}

func fetchBranches(url string, accessToken string) (*BranchResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var branchesResp BranchResponse
	err = json.Unmarshal(body, &branchesResp)
	if err != nil {
		return nil, err
	}

	return &branchesResp, nil
}

func fetchFiles(url string, accessToken string) (*FileResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var filesResp FileResponse
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return nil, err
	}

	return &filesResp, nil
}

func fetchBranchSize(projectKey string, repoSlug string, branchName string, accessToken string, bitbucketURLBase string, apiver string) (int, error) {
	url := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse?at=refs/heads/%s", bitbucketURLBase, apiver, projectKey, repoSlug, branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var filesResp FileResponse
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return 0, err
	}

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Children.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Children.Values {
		go func(fileInfo File) {
			defer wg.Done()
			if fileInfo.Type == "FILE" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "DIRECTORY" {
				dirSize, err := fetchDirectorySize(projectKey, repoSlug, branchName, fileInfo.Path.Components, accessToken, bitbucketURLBase, apiver)
				if err != nil {
					fmt.Println("Error fetchDirectorySize:", err)
					return
				}
				resultCh <- dirSize
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for size := range resultCh {
		totalSize += size
	}

	return totalSize, nil
}

func fetchDirectorySize(projectKey string, repoSlug string, branchName string, components []string, accessToken string, bitbucketURLBase string, apiver string) (int, error) {
	url := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse/%s?at=refs/heads/%s", bitbucketURLBase, apiver, projectKey, repoSlug, strings.Join(components, "/"), branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var filesResp FileResponse
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return 0, err
	}

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Children.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Children.Values {
		go func(fileInfo File) {
			defer wg.Done()
			if fileInfo.Type == "FILE" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "DIRECTORY" {
				subdirSize, err := fetchDirectorySize(projectKey, repoSlug, branchName, append(components, fileInfo.Path.Components...), accessToken, bitbucketURLBase, apiver)
				if err != nil {
					//fmt.Println("Error fetchDirectorySize on subdirSize :", err)
					return
				}
				resultCh <- subdirSize
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for size := range resultCh {
		totalSize += size
	}

	return totalSize, nil
}

func isRepositoryEmpty(projectKey, repoSlug, accessToken, bitbucketURLBase, apiver string) (bool, error) {
	urlFiles := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse", bitbucketURLBase, apiver, projectKey, repoSlug)
	filesResp, err := fetchFiles(urlFiles, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo : %s is empty - Function :%s - %v", repoSlug, "getbibucketdc-isRepositoryEmpty", err)
	}
	if filesResp.Children.Size == 0 {
		//fmt.Println("Repo %s is empty", repoSlug)

		return true, nil
	}

	return false, nil
}
