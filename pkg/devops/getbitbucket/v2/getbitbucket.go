package getbibucketv2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ktrysmt/go-bitbucket"
)

type ProjectBranch struct {
	Org         string
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
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

type AnalysisResult struct {
	NumProjects     int
	NumRepositories int
	ProjectBranches []ProjectBranch
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

type ProjectcsResponse struct {
	Values []Projectc `json:"values"`
	Next   string     `json:"next"`
}
type ExclusionList struct {
	Projects map[string]bool
	Repos    map[string]bool
}

type ParamsProjectBitbucket struct {
	Client           *bitbucket.Client
	Projects         []Projectc
	Workspace        string
	URL              string
	BaseAPI          string
	APIVersion       string
	AccessToken      string
	BitbucketURLBase string
	Organization     string
	Exclusionlist    *ExclusionList
	Excludeproject   int
	Spin             *spinner.Spinner
	Period           int
	Stats            bool
	DefaultB         bool
	SingleRepos      string
	SingleBranch     string
}

type Response1 struct {
	Values  []FileInfo `json:"values"`
	Pagelen int        `json:"pagelen"`
	Page    int        `json:"page"`
	Next    string     `json:"next"`
}
type FileInfo struct {
	Path     string `json:"path"`
	Commit   Commit `json:"commit"`
	Type     string `json:"type"`
	Size     int    `json:"size,omitempty"`
	MimeType string `json:"mimetype,omitempty"`
	Links    struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}
type Commit struct {
	Hash  string `json:"hash"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Type string `json:"type"`
}

type Reposize struct {
	Size int `json:"size"`
}

const PrefixMsg = "Get Projects..."

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
			// Exclusion de projet
			exclusionList.Projects[parts[0]] = true
		} else if len(parts) == 2 {
			// Exclusion de répertoire
			exclusionList.Repos[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exclusionList, nil
}

func isRepoExcluded(exclusionList *ExclusionList, projectKey, repoKey string) bool {
	_, repoExcluded := exclusionList.Repos[projectKey+"/"+repoKey]
	return repoExcluded
}

// Fonction pour vérifier si un projet est exclu
func isProjectExcluded(exclusionList *ExclusionList, projectKey string) bool {
	_, projectExcluded := exclusionList.Projects[projectKey]
	return projectExcluded
}

func GetProjectBitbucketListCloud(platformConfig map[string]interface{}, exclusionFile string) ([]ProjectBranch, error) {

	var totalExclude, totalArchiv, emptyRepo, TotalBranches, exludedprojects int
	var nbRepos int

	var largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var projects []Projectc
	var exclusionList *ExclusionList
	var err error
	var totalSize int
	//	result := AnalysisResult{}

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Processing"
	spin.Color("green", "bold")

	exclusionList, err = loadExclusionFileOrCreateNew(exclusionFile)
	if err != nil {
		fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exclusionFile, err)
		spin.Stop()
		return nil, err
	}

	client := bitbucket.NewOAuthbearerToken(platformConfig["AccessToken"].(string))

	project := platformConfig["Project"].(string)
	repos := platformConfig["Repos"].(string)
	bitbucketURLBase := fmt.Sprintf("%s%s/", platformConfig["Url"].(string), platformConfig["Apiver"].(string))

	if len(project) == 0 && len(repos) == 0 {
		// Get All Project
		projects, exludedprojects, err = getAllProjects(client, platformConfig["Workspace"].(string), exclusionList)
		if err != nil {
			fmt.Println("\r❌ Error Get All Projects:", err)
			spin.Stop()
			return nil, err
		}
	} else if len(project) != 0 {
		//else if len(project) != 0 && len(repos) == 0 {
		projects, exludedprojects, err = getSepecificProjects(client, platformConfig["Workspace"].(string), project, exclusionList)
		if err != nil {
			spin.Stop()
			return nil, err
		}
	}
	spin.Stop()

	params := getCommonParams(client, platformConfig, projects, exclusionList, exludedprojects, spin, bitbucketURLBase)
	importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv, err = getRepoAnalyse(params)
	if err != nil {
		spin.Stop()
		return nil, err
	}

	largestRepoBranch, largesRepo = findLargestRepository(importantBranches, &totalSize)

	result := AnalysisResult{
		NumRepositories: nbRepos,
		ProjectBranches: importantBranches,
	}
	if err := SaveResult(result); err != nil {
		fmt.Println("❌ Error Save Result of Analysis :", err)
		os.Exit(1)
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

	printSummary(params.Organization, stats)

	return importantBranches, nil
}

func findLargestRepository(importantBranches []ProjectBranch, totalSize *int) (string, string) {

	var largestRepoBranch, largesRepo string
	largestRepoSize := 0

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug

		}
		*totalSize += branch.LargestSize
	}
	return largestRepoBranch, largesRepo
}

func printSummary(Org string, stats SummaryStats) {
	fmt.Printf("\n✅ The largest Repository is <%s> in the organization <%s> with the branch <%s> \n", stats.LargestRepo, Org, stats.LargestRepoBranch)
	fmt.Printf("\r✅ Total Repositories that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", stats.NbRepos-stats.EmptyRepo-stats.TotalExclude-stats.TotalArchiv, stats.EmptyRepo, stats.TotalExclude, stats.TotalArchiv)
	fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", stats.TotalBranches)
}

func loadExclusionFileOrCreateNew(exclusionFile string) (*ExclusionList, error) {
	if exclusionFile == "0" {
		return &ExclusionList{
			Projects: make(map[string]bool),
			Repos:    make(map[string]bool),
		}, nil
	}
	return loadExclusionList(exclusionFile)
}

func GetSize(parms ParamsProjectBitbucket, repo *bitbucket.Repository) (int, error) {

	url := fmt.Sprintf("%srepositories/%s/%s/?fields=size", parms.BitbucketURLBase, parms.Workspace, repo.Slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+parms.AccessToken)

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

	if strings.Contains(string(body), "error") || strings.Contains(string(body), "size not found") {
		// Branch does not exist, return nil response
		return 0, nil
	}

	var Repostruct Reposize
	err = json.Unmarshal(body, &Repostruct)
	if err != nil {
		return 0, err
	}

	return Repostruct.Size, nil

}

func getCommonParams(client *bitbucket.Client, platformConfig map[string]interface{}, project []Projectc, exclusionList *ExclusionList, excludeproject int, spin *spinner.Spinner, bitbucketURLBase string) ParamsProjectBitbucket {
	return ParamsProjectBitbucket{
		Client:           client,
		Projects:         project,
		Workspace:        platformConfig["Workspace"].(string),
		URL:              platformConfig["Url"].(string),
		BaseAPI:          platformConfig["Baseapi"].(string),
		APIVersion:       platformConfig["Apiver"].(string),
		AccessToken:      platformConfig["AccessToken"].(string),
		BitbucketURLBase: bitbucketURLBase,
		Organization:     platformConfig["Organization"].(string),
		Exclusionlist:    exclusionList,
		Excludeproject:   excludeproject,
		Spin:             spin,
		Period:           int(platformConfig["Period"].(float64)),
		Stats:            platformConfig["Stats"].(bool),
		DefaultB:         platformConfig["DefaultBranch"].(bool),
		SingleRepos:      platformConfig["Repos"].(string),
		SingleBranch:     platformConfig["Branch"].(string),
	}
}

func getAllProjects(client *bitbucket.Client, workspace string, exclusionList *ExclusionList) ([]Projectc, int, error) {

	var projects []Projectc
	var excludedCount int

	projectsRes, err := client.Workspaces.Projects(workspace)
	if err != nil {
		return nil, 0, err
	}

	for _, project := range projectsRes.Items {
		if isProjectExcluded(exclusionList, project.Key) {
			excludedCount++
			continue
		}

		projects = append(projects, Projectc{
			Key:         project.Key,
			UUID:        project.Uuid,
			IsPrivate:   project.Is_private,
			Name:        project.Name,
			Description: project.Description,
		})
	}

	return projects, excludedCount, nil
}

func getSepecificProjects(client *bitbucket.Client, workspace, projectKeys string, exclusionList *ExclusionList) ([]Projectc, int, error) {

	var projects []Projectc
	var excludedCount int

	projectsRes, err := client.Workspaces.GetProject(&bitbucket.ProjectOptions{
		Owner: workspace,
		Key:   projectKeys,
	})
	if err != nil {
		errmessage := fmt.Sprintf("%s - %v", projectKeys, err)
		err1 := fmt.Errorf(errmessage)
		return nil, 0, err1
	}

	if isProjectExcluded(exclusionList, projectsRes.Key) {
		excludedCount++
		errmessage := fmt.Sprintf(" - Skipping analysis for Project %s , it is excluded", projectKeys)
		err = fmt.Errorf(errmessage)
		return projects, excludedCount, err

	} else {

		projects = append(projects, Projectc{
			Key:         projectsRes.Key,
			UUID:        projectsRes.Uuid,
			IsPrivate:   projectsRes.Is_private,
			Name:        projectsRes.Name,
			Description: projectsRes.Description,
		})
	}

	return projects, excludedCount, nil
}

func getRepoAnalyse(params ParamsProjectBitbucket) ([]ProjectBranch, int, int, int, int, int, error) {

	var emptyRepos = 0
	var totalexclude = 0
	var importantBranches []ProjectBranch
	var NBRrepo, TotalBranches int
	var messageF = ""
	NBRrepos := 0
	cptarchiv := 0

	cpt := 1

	message4 := "Repo(s)"

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = PrefixMsg
	spin1.Color("green", "bold")

	params.Spin.Start()
	if params.Excludeproject > 0 {
		messageF = fmt.Sprintf("✅ The number of project(s) to analyze is %d - Excluded : %d\n", len(params.Projects), params.Excludeproject)
	} else {
		messageF = fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(params.Projects))
	}
	params.Spin.FinalMSG = messageF
	params.Spin.Stop()

	// Get Repository in each Project
	for _, project := range params.Projects {

		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)

		emptyOrArchivedCount, excludedCount, repos, err := listReposForProject(params, project.Key)
		if err != nil {
			if len(params.SingleRepos) == 0 {
				fmt.Println("\r❌ Get Repos for each Project:", err)
				spin1.Stop()
				continue
			} else {
				errmessage := fmt.Sprintf(" Get Repo %s for Project %s %v", params.SingleRepos, project.Key, err)
				spin1.Stop()
				return importantBranches, emptyRepos, NBRrepos, TotalBranches, totalexclude, cptarchiv, fmt.Errorf(errmessage)
			}
		}
		emptyRepos = emptyRepos + emptyOrArchivedCount
		totalexclude = totalexclude + excludedCount

		spin1.Stop()
		if emptyOrArchivedCount > 0 {
			NBRrepo = len(repos) + emptyOrArchivedCount
			fmt.Printf("\t  ✅ The number of %s found is: %d - Find empty %d:\n", message4, NBRrepo, emptyOrArchivedCount)
		} else {
			NBRrepo = len(repos)
			fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, NBRrepo)
		}

		for _, repo := range repos {

			largestRepoBranch, repobranches, brsize, err := analyzeRepoBranches(params, repo, cpt, spin1)
			if err != nil {
				largestRepoBranch = repo.Mainbranch.Name

			}

			importantBranches = append(importantBranches, ProjectBranch{
				Org:         params.Organization,
				ProjectKey:  project.Key,
				RepoSlug:    repo.Slug,
				MainBranch:  largestRepoBranch,
				LargestSize: brsize,
			})
			TotalBranches += len(repobranches)

			cpt++
		}

		NBRrepos += NBRrepo

	}

	return importantBranches, emptyRepos, NBRrepos, TotalBranches, totalexclude, cptarchiv, nil

}
func listReposForProject(parms ParamsProjectBitbucket, projectKey string) (int, int, []*bitbucket.Repository, error) {
	var allRepos []*bitbucket.Repository
	var excludedCount, emptyOrArchivedCount int

	page := 1
	for {
		reposRes, err := parms.Client.Repositories.ListProject(&bitbucket.RepositoriesOptions{
			Owner:   parms.Workspace,
			Project: projectKey,
			Page:    &page,
		})
		if err != nil {
			return 0, 0, nil, err
		}

		eoc, exc, repos, err := listRepos(parms, projectKey, reposRes)
		if err != nil {
			return 0, 0, nil, err
		}
		emptyOrArchivedCount += eoc
		excludedCount += exc
		allRepos = append(allRepos, repos...)

		if len(reposRes.Items) < int(reposRes.Pagelen) {
			break
		}

		page++
	}

	return emptyOrArchivedCount, excludedCount, allRepos, nil
}

func listRepos(parms ParamsProjectBitbucket, projectKey string, reposRes *bitbucket.RepositoriesRes) (int, int, []*bitbucket.Repository, error) {
	var allRepos []*bitbucket.Repository
	var excludedCount, emptyOrArchivedCount int

	if len(parms.SingleRepos) == 0 {

		for _, repo := range reposRes.Items {
			repoCopy := repo
			if isRepoExcluded(parms.Exclusionlist, projectKey, repo.Slug) {
				excludedCount++
				continue
			}

			isEmpty, err := isRepositoryEmpty(parms.Workspace, repo.Slug, repo.Mainbranch.Name, parms.AccessToken, parms.BitbucketURLBase)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Slug, err)
			}
			if isEmpty {
				emptyOrArchivedCount++
				continue
			}
			allRepos = append(allRepos, &repoCopy)
		}
	} else {

		var repoFound bool
		for _, repo := range reposRes.Items {

			if repo.Slug == parms.SingleRepos {
				repoFound = true
				repoCopy := repo

				if isRepoExcluded(parms.Exclusionlist, projectKey, repo.Slug) {
					excludedCount++
					errmessage := fmt.Sprintf(" - Skipping analysis for Repo %s , it is excluded", repo.Slug)
					err := fmt.Errorf(errmessage)
					return 0, excludedCount, allRepos, err
				}

				isEmpty, err := isRepositoryEmpty(parms.Workspace, repo.Slug, repo.Mainbranch.Name, parms.AccessToken, parms.BitbucketURLBase)
				if err != nil {
					fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Slug, err)
				}
				if isEmpty {
					emptyOrArchivedCount++
					errmessage := fmt.Sprintf(" - Skipping analysis for Repo %s , it is empty", repo.Slug)
					err := fmt.Errorf(errmessage)
					return emptyOrArchivedCount, excludedCount, allRepos, err
				}

				allRepos = append(allRepos, &repoCopy)
				break
			}
		}

		if !repoFound {
			excludedCount++
		}
	}
	return emptyOrArchivedCount, excludedCount, allRepos, nil
}

// Test is Repository is empty
func isRepositoryEmpty(workspace, repoSlug, mainbranch, accessToken, bitbucketURLBase string) (bool, error) {

	urlMain := fmt.Sprintf("%srepositories/%s/%s/src/%s/?pagelen=100", bitbucketURLBase, workspace, repoSlug, mainbranch)

	filesResp, err := fetchFiles(urlMain, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
	}

	if filesResp == nil {
		urlMaster := fmt.Sprintf("%srepositories/%s/%s/src/master/?pagelen=100", bitbucketURLBase, workspace, repoSlug)
		filesResp, err = fetchFiles(urlMaster, accessToken)
		if err != nil {
			return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
		}
	}

	if len(filesResp.Values) == 0 {
		return true, nil
	}

	return false, nil
}

func fetchFiles(url string, accessToken string) (*Response1, error) {
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

	if strings.Contains(string(body), "error") || strings.Contains(string(body), "Commit not found") {
		// Branch does not exist, return nil response
		return nil, nil
	}

	var filesResp Response1
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return nil, err
	}

	return &filesResp, nil
}

func analyzeRepoBranches(parms ParamsProjectBitbucket, repo *bitbucket.Repository, cpt int, spin1 *spinner.Spinner) (string, []*bitbucket.RepositoryBranch, int, error) {

	var repoBranches []*bitbucket.RepositoryBranch
	var largestRepoBranch string
	var err error
	var brsize, nbrbranche int

	spin1.Prefix = "\r Analyzing branches"
	spin1.Start()

	if parms.DefaultB || len(parms.SingleBranch) != 0 {
		var branchName string
		if parms.DefaultB {
			branchName = repo.Mainbranch.Name
		} else if len(parms.SingleBranch) != 0 {
			branchName = parms.SingleBranch
		}
		repoBranches, largestRepoBranch, brsize, err = getSingleBranches(parms, branchName, repo.Slug, spin1)
		if err != nil {
			spin1.Stop()
			return "", nil, 0, err
		}
		nbrbranche = 1

	} else {
		repoBranches, err := getAllBranches(parms.Client, parms.Workspace, repo.Slug)
		if err != nil {
			spin1.Stop()
			return "", nil, 0, err
		}

		// Determine the largest branch based on the number of commits
		largestRepoBranch, brsize = determineLargestBranch(parms, repo, repoBranches)
		if err != nil {
			spin1.Stop()
			return "", nil, 1, err
		}
		nbrbranche = len(repoBranches)

	}

	spin1.Stop()

	// Print analysis summary
	fmt.Printf("\t\t✅ Repo %d: %s - Number of branches: %d - Largest Branch: %s\n", cpt, repo.Slug, nbrbranche, largestRepoBranch)

	return largestRepoBranch, repoBranches, brsize, nil
}

func getSingleBranches(parms ParamsProjectBitbucket, singlebranch string, repoSlug string, spin1 *spinner.Spinner) ([]*bitbucket.RepositoryBranch, string, int, error) {

	var repoBranches1 []*bitbucket.RepositoryBranch

	branchesRes1, err := parms.Client.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{
		Owner:      parms.Workspace,
		RepoSlug:   repoSlug,
		BranchName: singlebranch,
	})
	if err != nil {
		spin1.Stop()
		return repoBranches1, "", 0, err
	}
	for _, branch := range branchesRes1.Branches {
		branchCopy := branch
		repoBranches1 = append(repoBranches1, &branchCopy)
	}

	return repoBranches1, singlebranch, 1, nil

}

func getAllBranches(client *bitbucket.Client, workspace, repoSlug string) ([]*bitbucket.RepositoryBranch, error) {
	var allBranches []*bitbucket.RepositoryBranch
	options := &bitbucket.RepositoryBranchOptions{
		Owner:    workspace,
		RepoSlug: repoSlug,
		Pagelen:  100,
	}
	page := 1

	for {
		// Set the page number for pagination
		options.PageNum = page

		// Get a page of branches for the repository
		branchesRes, err := client.Repositories.Repository.ListBranches(options)
		if err != nil {
			return nil, err
		}

		// Convert branchesRes.Values to []*bitbucket.RepositoryBranch
		for i := range branchesRes.Branches {
			branch := branchesRes.Branches[i]
			allBranches = append(allBranches, &branch)
		}

		// Check if there are more pages to fetch
		if len(branchesRes.Branches) < options.Pagelen {
			break
		}

		page++
	}

	return allBranches, nil
}
func determineLargestBranch(parms ParamsProjectBitbucket, repo *bitbucket.Repository, branches []*bitbucket.RepositoryBranch) (string, int) {
	var largestRepoBranch string
	var maxCommits, branchSize int

	for _, branch := range branches {
		commits, err := getCommitsForLastMonth(parms.Client, parms.Workspace, repo.Slug, branch.Name, parms.Period)
		if err != nil {
			fmt.Printf("❌ Error getting commits for branch %s: %v\n", branch.Name, err)
			continue
		}
		if len(commits) == 0 {
			branchSize, _ = GetSize(parms, repo)
		} else {
			branchSize = len(commits)
		}

		if branchSize > maxCommits {
			maxCommits = branchSize
			largestRepoBranch = branch.Name
		}
	}

	// If no branch has commits, use the default main branch
	if largestRepoBranch == "" {
		largestRepoBranch = repo.Mainbranch.Name
	}

	return largestRepoBranch, branchSize
}

func getCommitsForLastMonth(client *bitbucket.Client, workspace, repoSlug, branchName string, periode int) ([]interface{}, error) {
	now := time.Now()
	lastMonth := now.AddDate(0, -periode, 0)

	commits, err := client.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{
		Owner:       workspace,
		RepoSlug:    repoSlug,
		Branchortag: branchName,
	})
	if err != nil {
		return nil, err
	}

	var recentCommits []interface{}
	for _, commit := range commits.(map[string]interface{})["values"].([]interface{}) {
		dateStr := commit.(map[string]interface{})["date"].(string)
		commitDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			fmt.Printf("Error parsing commit date: %v\n", err)
			continue
		}

		if commitDate.After(lastMonth) {
			recentCommits = append(recentCommits, commit)
		}
	}

	return recentCommits, nil
}

func SaveResult(result AnalysisResult) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_analysis_result.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the result and write it to the file
	if err := encoder.Encode(result); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_result_bitbucket.json> :", err)
		return err
	}

	fmt.Println("✅ Result saved successfully!")
	return nil
}
