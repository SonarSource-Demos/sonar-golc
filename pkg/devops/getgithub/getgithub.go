package getgithub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const baseURL = "api.github.com"

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"full_name"`
	SizeR         int64  `json:"size"`
}

// Browsing number of pages
func FetchRepositoriesGithub(url string, page int, accessToken string) ([]Repository, string, error) {

	var repos []Repository
	url1 := fmt.Sprintf("%s&page=%d", url, page)

	req, _ := http.NewRequest("GET", url1, nil)
	req.Header.Set("Authorization", ""+accessToken+":")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print("-- Stack: getgithub.FetchRepositories Request API -- ")
		return nil, "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &repos)
	if err != nil {
		fmt.Print("-- Stack: getgithub.FetchRepositories JSON Load -- ")
		return nil, "", err
	}

	nextPageURL := ""
	linkHeader := resp.Header.Get("Link")
	if linkHeader != "" {
		links := strings.Split(linkHeader, ",")
		for _, link := range links {
			parts := strings.Split(strings.TrimSpace(link), ";")
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
				nextPageURL = strings.Trim(parts[0], "<>")
			}
		}
	}

	return repos, nextPageURL, nil
}

// Get Infos for all Repositories in Organization for Main Branch
func GetRepoGithubList(accessToken, organization string) ([]Repository, error) {
	var url = ""
	var repositories []Repository

	url = fmt.Sprintf("https://%s/orgs/%s/repos?type=all&recurse_submodules=false", baseURL, organization)

	page := 1
	for {
		repos, nextPageURL, err := FetchRepositoriesGithub(url, page, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		repositories = append(repositories, repos...)
		if nextPageURL == "" {
			break
		}
		page++
	}

	return repositories, nil
}
