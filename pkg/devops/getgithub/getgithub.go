package getgithub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/SonarSource-Demos/sonar-golc/assets"
	"github.com/briandowns/spinner"
	"github.com/google/go-github/v62/github"
)

type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
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
func loadExclusionRepos(filename string) (ExclusionRepos, error) {
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

	var message4 string       // Message for is one or more Repos
	var TotalBranches int = 0 // Counter Number of Branches on All Repositories

	var largestRepoBranch string
	var importantBranches []ProjectBranch
	//var branches []*github.Branch

	result := AnalysisResult{}               // Result structure
	notAnalyzedCount := 0                    // Counter Number of repositories excluded
	emptyRepo := 0                           // Counter Number of repositories empty
	cpt := 1                                 // Counter Number of Repos
	cptarchiv := 0                           // Counter archiv repos
	opt := &github.ListOptions{PerPage: 100} // Number Object by page in API Request
	opt1 := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	oneMonthAgo := time.Now().AddDate(0, parms.Period, 0)

	parms.Spin.Stop()
	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Color("green", "bold")

	message4 = "Repo(s)"
	fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

	for _, repo := range parms.Repos {
		var branches1 []*github.Branch
		var AllBranches []RepoBranch
		var allEvents []*github.Event

		largestRepoBranch = ""

		repoName := *repo.Name
		repoID := *repo.ID

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
			// Test if we pass branch name as a parameter in the config file
			if len(parms.Branch) == 0 {

				messageB := fmt.Sprintf("\t   Analysis top branch(es) in repository <%s> ...", repoName)
				spin1.Prefix = messageB
				spin1.Start()
				TotalRepoBranches := 0

				/* ----  Get All Branches for Repository ---- */
				for {
					branch0, resp, err := client.Repositories.ListBranches(ctx, parms.Organization, repoName, opt1)

					if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
						fmt.Println(MessageApiRate)
						waitTime := rateLimitErr.GetRetryAfter()
						// Sleep until the rate limit resets
						time.Sleep(waitTime)
					}
					if err != nil {
						fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repoName, err)
						continue
					}
					TotalRepoBranches += len(branch0)
					branches1 = append(branches1, branch0...)
					if resp.NextPage == 0 {
						break
					}
					opt1.Page = resp.NextPage

				}
				TotalBranches += TotalRepoBranches

				/* ----  End Get All Branches for Repository ---- */

				/* ---- Save List of Current Branch ---- */
				AllBranches = append(AllBranches, RepoBranch{
					ID:       repoID,
					Name:     repoName,
					Branches: branches1,
				})
				err = SaveBranch(AllBranches[0])
				if err != nil {
					fmt.Printf("❌ Error saving repositories in file Results/config/analysis_branch_github.json: %v\n", err)
				}
				/* ---- End Save List of Current Branch ---- */

				/* ----  Get List Repository Events ---- */
				for {
					events, resp, err := client.Activity.ListRepositoryEvents(ctx, parms.Organization, repoName, opt)
					if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
						fmt.Println(MessageApiRate)
						waitTime := rateLimitErr.GetRetryAfter()
						// Sleep until the rate limit resets
						time.Sleep(waitTime)
					}
					if err != nil {
						fmt.Println("❌ Error fetching repository events:", err)
						//os.Exit(1)
						continue
					}
					allEvents = append(allEvents, events...)

					if resp.NextPage == 0 {
						break
					}
					opt.Page = resp.NextPage
				}
				/* ----  End Get List tRepository Events ---- */

				/* ---- Count push events for each branch ---- */
				branchPushes := make(map[string]*BranchInfoEvents)
				for _, event := range allEvents {
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
								// Branch references begin with "refs/heads/"
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
				/* ---- End Count push events for each branch ---- */

				/* ---- Get the number of commits and lines of code for selected branches ---- */
				for _, info := range branchPushes {

					// Retrieve depot activity statistics for the last week
					// this blog is deactivated because if the statistics are not activated poses a problem, the analysis is false :
					// job scheduled on GitHub side; try again later
					// Statistics are Commits,Deletes,Adds a last Month
					if parms.Stats {
						contributorsStats, _, err := client.Repositories.ListContributorsStats(ctx, parms.Organization, repoName)
						if rateLimitErr, ok := err.(*github.AbuseRateLimitError); ok {
							fmt.Println(MessageApiRate)
							waitTime := rateLimitErr.GetRetryAfter()
							// Sleep until the rate limit resets
							time.Sleep(waitTime)
						}
						if err != nil {
							//	fmt.Printf("❌ Error fetching contributors stats: %v\n", err)
							continue
						}

						/* ---- Get the number of commits and lines of code for selected branches ---- */
						for _, contributorStats := range contributorsStats {
							// Browse contribution weeks
							for _, week := range contributorStats.Weeks {
								// If the week is included in the last month
								if week.Week.After(oneMonthAgo) {
									// Add line additions and deletions for each contributor to info.Additions and info.Deletions and the number of commits info.Commits
									info.Additions += *week.Additions
									info.Deletions += *week.Deletions
									info.Commits += *week.Commits
								}
							}
						}
						/* ---- End Get the number of commits and lines of code for selected branches ---- */
					} else {
						/* ---- Get the number of commits and lines of code for selected branches ---- */

						// Browse contribution weeks
						opt := &github.CommitsListOptions{
							SHA:         info.Name,
							Since:       oneMonthAgo,
							ListOptions: github.ListOptions{PerPage: 100},
						}
						var allCommits []*github.RepositoryCommit
						for {
							commits, resp, err := client.Repositories.ListCommits(ctx, parms.Organization, repoName, opt)
							if err != nil {
								//fmt.Printf("Error fetching commits for branch %s: %v\n", info.name, err)
								break
							}
							allCommits = append(allCommits, commits...)

							if resp.NextPage == 0 {
								break
							}
							opt.Page = resp.NextPage
						}
						branchPushes[info.Name].Commits = len(allCommits)

					}

				}
				/* ---- End Get the number of commits and lines of code for selected branches ---- */

				/* ---- Sort branches by number of commits, then by additions and deletions of lines of code ---- */
				var branchList []*BranchInfoEvents
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
				/* ---- End Sort branches by number of commits, then by additions and deletions of lines of code ---- */

				// Display the most active and important branch in terms of lines of code
				if len(branchList) > 0 {
					bc := branchList[0]
					largestRepoBranch = bc.Name
					importantBranches = append(importantBranches, ProjectBranch{
						Org:         parms.Organization,
						RepoSlug:    repoName,
						MainBranch:  largestRepoBranch,
						LargestSize: int64(bc.Commits),
					})
					TotalRepoBranches = len(branchPushes)

				} else {
					largestRepoBranch = *repo.DefaultBranch
					importantBranches = append(importantBranches, ProjectBranch{
						Org:         parms.Organization,
						RepoSlug:    repoName,
						MainBranch:  largestRepoBranch,
						LargestSize: int64(*repo.Size),
					})
					TotalRepoBranches = 1
				}

				spin1.Stop()
				////fmt.Println("\r\t\t Repo:", repoName)
				//fmt.Println("\r\t\t Size BrancheIno:", len(branchInfos))
				fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d - largest Branch: %s \n", cpt, repoName, TotalRepoBranches, largestRepoBranch)

			} else {
				largestRepoBranch = parms.Branch
				importantBranches = append(importantBranches, ProjectBranch{
					Org:         parms.Organization,
					RepoSlug:    repoName,
					MainBranch:  largestRepoBranch,
					LargestSize: int64(*repo.Size),
				})
				TotalBranches = 1
			}

		} else {
			emptyRepo++
		}
		cpt++

		//spin1.Start()
	}

	result.NumRepositories = parms.NBRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	err := SaveResult(result)
	if err != nil {
		fmt.Println("❌ Error Save Result of Analysis :", err)
		os.Exit(1)

	}

	return importantBranches, emptyRepo, parms.NBRepos, TotalBranches, notAnalyzedCount, cptarchiv
}

// Get Infos for all Repositories in Organization

func GetRepoGithubList(platformConfig map[string]interface{}, exlusionfile string, fast bool) ([]ProjectBranch, error) {

	var largestRepoSize int64
	var totalSize int64
	var totalExclude int
	var totalArchiv int
	var largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var repositories []*github.Repository
	var exclusionList ExclusionRepos
	var err1 error
	var emptyRepo int
	var TotalBranches int
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
		exclusionList, err1 = loadExclusionRepos(exlusionfile)
		if err1 != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exlusionfile, err1)
			spin.Stop()
			return nil, err1
		}

	}
	// No repo name in config.json file
	if len(platformConfig["Repos"].(string)) == 0 {

		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(platformConfig["AccessToken"].(string))

		// Get all Repositories in Organization
		for {
			repos, resp, err := client.Repositories.ListByOrg(ctx, platformConfig["Organization"].(string), opt)

			if err != nil {
				fmt.Printf("❌ Error fetching repositories: %v\n", err)
				return importantBranches, nil
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

		importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv = GetReposGithub(parms, ctx, client)

	} else {
		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(platformConfig["AccessToken"].(string))

		repos, _, err := client.Repositories.Get(ctx, platformConfig["Organization"].(string), platformConfig["Repos"].(string))
		if err != nil {
			fmt.Printf("❌ Error fetching repository: %v\n", err)
			return importantBranches, nil
		}

		var repositories []*github.Repository

		// Append the repository to the slice
		repositories = append(repositories, repos)

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
		err = SaveRepos(repositories)
		if err != nil {
			fmt.Printf(ErrorMesssage1, err)
		}

		importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv = GetReposGithub(parms, ctx, client)

	}

	largestRepoSize = 0
	largestRepoBranch = ""
	largesRepo = ""

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}

	fmt.Printf("\n✅ The largest Repository is <%s> in the organization <%s> with the branch <%s> \n", largesRepo, platformConfig["Organization"].(string), largestRepoBranch)
	fmt.Printf("\r✅ Total Repositories that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", nbRepos-emptyRepo-totalExclude-totalArchiv, emptyRepo, totalExclude, totalArchiv)
	fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", TotalBranches)

	return importantBranches, nil
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
		exclusionList, err1 = loadExclusionRepos(exlusionfile)
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
		fmt.Printf("\n❌ Error getting commits for repository %s: %v\n", repoName, err)
		return true, fmt.Errorf("\n❌ Failed to check repository <%s> is empty - : %v", repoName, err)
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
