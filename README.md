![Static Badge](https://img.shields.io/badge/Go-v1.22-blue:)


## Introduction

![logo](imgs/Logob.png)

**GoLC** is a clever abbreviation for "Go Line Counter," drawing inspiration from [CLOC](https://github.com/AlDanial/cloc "AlDanial") and various other line-counting tools in Go like [GCloc](https://github.com/JoaoDanielRufino/gcloc "João Daniel Rufino").

**GoLC** counts physical lines of source code in numerous programming languages supported by the Developer, Enterprise, and Data Center editions of [SonarQube](https://www.sonarsource.com/knowledge/languages/) across your Bitbucket Cloud, Bitbucket Data Center, GitHub, GitLab, and Azure DevOps repositories.
GoLC can be used to estimate LoC counts that would be produced by a Sonar analysis of these projects, without having to implement this analysis.

GoLC The tool analyzes your repositories and identifies the largest branch of each repository, counting the total number of lines of code per language for that branch. At the end of the analysis, a text and PDF report is generated, along with a JSON results file for each repository.It starts an HTTP service to display an HTML page with the results.

> This version ver1.0.2 is available for Bitbucket Cloud , Bitbucket DC, GitHub , GitLab and Files, for Azure DevOps the next updates will be available soon, integrating these platforms.A Docker version will be planned.

---
## Installation

You can install from the stable release by clicking [here](https://github.com/emmanuel-colussi-sonarsource/sonar-golc/releases/tag/V1.0.2)

## Prerequisites 

* A personal access tokens for : Bitbucket Cloud,Bitbucket DC,GitHub, GitLab and Azure DevOps.The token must have repo scope.
     - Perform pull request actions
     - Push, pull and clone repositories
  
* [Go language installed](https://go.dev/) : If you want to use the sources...


## Supported languages

To show all supported languages use the subcommand languages :

 ```
$:> golc.go -languages

Language           | Extensions                               | Single Comments | Multi Line
                    |                                          |                 | Comments
-------------------+------------------------------------------+-----------------+--------------
Objective-C        | .m                                       | //              | /* */ 
Ruby               | .rb                                      | #               | =begin =end 
Visual Basic .NET  | .vb                                      | '               | 
YAML               | .yaml, .yml                              | #               | 
C#                 | .cs                                      | //              | /* */ 
Flex               | .as                                      | //              | /* */ 
C++ Header         | .hh, .hpp                                | //              | /* */ 
CSS                | .css                                     | //              | /* */ 
Abap               | .abap, .ab4, .flow                       | "               | /* */ 
PL/I               | .pl1                                     | --              | /* */ 
RPG                | .rpg                                     | #               | 
Swift              | .swift                                   | //              | /* */ 
JCL                | .jcl, .JCL                               | //              | /* */ 
Apex               | .cls, .trigger                           | //              | /* */ 
PHP                | .php, .php3, .php4, .php5, .phtml, .inc  | //, #           | /* */ 
TypeScript         | .ts, .tsx                                | //              | /* */ 
XML                | .xml, .XML                               | <!--            | <!-- --> 
XHTML              | .xhtml                                   | <!--            | <!-- --> 
Terraform          | .tf                                      |                 | 
T-SQL              | .tsql                                    | --              | 
Vue                | .vue                                     | <!--            | <!-- --> 
COBOL              | .cbl, .ccp, .cob, .cobol, .cpy           | *, /            | 
HTML               | .html, .htm, .cshtml, .vbhtml, .aspx,    |                 | <!-- --> 
                    | .ascx, .rhtml, .erb, .shtml, .shtm, cmp  |                 | <!-- -->
JavaScript         | .js, .jsx, .jsp, .jspf                   | //              | /* */ 
Python             | .py                                      | #               | """ """ 
Scss               | .scss                                    | //              | /* */ 
SQL                | .sql                                     | --              | /* */ 
C Header           | .h                                       | //              | /* */ 
C++                | .cpp, .cc                                | //              | /* */ 
Golang             | .go                                      | //              | /* */ 
Oracle PL/SQL      | .pkb                                     | --              | /* */ 
ActionScript       | .as                                      | //              | /* */ 
C                  | .c                                       | //              | /* */ 
Java               | .java, .jav                              | //              | /* */ 
Kotlin             | .kt, .kts                                | //              | /* */ 
Scala              | .scala                                   | //              | /* */ 

 ```

 ❗️ To add a new language, you need to add an entry to the Languages structure defined in the file [assets/languages.go](assets/languages.go).


 ## Usage

 ✅ Environment Configuration

 Before running GoLC, you need to configure your environment by initializing the various values in the config.json file.
 Copy the **config_sample.json** file to **config.json** and modify the various entries.

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
        "DefaultBranch": false,
        "Url": "http://X.X.X.X/",
        "Apiver": "1.0",
        "Baseapi": "rest/api/",
        "Protocol": "http",
        "FileExclusion":".cloc_bitbucketdc_ignore",
        "Period":-1,
        "Factor":33,
        "Multithreading":true,
        "Stats": false,
        "Workers": 50,
        "NumberWorkerRepos":50
      },
      "BitBucket": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxx",
        "DevOps": "bitbucket",
        "Workspace":"xxxxxxxxxxxxx",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "DefaultBranch": false,
        "Url": "https://api.bitbucket.org/",
        "Apiver": "2.0",
        "Baseapi": "bitbucket.org",
        "Protocol": "http",
        "FileExclusion":".cloc_bitbucket_ignore",
        "Period":-1,
        "Factor":33,
        "Multithreading":true,
        "Stats": false,
        "Workers": 50,
        "NumberWorkerRepos":50
      },
      
      "Github": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxxxx",
        "DevOps": "github",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "DefaultBranch": false,
        "Url": "https://api.github.com/",
        "Apiver": "",
        "Baseapi": "api.github.com/",
        "Protocol": "https",
        "FileExclusion":".cloc_github_ignore",
        "Period":-1,
        "Factor":33,
        "Multithreading":true,
        "Stats": false,
        "Workers": 50,
        "NumberWorkerRepos":50
      },
      "Gitlab": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxx",
        "DevOps": "gitlab",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "DefaultBranch": false,
        "Url": "https://gitlab.com/",
        "Apiver": "v4",
        "Baseapi": "api/",
        "Protocol": "https",
        "FileExclusion":".cloc_gitlab_ignore",
        "Factor":33,
        "Multithreading":true,
        "Stats": false,
        "Workers": 50,
        "NumberWorkerRepos":50

      },
      "Azure": {
        "Users": "xxxxxxxxxxxxxx",
        "AccessToken": "xxxxxxxxxxxxxx",
        "Organization": "xxxxxxxx",
        "DevOps": "azure",
        "Project": "",
        "Repos": "",
        "Branch": "",
        "DefaultBranch": false,
        "Url": "https://dev.azure.com/",
        "Apiver": "7.1",
        "Baseapi": "_apis/git/",
        "Protocol": "https",
        "FileExclusion":".cloc_azure_ignore",
        "Factor":33,
        "Multithreading":true,
        "Stats": false,
        "Workers": 50,
        "NumberWorkerRepos":50
      },
      "File": {
        "Organization": "xxxxxxxxx",
        "DevOps": "file",
        "Directory":"",
        "FileExclusion":".cloc_file_ignore",
        "FileLoad":".cloc_file_load"

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

If '**Projects**' and '**Repos**' are not specified, the analysis will be conducted on all repositories. You can specify a project name (PROJECT_KEY) in '**Projects**', and the analysis will be limited to the specified project. If you specify '**Repos**' (REPO_SLUG), the analysis will be limited to the specified repositories.
```json
"Project": "",
"Repos": "",
```
❗️ The '**Projects**' entry is supported exclusively on the BitBucket platform.

For Bitbucket DC, you must provide the URL with your server address and change the '**Protocol**' entry if you are using an https connection , ending with '**/**'. The '**Branch**' input allows you to select a specific branch for all repositories within an organization or project, or for a single repository. For example, if you only want all branches to be "main", '**"Branch":"main"**' .
```json
 "Url": "http://X.X.X.X/"
 ```
You can create a **.cloc_'your_platform'_ignore** file to ignore projects or repositories in the analysis. 
```json
   "FileExclusion":".cloc_bitbucketdc_ignore"
```
The syntax of this file is as follows for BitBucket:

```
REPO_SLUG
PROJECT_KEY 
PROJECT_KEY/REPO_SLUG
```

```
- REPO_SLUG = for one Repository
- PROJECT_KEY = for one Project
- PROJECT_KEY/REPO_SLUG For un Repository in one Project
```

The syntax of this file is as follows for GitHub:

```
REPO1_SLUG
REPO2_SLUG
...
```

```
- REPO1_SLUG = for one Repository
```

The syntax of this file is as follows for File:

```
DIRECTORY_NAME
FILE_NAME
...
```

 ✅  Config.json File Settings

❗️ For the **File** mode, if you want to have a list of directories to analyze, you create a **.cloc_file_load** file and add the directories or files to be analyzed line by line.If the **.cloc_file_load**. file is provided, its contents will override the **Directory** parameter."

❗️ The parameters **'Period'**, **'Factor'**, and **'Stats'** should not be modified as they will be used in a future version.

❗️ The parameters **'Multithreading'** and **'Workers'** initialize whether multithreading is enabled or not, allowing parallel analysis. You can disable it by setting **'Multithreading'** to **false**. **'Workers'** corresponds to the number of concurrent analyses.

❗️ The boolean parameters **DefaultBranch**, if set to true, specifies that only the default branch of each repository should be analyzed. If set to false, it will analyze all branches of each repository to determine the most important one.

 ✅ Run GoLC

 To launch GoLC with the following command, you must specify your DevOps platform. In this example, we analyze repositories hosted on Bitbucket Cloud. The supported flags for -devops are :
 ```bash
flag : <BitBucketSRV>||<BitBucket>||<Github>||<Gitlab>||<Azure>||<File>

 ```
 ❗️ And for now, only the **BitBucketSRV** and **BitBucket** flags are supported...

```bash

If the Results directory exists, GoLC will prompt you to delete it before starting a new analysis and will also offer to save the previous analysis. If you respond 'y', a Saves directory will be created containing a zip file, which will be a compressed version of the Results directory.

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

        🟢  Analyse Projet: sri 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: Bitbucket Pipes 
          ✅ The number of Repository found is: 5
        ✅ Repo: sonarcloud-quality-gate - Number of branches: 9
        ✅ Repo: sonarcloud-scan - Number of branches: 8
        ✅ Repo: official-pipes - Number of branches: 14
        ✅ Repo: sonarqube-scan - Number of branches: 7
        ✅ Repo: sonarqube-quality-gate - Number of branches: 2
         ........


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

