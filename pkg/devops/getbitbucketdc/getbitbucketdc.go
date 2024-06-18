package getbibucketdc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
	"github.com/briandowns/spinner"
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
}

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

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	var message4 string
	emptyRepo := 0
	result := AnalysisResult{}

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = "Get Projects... "
	spin1.Color("green", "bold")

	parms.Spin.Start()
	messageF := fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(projects))
	parms.Spin.FinalMSG = messageF
	parms.Spin.Stop()

	for _, project := range projects {

		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)
		largestRepoSize = 0
		largestRepoBranch = ""

		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos", parms.URL, parms.BaseAPI, parms.APIVersion, project.Key)

		repos, err := fetchAllRepos(urlrepos, parms.AccessToken, exclusionList)
		if err != nil {
			fmt.Println("\r❌ Get Repos for each Project:", err)
			parms.Spin.Stop()
			continue
		}

		nbRepos += len(repos)
		message4 = "Repo(s)"

		fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, len(repos))

		for _, repo := range repos {
			largestRepoSize = 0
			largestRepoBranch = ""
			var branches []Branch
			var Nobranch int = 0

			isEmpty, err := isRepositoryEmpty(project.Key, repo.Slug, parms.AccessToken, bitbucketURLBase, parms.APIVersion)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
				continue
			}

			if !isEmpty {

				if len(parms.Branch) == 0 {

					urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", parms.URL, parms.BaseAPI, parms.APIVersion, project.Key, repo.Slug)

					branches, err = fetchAllBranches(urlrepos, parms.AccessToken)
					if err != nil {
						fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
						parms.Spin.Stop()
						os.Exit(1)
					}
				} else {
					urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?filterText=%s", parms.URL, parms.BaseAPI, parms.APIVersion, project.Key, repo.Slug, parms.Branch)

					branches, err = ifExistBranches(urlrepos, parms.AccessToken)
					if err != nil || len(branches) == 0 {
						fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your <config.json> file  \n", parms.Branch, repo.Slug)
						Nobranch = 1
						continue

					}

				}
				if Nobranch == 0 {

					// Display Number of branches by repo
					fmt.Printf("\n\t   ✅ Repo: <%s> - Number of branches: %d\n", repo.Name, len(branches))

					// Finding the branch with the largest size
					if len(branches) > 1 {
						for _, branch := range branches {
							messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
							spin1.Prefix = messageB
							spin1.Start()

							size, err := fetchBranchSize(project.Key, repo.Slug, branch.Name, parms.AccessToken, parms.URL, parms.APIVersion)
							messageF = ""
							spin1.FinalMSG = messageF

							spin1.Stop()

							if err != nil {
								fmt.Println("\n❌ Error retrieving branch size:", err)
								parms.Spin.Stop()
								os.Exit(1)
							}

							if size > largestRepoSize {
								largestRepoSize = size
								//largestRepoProject = project.Name
								largestRepoBranch = branch.Name
							}

						}

						fmt.Printf("\t     ✅ The largest branch of the repo is <%s> of size : %s\n", largestRepoBranch, utils.FormatSize(int64(largestRepoSize)))
					} else {
						size1, err1 := fetchBranchSize1(project.Key, repo.Slug, parms.AccessToken, parms.URL)
						if err1 != nil {
							fmt.Println("\n❌ Error retrieving branch size:", err1)
							parms.Spin.Stop()
							os.Exit(1)
						}
						largestRepoSize = size1
						largestRepoBranch = branches[0].Name
					}
					importantBranches = append(importantBranches, ProjectBranch{
						ProjectKey:  project.Key,
						RepoSlug:    repo.Slug,
						MainBranch:  largestRepoBranch,
						LargestSize: largestRepoSize,
					})
					Nobranch = 0
				}
			} else {
				emptyRepo++
				Nobranch = 0
			}

		}

	}

	result.NumProjects = len(projects)
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	file, err := os.Create("Results/config/analysis_repos.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, nbRepos, emptyRepo
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos.json> :", err)
		return importantBranches, nbRepos, emptyRepo
	}

	return importantBranches, nbRepos, emptyRepo
}

func GetRepos(project string, repos []Repo, parms ParamsReposDC, bitbucketURLBase string, exclusionList *ExclusionList) ([]ProjectBranch, int, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	var Nobranch int = 0
	var branches []Branch
	emptyRepo := 0
	nbRepos := 1
	result := AnalysisResult{}

	fmt.Printf("\n🟢 Analyse Projet: %s \n", project)

	isEmpty, err := isRepositoryEmpty(project, repos[0].Slug, parms.AccessToken, bitbucketURLBase, parms.APIVersion)
	if err != nil {
		fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repos[0].Name, err)
		parms.Spin.Stop()
		os.Exit(1)
	}
	if !isEmpty {

		if len(parms.Branch) == 0 {

			urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", parms.URL, parms.BaseAPI, parms.APIVersion, project, repos[0].Slug)

			branches, err = fetchAllBranches(urlrepos, parms.AccessToken)
			if err != nil {
				fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repos[0].Name, err)
				parms.Spin.Stop()
				os.Exit(1)
			}

			// Display Number of branches by repo
			fmt.Printf("\n\t   ✅ Repo: <%s> - Number of branches: %d\n", repos[0].Name, len(branches))
		} else {

			urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches?filterText=%s", parms.URL, parms.BaseAPI, parms.APIVersion, project, repos[0].Slug, parms.Branch)

			branches, err = ifExistBranches(urlrepos, parms.AccessToken)

			if err != nil || len(branches) == 0 {
				fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your <config.json> file  \n", parms.Branch, repos[0].Slug)
				Nobranch = 1
				os.Exit(1)

			}
		}

		// Finding the branch with the largest size
		if Nobranch == 0 {
			if len(branches) > 1 {
				for _, branch := range branches {
					messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
					parms.Spin.Prefix = messageB
					parms.Spin.Start()

					size, err := fetchBranchSize(project, repos[0].Slug, branch.Name, parms.AccessToken, parms.URL, parms.APIVersion)
					if err != nil {
						fmt.Println("❌ Error retrieving branch size:", err)
						parms.Spin.Stop()
						continue
					}
					messageF := ""
					parms.Spin.FinalMSG = messageF

					parms.Spin.Stop()

					if size > largestRepoSize {
						largestRepoSize = size
						//largestRepoProject = project.Name
						largestRepoBranch = branch.Name
					}

				}
			} else {
				size1, err1 := fetchBranchSize1(project, repos[0].Slug, parms.AccessToken, parms.URL)
				if err1 != nil {
					fmt.Println("\n❌ Error retrieving branch size:", err1)
					parms.Spin.Stop()
					os.Exit(1)
				}
				largestRepoSize = size1
				largestRepoBranch = branches[0].Name
			}

		}
		Nobranch = 0
		fmt.Printf("\t     ✅ The largest branch of the repo is <%s> of size : %s\n", largestRepoBranch, utils.FormatSize(int64(largestRepoSize)))

		importantBranches = append(importantBranches, ProjectBranch{
			ProjectKey:  project,
			RepoSlug:    repos[0].Slug,
			MainBranch:  largestRepoBranch,
			LargestSize: largestRepoSize,
		})
	} else {
		fmt.Println("❌ Repo is empty:", repos[0].Name)
		return importantBranches, nbRepos, emptyRepo
	}

	result.NumProjects = 1
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	file, err := os.Create("Results/config/analysis_repos_bitbucketdc.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, nbRepos, emptyRepo
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos_bitbucketdc.json> :", err)
		return importantBranches, nbRepos, emptyRepo
	}

	return importantBranches, nbRepos, emptyRepo

}

//func GetProjectBitbucketList(url, baseapi, apiver, accessToken, exlusionfile, project, repo, branchmain string) ([]ProjectBranch, error) {

func GetProjectBitbucketList(platformConfig map[string]interface{}, exlusionfile string) ([]ProjectBranch, error) {

	var largestRepoSize int
	var totalSize int
	var largestRepoProject, largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var exclusionList *ExclusionList
	var err1 error

	totalSize = 0
	nbRepos := 0
	emptyRepo := 0
	bitbucketURLBase := platformConfig["Url"].(string)
	bitbucketURL := fmt.Sprintf("%s%s%s/projects", platformConfig["Url"].(string), platformConfig["Baseapi"].(string), platformConfig["Apiver"].(string))

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")

	// Get All Projects

	if exlusionfile == "0" {
		exclusionList = &ExclusionList{
			Projects: make(map[string]bool),
			Repos:    make(map[string]bool),
		}

	} else {
		exclusionList, err1 = loadExclusionList(exlusionfile)
		if err1 != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exlusionfile, err1)
			return nil, err1
		}

	}

	if len(platformConfig["Project"].(string)) == 0 && len(platformConfig["Repos"].(string)) == 0 {

		spin.Start()
		projects, err := fetchAllProjects(bitbucketURL, platformConfig["AccessToken"].(string), exclusionList)
		if err != nil {
			fmt.Println("\r❌ Error Get All Projects:", err)
			spin.Stop()
			return nil, err
		}

		parms := ParamsReposProjectDC{
			Projects:         projects,
			URL:              platformConfig["Url"].(string),
			BaseAPI:          platformConfig["Baseapi"].(string),
			APIVersion:       platformConfig["Apiver"].(string),
			AccessToken:      platformConfig["AccessToken"].(string),
			BitbucketURLBase: bitbucketURLBase,
			NBRepos:          nbRepos,
			ExclusionList:    exclusionList,
			Spin:             spin,
			Branch:           platformConfig["Branch"].(string),
		}

		importantBranches, nbRepos, emptyRepo = GetReposProject(projects, parms, bitbucketURLBase, nbRepos, exclusionList)

	} else if len(platformConfig["Project"].(string)) > 0 && len(platformConfig["Repos"].(string)) == 0 {
		if isProjectExcluded1(platformConfig["Project"].(string), *exclusionList) {
			fmt.Println("\n❌ Projet", platformConfig["Project"].(string), "is excluded from the analysis... edit <.cloc_bitbucket_ignore> file")
			os.Exit(1)
		} else {

			spin.Start()
			bitbucketURLProject := fmt.Sprintf("%s/%s", bitbucketURL, platformConfig["Project"].(string))

			projects, err := fetchOnelProjects(bitbucketURLProject, platformConfig["AccessToken"].(string), exclusionList)
			if err != nil {
				fmt.Printf("\n❌ Error Get Project:%s - %v", platformConfig["Project"].(string), err)
				spin.Stop()
				return nil, err
			}
			spin.Stop()

			if len(projects) == 0 {
				fmt.Printf("\n❌ Error Project:%s not exist - %v", platformConfig["Project"].(string), err)
				spin.Stop()
				return nil, err
			} else {
				parms := ParamsReposProjectDC{
					Projects:         projects,
					URL:              platformConfig["Url"].(string),
					BaseAPI:          platformConfig["Baseapi"].(string),
					APIVersion:       platformConfig["Apiver"].(string),
					AccessToken:      platformConfig["AccessToken"].(string),
					BitbucketURLBase: bitbucketURLBase,
					NBRepos:          nbRepos,
					ExclusionList:    exclusionList,
					Spin:             spin,
					Branch:           platformConfig["Branch"].(string),
				}

				//importantBranches, nbRepos, emptyRepo = GetReposProject(projects, url, baseapi, apiver, accessToken, bitbucketURLBase, branchmain, nbRepos, exclusionList, spin)
				importantBranches, nbRepos, emptyRepo = GetReposProject(projects, parms, bitbucketURLBase, nbRepos, exclusionList)
			}
		}
	} else if len(platformConfig["Project"].(string)) > 0 && len(platformConfig["Repos"].(string)) > 0 {
		Texclude := platformConfig["Project"].(string) + "/" + platformConfig["Repos"].(string)
		if isProjectAndRepoExcluded(Texclude, *exclusionList) {
			fmt.Println("\n❌ Projet ", platformConfig["Project"].(string), "and the repository ", platformConfig["Repos"].(string), "are excluded from the analysis...edit <.cloc_bitbucket_ignore> file")
			os.Exit(1)
		} else {

			spin.Start()
			bitbucketURLProject := fmt.Sprintf("%s/%s/repos/%s", bitbucketURL, platformConfig["Project"].(string), platformConfig["Repos"].(string))
			Repos, err := fetchOneRepos(bitbucketURLProject, platformConfig["AccessToken"].(string), exclusionList)
			if err != nil {
				fmt.Printf("\n❌ Error Get Repo:%s/%s - %v", platformConfig["Project"].(string), platformConfig["Repos"].(string), err)
				spin.Stop()
				return nil, err
			}
			fmt.Printf("Taille : %d", len(Repos))
			spin.Stop()

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
			}

			importantBranches, nbRepos, emptyRepo = GetRepos(platformConfig["Project"].(string), Repos, parms, bitbucketURLBase, exclusionList)
		}
	} else {
		spin.Stop()
		fmt.Println("❌ Error Project name is empty")
		os.Exit(1)
	}

	largestRepoSize = 0
	largestRepoBranch = ""
	largestRepoProject = ""
	largesRepo = ""

	for _, branch := range importantBranches {

		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largestRepoProject = branch.ProjectKey
			largesRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}
	totalSizeMB := utils.FormatSize(int64(totalSize))
	largestRepoSizeMB := utils.FormatSize(int64(largestRepoSize))

	fmt.Printf("\n✅ The largest repo is <%s> in the project <%s> with the branch <%s> and a size of %s\n", largesRepo, largestRepoProject, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories: %s\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories analyzed: %d - Find empty : %d\n", nbRepos-emptyRepo, emptyRepo)

	return importantBranches, nil
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

	if len(*&project.Key) == 0 {
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

	if len(*&repo.Name) == 0 {
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

func fetchBranchSize1(projectKey string, repoSlug string, accessToken string, bitbucketURLBase string) (int, error) {
	url := fmt.Sprintf("%sprojects/%s/repos/%s/sizes", bitbucketURLBase, projectKey, repoSlug)
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

	var data RepositoryData

	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	totalSize := data.Repository

	return totalSize, nil
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
					fmt.Println("Error:", err)
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
					fmt.Println("Error:", err)
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
