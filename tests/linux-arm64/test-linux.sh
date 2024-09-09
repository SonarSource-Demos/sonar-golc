#!/bin/bash

# Requires JQ to edit the config file
VERSION=1.0.6
LINUX_DISTRIBUTION=linux_arm64

CONFIG_FILE="config.json"
SONAR_GOLC_GITHUB_USERNAME="XXXXX"
SONAR_GOLC_GITHUB_TOKEN="XXXXX"
SONAR_GOLC_GITHUB_ORGANIZATION="XXXXX"

# Fresh install the SonarGoLC Artifacts
wget https://github.com/SonarSource-Demos/sonar-golc/releases/download/V$VERSION/golc_v$VERSION\_$LINUX_DISTRIBUTION.zip
unzip golc_v$VERSION\_$LINUX_DISTRIBUTION.zip -d golc_v$VERSION\_$LINUX_DISTRIBUTION/

# Use jq to modify the desired lines
cd golc_v$VERSION\_$LINUX_DISTRIBUTION/golc_v$VERSION\_$LINUX_DISTRIBUTION/ # TODO there is a double directory here
jq ".platforms.Github.Users = \"$SONAR_GOLC_GITHUB_USERNAME\" | .platforms.Github.AccessToken = \"$SONAR_GOLC_GITHUB_TOKEN\" | .platforms.Github.Organization = \"$SONAR_GOLC_GITHUB_ORGANIZATION\" | .platforms.Github.Workers = 1 | .platforms.Github.NumberWorkerRepos= 1" $CONFIG_FILE > tmp.$$.json && mv tmp.$$.json $CONFIG_FILE

# Run the SonarGoLC tool on Github
./golc -devops Github

# Check if the TotalLines results are the correct amount for at least one of the repositories
cd Results/byfile-report/
JSON_FILE="Result_SonarSource-Demos_opencv_4.x_byfile.json"
TOTAL_LINES=$(jq -r '.TotalLines' $JSON_FILE)
if [ "$TOTAL_LINES" -e 2940614 ]; then
  echo "Success: TOTAL_LINES of $JSON_FILE is equal to 2940614."
  exit 0
else
  echo "Error: TOTAL_LINES of $JSON_FILE is not equal to 2940614."
  exit 1
fi
