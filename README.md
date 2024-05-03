![Static Badge](https://img.shields.io/badge/Go-v1.22-blue:)

## Introduction

![logo](imgs/Logob.png)

**GoLC** is a clever abbreviation for "Go Line Counter," drawing inspiration from [CLOC](https://github.com/AlDanial/cloc "AlDanial") and various other line-counting tools in Go like [GCloc](https://github.com/JoaoDanielRufino/gcloc "João Daniel Rufino").

**GoLC** counts physical lines of source code in numerous programming languages supported by the Developer, Enterprise, and Data Center editions of [SonarQube](https://www.sonarsource.com/knowledge/languages/) across your Bitbucket Cloud, Bitbucket Data Center, GitHub, GitLab, and Azure DevOps repositories.
GoLC can be used to estimate LoC counts that would be produced by a Sonar analysis of these projects, without having to implement this analysis.

GoLC The tool analyzes your repositories and identifies the largest branch of each repository, counting the total number of lines of code per language for that branch. At the end of the analysis, a text and PDF report is generated, along with a JSON results file for each repository.It starts an HTTP service to display an HTML page with the results.

> This initial version is available for Bitbucket Cloud and Bitbucket DC, and for GitHub, GitLab, Azure DevOps, and Files the next updates will be available soon, integrating these platforms.A Docker version will be planned.

---
## Installation

You can install from the stable release by clicking [here](https://github.com/colussim/GoLC/releases/tag/v1.0.0.0)

## Prerequisites 

* A personal access tokens for : Bitbucket Cloud,Bitbucket DC,GitHub, GitLab and Azure DevOps.The token must have repo scope.
* [Go language installed](https://go.dev/) : If you want to use the sources...

 ## Usage

 ✅ Environment Configuration

 Before running GoLC, you need to configure your environment by initializing the various values in the config.json file.

 ```json
{
    "platforms": {
      "BitBucketSRV": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxx",
        "DevOps": "bitbucket_dc",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "http://X.X.X.X/",
        "Apiver": "1.0",
        "Baseapi": "rest/api/",
        "Protocol": "http",
        "FileExclusion":".cloc_bitbucketdc_ignore"
      },
      "BitBucket": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxx",
        "DevOps": "bitbucket",
        "Workspace":"sonarsource",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "https://api.bitbucket.org/",
        "Apiver": "2.0",
        "Baseapi": "bitbucket.org",
        "Protocol": "http",
        "FileExclusion":".cloc_bitbucket_ignore"
      },
      "Github": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxxxx",
        "DevOps": "github",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "https://api.github.com/",
        "Apiver": "",
        "Baseapi": "api.github.com/",
        "Protocol": "https",
        "FileExclusion":".cloc_github_ignore"
      },
      "Gitlab": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxx",
        "DevOps": "gitlab",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "https://gitlab.com/",
        "Apiver": "v4",
        "Baseapi": "api/",
        "Protocol": "https",
        "FileExclusion":".cloc_gitlab_ignore"
      },
      "Azure": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxx",
        "DevOps": "azure",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "https://dev.azure.com/",
        "Apiver": "7.1",
        "Baseapi": "_apis/git/",
        "Protocol": "https",
        "FileExclusion":".cloc_azure_ignore"
      },
      "File": {
        "Users": "",
        "AccessToken": "",
        "Organization": "xxxxxxxxx",
        "DevOps": "file",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "Url": "",
        "Apiver": "",
        "Baseapi": "",
        "Protocol": "",
        "FileExclusion":".cloc_file_ignore"
      }
    }
  }
  
 ```
This file represents the 6 supported platforms for analysis: BitBucketSRV (Bitbucket DC), BitBucket (cloud), GitHub, GitLab, Azure (Azure DevOps), and File. Depending on your platform, for example, Bitbucket DC (enter BitBucketSRV), specify the parameters:

 ```json
"Users": "xxxxxxxxxxxxxx" : Your User login
"AccessToken": "xxxxxxxxxxxxxx" : Your Token
"Organization": "xxxxxx": Your organization
 ```

If '**Projects**' and '**Repos**' are not specified, the analysis will be conducted on all repositories. You can specify a project name in '**Projects**', and the analysis will be limited to the specified project. If you specify '**Repos**', the analysis will be limited to the specified repositories.
```json
"Project": "",
"Repos": "",
```
For Bitbucket DC, you must provide the URL with your server address and change the '**Protocol**' entry if you are using an https connection , ending with '**/**'. The '**Branch**' entry is not used at the moment.
```json
 "Url": "http://X.X.X.X/"
 ```
You can create a **.cloc_'your_platform'_ignore** file to ignore projects or repositories in the analysis. 
```json
   "FileExclusion":".cloc_bitbucketdc_ignore"
```
The syntax of this file is as follows:

```
REPO_KEY
PROJECT_KEY 
PROJECT_KEY/REPO_KEY
```

```
- REPO_KEY = for one Repository
- PROJECT_KEY = for one Project
- PROJECT_KEY/REPO_KEY For un Repository in one Project
```

 ✅ Run GoLC

 To launch GoLC with the following command, you must specify your DevOps platform. In this example, we analyze repositories hosted on Bitbucket Cloud. The supported flags for -devops are :
 ```bash
flag : <BitBucketSRV>||<BitBucket>||<Github>||<Gitlab>||<Azure>||<File>

 ```
 ❗️ And for now, only the **BitBucketSRV** and **BitBucket** flags are supported...

 If the '**Results**' directory exists, GoLC will prompt you to delete it before starting a new analysis and will also offer to save the previous analysis. If you respond '**y**', a '**Saves**' directory will be created containing a zip file, which will be a compressed version of the '**Results**' directory.

```bash


$:> golc -devops BitBucket

✅ Using configuration for DevOps platform 'BitBucket'

❗️ Directory <'Results'> already exists. Do you want to delete it? (y/n): y
❗️ Do you want to create a backup of the directory before deleting? (y/n): n


🔎 Analysis of devops platform objects ...

✅ The number of project(s) to analyze is 8

        🟢  Analyse Projet: test2 
          ✅ The number of Repositories found is: 1

        🟢  Analyse Projet: tests 
          ✅ The number of Repository found is: 1
        ✅ Repo: testempty - Number of branches: 1

        🟢  Analyse Projet: LSA 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: AdfsTestingTools 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: cloc 
          ✅ The number of Repository found is: 1
        ✅ Repo: gcloc - Number of branches: 2

        🟢  Analyse Projet: sri 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: Bitbucket Pipes 
          ✅ The number of Repository found is: 5
        ✅ Repo: sonarcloud-quality-gate - Number of branches: 9
        ✅ Repo: sonarcloud-scan - Number of branches: 8
        ✅ Repo: official-pipes - Number of branches: 14
        ✅ Repo: sonarqube-scan - Number of branches: 7
        ✅ Repo: sonarqube-quality-gate - Number of branches: 2

        🟢  Analyse Projet: SonarCloud Analysis Samples 
          ✅ The number of Repository found is: 4
        ✅ Repo: sample-maven-project - Number of branches: 6
        ✅ Repo: sample-gradle-project - Number of branches: 3
        ✅ Repo: sample-nodejs-project - Number of branches: 6
        ✅ Repo: sample-dotnet-project-azuredevops - Number of branches: 2

✅ The largest repo is <sample-nodejs-project> in the project <SAMPLES> with the branch <demo-app-week> and a size of 425.45 KB

✅ Total size of your organization's repositories: 877.65 KB
✅ Total repositories analyzed: 11 - Find empty : 1

🔎 Analysis of Repos ...

Extracting files from repo : testempty 
        ✅ json report exported to Results/Result_TES_testempty_main.json
        ✅ The repository <testempty> has been analyzed
                                                                                                    
        ✅ json report exported to Results/Result_CLOC_gcloc_DEV.json
        ✅ The repository <gcloc> has been analyzed
                                                                                              
        ✅ json report exported to Results/Result_BBPIPES_sonarcloud-quality-gate_master.json
        ✅ The repository <sonarcloud-quality-gate> has been analyzed
                                                                                              
        ✅ json report exported to Results/Result_BBPIPES_sonarcloud-scan_master.json
        ✅ The repository <sonarcloud-scan> has been analyzed
         ........

🔎 Analyse Report ...

✅ Number of Repository analyzed in Organization <sonar-demo> is 11 
✅ The repository with the largest line of code is in project <CLOC> the repo name is <gcloc> with <2.05M> lines of code
✅ The total sum of lines of code in Organization <sonar-demo> is : 2.06M Lines of Code


✅ Reports are located in the <'Results'> directory
✅ Time elapsed : 00:01:01

ℹ️  To generate and visualize results on a web interface, follow these steps: 

        ✅ run : ResultsAll
$:>        

```

✅ Run Report

To generate a comprehensive PDF report and view the results on a web interface, you need to launch the '**ResultsAll**' program.

The '**ResultsAll**' program generates a 'GlobalReport.pdf' file in the 'Results' directory. It prompts you if you want to view the results on a web interface; it starts an HTTP service on the default port 8080. If this port is in use, you can choose another port.
To stop the local HTTP service, press the Ctrl+C keys


```bash
$:> ./ResultsAll

✅ Results analysis recorded in Results/code_lines_by_language.json
✅ PDF generated successfully!
Would you like to launch web visualization? (Y/N)
✅ Launching web visualization...
❗️ Port 8080 is already in use.
✅ Please enter the port you wish to use :  9090
✅ Server started on http://localhost:9090
✅ please type < Ctrl+C > to stop the server
$:> 
```

✅  Web UI

![webui](imgs/webui.png)

✅  Report example

![report](imgs/report.png)

