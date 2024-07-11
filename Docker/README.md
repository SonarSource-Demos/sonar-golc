**GoLC** counts physical lines of source code in numerous programming languages across your Bitbucket Cloud, Bitbucket Data Center, GitHub, GitLab, Azure DevOps and local repositories.

---

## Quick reference
- Maintained by:
[colussim](https://github.com/colussim/GoLC/Docker)

## Supported tags and respective Dockerfile links
- [arm64-1.0.3](https://github.com/colussim/GoLC/blob/ver1.0.3/Docker/v1.0.3/Dockerfile.golc.arm64), [amd64-1.0.3](https://github.com/colussim/GoLC/blob/ver1.0.3/Docker/v1.0.3/Dockerfile.golc.amd64)

## Quick reference (cont.)
- Supported architectures: ([more info⁠](https://github.com/docker-library/official-images#architectures-other-than-amd64))

    arm64v8, amd64

## How to use this image

Before running GoLC, you need to configure your environment by initializing the various values in the config.json file.You will need to map a volume for the Scan Results.See the documentation [here](https://github.com/colussim/GoLC/)

 ✅ Running the container: 
 ```
:> docker run --rm -v /custom/Results_volume:/app/Results -v /custom/config.json:/app/config.json golc:arm64-1.0.3 -devops Github -docker

✅ Using configuration for DevOps platform 'Github'
Running in Docker mode


🔎 Analysis of devops platform objects ...
 Repos saved successfully!
          ✅ The number of Repo(s) found is: 1
                ✅ 1 Repo: sonar-golc - Number of branches: 4 - largest Branch: ver1.0.3 
✅ Result saved successfully!

✅ The largest Repository is <sonar-golc> in the organization <SonarSource-Demos> with the branch <ver1.0.3> 
✅ Total Repositories that will be analyzed: 1 - Find empty : 0 - Excluded : 0 - Archived : 0
✅ Total Branches that will be analyzed: 4

🔎 Analysis of Repos ...
 Waiting for workers...
                                                                                                 
        ✅ json report exported to /app/Results/Result_SonarSource-Demos_sonar-golc_ver1.0.3.json
✅ 2 The repository <sonar-golc> has been analyzed

🔎 Analyse Report ...

✅ Number of Repository analyzed in Organization <SonarSource-Demos> is 1 
✅ The repository with the largest line of code is in project <SonarSource-Demos> the repo name is <sonar-golc> with <41.48K> lines of code
✅ The total sum of lines of code in Organization <SonarSource-Demos> is : 41.48K Lines of Code


✅ Reports are located in the <'Results'> directory

✅ Time elapsed : 00:00:06


ℹ️  To generate and visualize results on a web interface, follow these steps: 
        ✅ run : ResultsAll
 ```

 ✅ Run Report

 Now we can start generating the report with the resultsall container.Install the [resultsall container](https://hub.docker.com/r/mcolussi/resultsall)
 You need to map the volume previously used for the analysis and map an available port for web access.

```
:> docker run --rm -p 8090:8090 -v /custom/Results_volume:/app/Results resultsall:arm64-1.0.3


✅ Results analysis recorded in Results/code_lines_by_language.json
✅ PDF generated successfully!
✅ Launching web visualization...
✅ Server started on http://localhost:8090
✅ please type < Ctrl+C> to stop the server
```


✅  Web UI
![report](https://github.com/colussim/GoLC/raw/main/imgs/webui.png)