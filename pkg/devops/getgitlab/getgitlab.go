package getgitlab

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
	"github.com/briandowns/spinner"
	"github.com/xanzy/go-gitlab"
)

type ProjectBranch struct {
	Org         string
	Namespace   string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
}

type AnalysisResult struct {
	NumRepositories int
	ProjectBranches []ProjectBranch
}

type AnalyzeProject struct {
	Project       *gitlab.Project
	GitlabClient  *gitlab.Client
	ExclusionList ExclusionRepos
	Spin1         *spinner.Spinner
	Org           string
}

// RepositoryMap represents a map of repositories to ignore
type ExclusionRepos map[string]bool

const PrefixMsg = "Get Project(s)..."
const MessageErro1 = "/\n❌ Failed to list projects for group %s: %v\n"
const MessageError2 = "\n ❌ Failed to get project %s: %v\n"
const MessageError3 = "\n❗️ Project %s is in exclude file \n"
const MessageError4 = "\n❗️ Project %s is empty \n"
const MessageError5 = "\n❗️ Project %s is archived \n"
const MessageError6 = "\n❗️ Project %s is in exclude file or empty or archived \n"
const Message1 = "\t ✅ The number of %s found is: %d\n"
const Message2 = "\t   Analysis top branch(es) in project <%s> ..."
const Message3 = "\r\t\t\t\t ✅ %d Project: %s - Number of branches: %d - largest Branch: %s"
const Message4 = "Project(s)"

const (
	perPage = 100
)

// Load repository ignore map from file
func LoadExclusionRepos(filename string) (ExclusionRepos, error) {

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

// Function to Get Commit count
func getCommitCount(client *gitlab.Client, projectID int, branchName string, since, until time.Time) (int, error) {
	commits, _, err := client.Commits.ListCommits(projectID, &gitlab.ListCommitsOptions{
		RefName: &branchName,
		Since:   &since,
		Until:   &until,
	})
	if err != nil {
		return 0, err
	}
	return len(commits), nil
}

// Function to Retrieves all projects in the main group as well as those in its subgroups.
func getAllGroupProjects(client *gitlab.Client, groupName string) ([]*gitlab.Project, error) {
	group, err := getGroup(client, groupName)
	if err != nil {
		return nil, err
	}

	projects, err := getProjectsInGroup(client, group)
	if err != nil {
		return nil, err
	}

	subgroups, err := getSubgroups(client, group)
	if err != nil {
		return nil, err
	}

	for _, subgroup := range subgroups {
		subgroupProjects, err := getProjectsInGroup(client, subgroup)
		if err != nil {
			return nil, err
		}
		projects = append(projects, subgroupProjects...)
	}

	return projects, nil
}

// Function to Retrieves information about the primary group.
func getGroup(client *gitlab.Client, groupName string) (*gitlab.Group, error) {
	group, _, err := client.Groups.GetGroup(groupName, nil)
	return group, err
}

// Function to Retrieves the list of subgroups of a given group.
func getSubgroups(client *gitlab.Client, group *gitlab.Group) ([]*gitlab.Group, error) {
	subgroups, _, err := client.Groups.ListSubGroups(group.ID, nil)
	return subgroups, err
}

// Function to Retrieves all projects in a given group.
func getProjectsInGroup(client *gitlab.Client, group *gitlab.Group) ([]*gitlab.Project, error) {
	var projects []*gitlab.Project
	page := 1
	for {
		opts := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		projs, resp, err := client.Groups.ListGroupProjects(group.ID, opts)
		if err != nil {
			return nil, err
		}

		projects = append(projects, projs...)

		// Vérifier si c'est la dernière page
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Passer à la page suivante
		page++
	}
	return projects, nil
}

// Function to Get Branch Size
func getBranchSize(client *gitlab.Client, projectID int, branchName string) int {
	tree, resp, err := client.Repositories.ListTree(projectID, &gitlab.ListTreeOptions{
		Ref: &branchName,
	})
	if err != nil && resp.StatusCode != 404 {
		if resp != nil {
			return 2
		} else {
			//if resp != nil && resp.StatusCode == 404 {
			//if resp.StatusCode == 404 {
			return 1
		}
		//return 0
	}
	return len(tree)
}

// Function test if branch exist
func branchExists(gitlabClient *gitlab.Client, projectID interface{}, branchName string) bool {
	_, _, err := gitlabClient.Branches.GetBranch(projectID, branchName)
	return err == nil
}

// Function to Get Most important Branch
func getMainBranch(client *gitlab.Client, projectID int, since, until time.Time) (string, int, int, error) {

	branches := make([]*gitlab.Branch, 0)
	page := 1
	perPage := 100

	// List of project branches
	for {

		opts := &gitlab.ListBranchesOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		brs, resp, err := client.Branches.ListBranches(projectID, opts)
		if err != nil {
			return "", 0, 0, err
		}
		branches = append(branches, brs...)
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		page++
	}

	var largestBranch string
	largestSize := 0

	for _, branch := range branches {
		commitCount, err := getCommitCount(client, projectID, branch.Name, since, until)
		if err != nil {
			return "", 0, 0, err
		}
		if commitCount > largestSize {
			largestSize = commitCount
			largestBranch = branch.Name
		}
	}

	// If no commits are found for all branches, use the default branch
	if largestSize == 0 {

		project, _, err := client.Projects.GetProject(projectID, nil)
		if err != nil {
			return "", 0, len(branches), err
		}
		largestBranch = project.DefaultBranch
		largestSize = getBranchSize(client, project.ID, largestBranch)
		return largestBranch, largestSize, len(branches), nil

	}
	//fmt.Println("Branche", len(branches))
	return largestBranch, largestSize, len(branches), nil
}

// Function to check if a project should be excluded from analysis
func isExcluded(projectName string, exclusionList map[string]bool) bool {

	if _, ok := exclusionList[projectName]; ok {
		return true
	}

	// Check for subdomain match
	for excludedRepo := range exclusionList {
		if strings.HasPrefix(projectName, excludedRepo) {
			return true
		}
	}

	return false

}

func SaveResult(result AnalysisResult) error {

	loggers := utils.NewLogger()
	// Open or create the file
	file, err := os.Create("Results/config/analysis_result_github.json")
	if err != nil {

		loggers.Errorf("❌ Error creating Analysis file:%v", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the result and write it to the file
	if err := encoder.Encode(result); err != nil {
		loggers.Errorf("❌ Error encoding JSON file <Results/config/analysis_result_github.json> :%v", err)
		return err
	}

	fmt.Print("\n")
	loggers.Info("✅ Result saved successfully!")
	return nil
}

func analyzeProj(analyzeProject AnalyzeProject) (ProjectBranch, int, int, int) {

	largestSize := 0

	if isExcluded(analyzeProject.Project.PathWithNamespace, analyzeProject.ExclusionList) {
		return ProjectBranch{}, 1, 0, 0
	}

	// Check if the project is empty or archived
	if analyzeProject.Project.EmptyRepo || analyzeProject.Project.Archived {
		if analyzeProject.Project.EmptyRepo {
			return ProjectBranch{}, 0, 1, 0
		}
		if analyzeProject.Project.Archived {
			return ProjectBranch{}, 0, 0, 1
		}
	}

	messageB := fmt.Sprintf(Message2, analyzeProject.Project.Name)
	analyzeProject.Spin1.Prefix = messageB
	analyzeProject.Spin1.Start()
	// Retrieve project branches

	largestBranch := analyzeProject.Project.DefaultBranch
	largestSize = getBranchSize(analyzeProject.GitlabClient, analyzeProject.Project.ID, largestBranch)

	projectBranches := ProjectBranch{
		Org:         analyzeProject.Org,
		Namespace:   analyzeProject.Project.PathWithNamespace,
		RepoSlug:    analyzeProject.Project.Name,
		MainBranch:  largestBranch,
		LargestSize: largestSize,
	}

	return projectBranches, 0, 0, 0

}

func processProject(analyzeProject AnalyzeProject, cpt int, spin1 *spinner.Spinner, projectBranches []ProjectBranch, emptyRepos, archivedRepos, excludedProjects *int) ([]ProjectBranch, int) {
	projectBranche, ExcludedProject, EmptyRepos, ArchivedRepos := analyzeProj(analyzeProject)

	loggers := utils.NewLogger()

	if EmptyRepos > 0 {
		(*emptyRepos)++
		return projectBranches, cpt
	}
	if ArchivedRepos > 0 {
		(*archivedRepos)++
		return projectBranches, cpt
	}
	if ExcludedProject > 0 {
		(*excludedProjects)++
		return projectBranches, cpt
	}

	projectBranches = append(projectBranches, projectBranche)
	spin1.Stop()
	loggers.Infof(Message3, cpt, analyzeProject.Project.Name, 1, projectBranche.MainBranch)
	cpt++
	return projectBranches, cpt
}

func isProjectExcludedOrInvalid(project *gitlab.Project, exclusionList ExclusionRepos, emptyRepos, archivedRepos *int) (bool, bool, bool) {
	if isExcluded(project.PathWithNamespace, exclusionList) {
		return true, false, false
	}

	if project.EmptyRepo {
		*emptyRepos++
		return false, true, false
	}

	if project.Archived {
		*archivedRepos++
		return false, false, true
	}

	return false, false, false
}

func getMainBranchDetails(gitlabClient *gitlab.Client, project *gitlab.Project, since, until time.Time) (string, int, int, error) {
	mainBranch, largestSize, nbrsize, err := getMainBranch(gitlabClient, project.ID, since, until)
	if err != nil {
		return "", 0, 0, fmt.Errorf("❌ Failed to get main branch for project %s: %v\n", project.Name, err)
	}
	return mainBranch, largestSize, nbrsize, nil
}

func GetRepoGitLabList(platformConfig map[string]interface{}, exclusionfile string) ([]ProjectBranch, error) {

	var projectBranches []ProjectBranch
	var emptyRepos, archivedRepos int
	var TotalBranches int = 0 // Counter Number of Branches on All Repositories
	var exclusionList ExclusionRepos
	var err1 error
	var totalSize int
	loggers := utils.NewLogger()

	excludedProjects := 0
	result := AnalysisResult{}

	// Calculating the period
	until := time.Now()
	since := until.AddDate(0, int(platformConfig["Period"].(float64)), 0)
	ApiURL := platformConfig["Url"].(string) + platformConfig["Baseapi"].(string) + platformConfig["Apiver"].(string)

	loggers.Infof("🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")
	spin.Start()

	// Test if exclusion file exist
	if exclusionfile == "0" {
		exclusionList = make(map[string]bool)

	} else {

		exclusionList, err1 = LoadExclusionRepos(exclusionfile)
		if err1 != nil {
			loggers.Errorf("❌ Error Read Exclusion File <%s>: %v", exclusionfile, err1)
			spin.Stop()
			return nil, err1
		}

	}

	gitlabClient, err := gitlab.NewClient(platformConfig["AccessToken"].(string), gitlab.WithBaseURL(ApiURL))
	if err != nil {
		loggers.Fatalf("❌ Failed to create client: %v", err)
	}

	/* --------------------- Analysis a default branche  ---------------------  */
	if platformConfig["DefaultBranch"].(bool) {
		cpt := 1
		//switch {

		/* --------------------- Analysis all projects with a default branche  ---------------------  */
		if platformConfig["Project"].(string) == "" {

			projects, err := getAllGroupProjects(gitlabClient, platformConfig["Organization"].(string))
			if err != nil {
				loggers.Fatalf(MessageErro1, platformConfig["Organization"].(string), err)
			}

			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")

			loggers.Infof(Message1, Message4, len(projects))

			for _, project := range projects {
				//largestSize := 0

				parmsproject := AnalyzeProject{

					Project:       project,
					GitlabClient:  gitlabClient,
					ExclusionList: exclusionList,
					Spin1:         spin1,
					Org:           platformConfig["Organization"].(string),
				}

				projectBranches, cpt = processProject(parmsproject, cpt, spin1, projectBranches, &emptyRepos, &archivedRepos, &excludedProjects)
				TotalBranches++
			}
			/* --------------------- End Analysis all projects with a default branche  ---------------------  */
		} else {
			/* --------------------- Analysis a specific projects with a default branche  ---------------------  */
			cpt := 1
			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")
			//	largestSize := 0

			namespase := platformConfig["Organization"].(string) + "/" + platformConfig["Project"].(string)

			project, _, err := gitlabClient.Projects.GetProject(namespase, nil)
			if err != nil {
				loggers.Fatalf(MessageError2, platformConfig["Project"].(string), err)
			}

			parmsproject := AnalyzeProject{

				Project:       project,
				GitlabClient:  gitlabClient,
				ExclusionList: exclusionList,
				Spin1:         spin1,
				Org:           platformConfig["Organization"].(string),
			}

			projectBranches, _ = processProject(parmsproject, cpt, spin1, projectBranches, &emptyRepos, &archivedRepos, &excludedProjects)
			TotalBranches++

		}
		/* --------------------- End Analysis a specific projects with a default branche  ---------------------  */

		/* --------------------- End Analysis a default branche  ---------------------  */

	} else {

		/* --------------------- Analysis all Project and All Branches if not if you do not specify a specific project or branch ---------------------  */
		switch {
		case platformConfig["Project"].(string) == "" && platformConfig["Branch"].(string) == "":
			/*cpt := 1

			projects, err := getAllGroupProjects(gitlabClient, platformConfig["Organization"].(string))
			if err != nil {
				spin.Stop()
				log.Fatalf(MessageErro1, platformConfig["Organization"].(string), err)
			}

			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")

			fmt.Printf(Message1, Message4, len(projects))*/

			projects, cpt, spin1, err := getProjectsAndAnalyze(gitlabClient, platformConfig["Organization"].(string), spin)
			if err != nil {
				loggers.Fatalf(err.Error())
			}

			for _, project := range projects {
				//	branches := make([]*gitlab.Branch, 0)
				TotalRepoBranches := 0

				excluded, empty, archived := isProjectExcludedOrInvalid(project, exclusionList, &emptyRepos, &archivedRepos)
				if excluded {
					excludedProjects++
					continue
				}
				if empty || archived {
					continue
				}

				messageB := fmt.Sprintf(Message2, project.Name)
				spin1.Prefix = messageB
				spin1.Start()

				mainBranch, largestSize, nbrsize, err := getMainBranchDetails(gitlabClient, project, since, until)
				if err != nil {
					spin1.Stop()
					loggers.Fatalf(err.Error())
				}

				projectBranches = append(projectBranches, ProjectBranch{
					Org:         platformConfig["Organization"].(string),
					Namespace:   project.PathWithNamespace,
					RepoSlug:    project.Name,
					MainBranch:  mainBranch,
					LargestSize: largestSize,
				})
				TotalRepoBranches = nbrsize

				spin1.Stop()
				loggers.Infof(Message3, cpt, project.Name, nbrsize, mainBranch)
				cpt++
				TotalBranches += TotalRepoBranches
			}
		// Repeat similar extraction for other cases

		/* ---------------------End Analysis all Project and All Branches if not if you do not specify a specific project or branch ---------------------  */

		/* --------------------- Analysis a specific Project and All Branches  ---------------------  */

		case platformConfig["Project"].(string) != "" && platformConfig["Branch"].(string) == "":

			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")

			namespase := platformConfig["Organization"].(string) + "/" + platformConfig["Project"].(string)

			project, _, err := gitlabClient.Projects.GetProject(namespase, nil)
			if err != nil {
				loggers.Fatalf(MessageError2, platformConfig["Project"].(string), err)
			}

			excluded, empty, archived := isProjectExcludedOrInvalid(project, exclusionList, &emptyRepos, &archivedRepos)
			if excluded || empty || archived {
				loggers.Fatalf(MessageError6, platformConfig["Project"].(string))

			}

			messageB := fmt.Sprintf("\t   Analysis top branch(es) in repository <%s> ...", project.Name)
			spin1.Prefix = messageB
			spin1.Start()

			mainBranch, largestSize, nbrsize, err := getMainBranchDetails(gitlabClient, project, since, until)
			if err != nil {
				loggers.Fatalf("\n ❌ Failed to get main branch for project %s: %v\n", platformConfig["Project"].(string), err)
			}

			spin1.Stop()
			loggers.Infof("\r\t\t✅ 1 Project: %s - Number of branches: %d - largest Branch: %s", project.Name, nbrsize, mainBranch)

			projectBranches = append(projectBranches, ProjectBranch{
				Org:         platformConfig["Organization"].(string),
				Namespace:   project.PathWithNamespace,
				RepoSlug:    platformConfig["Project"].(string),
				MainBranch:  mainBranch,
				LargestSize: largestSize,
			})

			TotalBranches = nbrsize

		/* --------------------- End Analysis a specific Project and All Branches  ---------------------  */

		/* --------------------- Analysis a specific Project with a Branche  ---------------------  */

		case platformConfig["Project"].(string) != "" && platformConfig["Organization"].(string) != "":

			namespase := platformConfig["Organization"].(string) + "/" + platformConfig["Project"].(string)

			project, _, err := gitlabClient.Projects.GetProject(namespase, nil)
			if err != nil {
				loggers.Fatalf(MessageError2, platformConfig["Project"].(string), err)
			}
			if isExcluded(project.PathWithNamespace, exclusionList) {
				//excludedProjects++
				loggers.Fatalf(MessageError3, platformConfig["Project"].(string))

			}
			// Check if the project is empty or archived
			if project.EmptyRepo || project.Archived {
				if project.EmptyRepo {
					loggers.Fatalf(MessageError4, platformConfig["Project"].(string))
				}
				if project.Archived {
					loggers.Fatalf(MessageError5, platformConfig["Project"].(string))
				}
			}

			projectBranches = append(projectBranches, ProjectBranch{
				Org:         platformConfig["Organization"].(string),
				Namespace:   project.PathWithNamespace,
				RepoSlug:    platformConfig["Project"].(string),
				MainBranch:  platformConfig["Branch"].(string),
				LargestSize: 1,
			})
			TotalBranches = 1
		/* --------------------- End Analysis a specific Project with a Branche  ---------------------  */

		/* --------------------- Analysis all Project with a specific Branche  ---------------------  */
		case platformConfig["Project"].(string) == "" && platformConfig["Branch"].(string) != "":

			/*cpt := 1

			projects, err := getAllGroupProjects(gitlabClient, platformConfig["Organization"].(string))
			if err != nil {
				spin.Stop()
				log.Fatalf(MessageErro1, platformConfig["Organization"].(string), err)
			}

			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")

			fmt.Printf(Message1, Message4, len(projects))*/

			projects, cpt, spin1, err := getProjectsAndAnalyze(gitlabClient, platformConfig["Organization"].(string), spin)
			if err != nil {
				loggers.Fatalf(err.Error())
			}

			for _, project := range projects {
				//largestSize := 0

				excluded, empty, archived := isProjectExcludedOrInvalid(project, exclusionList, &emptyRepos, &archivedRepos)
				if excluded {
					excludedProjects++
					continue
				}
				if empty || archived {
					continue
				}
				messageB := fmt.Sprintf(Message2, project.Name)
				spin1.Prefix = messageB
				spin1.Start()

				largestBranch := platformConfig["Branch"].(string)
				if !branchExists(gitlabClient, project.ID, largestBranch) {
					spin1.Stop()
					continue
				}

				largestSize := getBranchSize(gitlabClient, project.ID, largestBranch)

				projectBranches = append(projectBranches, ProjectBranch{
					Org:         platformConfig["Organization"].(string),
					Namespace:   project.PathWithNamespace,
					RepoSlug:    project.Name,
					MainBranch:  largestBranch,
					LargestSize: largestSize,
				})

				spin1.Stop()
				loggers.Infof(Message3, cpt, project.Name, 1, largestBranch)
				cpt++
				TotalBranches++
			}

		}
		/* --------------------- End Analysis all Project with a specific Branche  ---------------------  */
	}
	largestRepoSize := 0
	largestRepoBranch := ""
	largesRepo := ""

	for _, branch := range projectBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}

	result.NumRepositories = len(projectBranches)
	result.ProjectBranches = projectBranches
	// Save Result of Analysis
	err = SaveResult(result)
	if err != nil {
		loggers.Errorf("❌ Error Save Result of Analysis :%v", err)
		os.Exit(1)

	}

	//fmt.Printf("\n✅ The largest Repository is <%s> in the Organizationa <%s> with the branch <%s> \n", largesRepo, platformConfig["Organization"].(string), largestRepoBranch)
	//fmt.Printf("\r✅ TotalProject(s) that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", len(projectBranches), emptyRepos, excludedProjects, archivedRepos)
	//fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", TotalBranches)

	fmt.Print("\n")
	loggers.Infof("✅ The largest Repository is <%s> in the Organizationa <%s> with the branch <%s>", largesRepo, platformConfig["Organization"].(string), largestRepoBranch)
	loggers.Infof("✅ TotalProject(s) that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d", len(projectBranches), emptyRepos, excludedProjects, archivedRepos)
	loggers.Infof("✅ Total Branches that will be analyzed: %d\n", TotalBranches)
	return projectBranches, nil
}

func getProjectsAndAnalyze(gitlabClient *gitlab.Client, organization string, spin *spinner.Spinner) ([]*gitlab.Project, int, *spinner.Spinner, error) {

	cpt := 1
	loggers := utils.NewLogger()

	projects, err := getAllGroupProjects(gitlabClient, organization)
	if err != nil {
		spin.Stop()
		log.Fatalf(MessageErro1, organization, err)
	}

	spin.Stop()
	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Color("green", "bold")

	loggers.Infof(Message1, Message4, len(projects))

	return projects, cpt, spin1, nil
}
