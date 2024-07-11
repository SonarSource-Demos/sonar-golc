The 'ResultsAll' program generates a 'GlobalReport.pdf' file in the 'Results' directory. It prompts you if you want to view the results on a web interface.It starts an HTTP service on the default port 8090.To stop the local HTTP service, press the Ctrl+C keys

---


## Quick reference
- Maintained by:
[colussim](https://github.com/colussim/GoLC/Docker)

## Supported tags and respective Dockerfile links
- [arm64-1.0.3](https://github.com/colussim/GoLC/blob/ver1.0.3/Docker/v1.0.3/Dockerfile.ResultsAll.arm64), [amd64-1.0.3](https://github.com/colussim/GoLC/blob/ver1.0.3/Docker/v1.0.3/Dockerfile.ResultsAll.amd64)

## Quick reference (cont.)
- Supported architectures: ([more info⁠](https://github.com/docker-library/official-images#architectures-other-than-amd64))

    arm64v8, amd64

## How to use this image

Before running ReportsAll, you need to run [GoLC](https://hub.docker.com/r/mcolussi/golc) and map the volume used for the analysis with GoLC.

✅ Running the container: 

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
