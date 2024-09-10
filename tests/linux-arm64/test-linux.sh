#!/bin/bash

# Requires JQ to edit the config file
VERSION=1.0.6
LINUX_DISTRIBUTION=linux_arm64

CONFIG_FILE="config.json"

# Fresh install the SonarGoLC Artifacts
wget https://github.com/SonarSource-Demos/sonar-golc/releases/download/V$VERSION/golc_v$VERSION\_$LINUX_DISTRIBUTION.zip
unzip golc_v$VERSION\_$LINUX_DISTRIBUTION.zip -d golc_v$VERSION\_$LINUX_DISTRIBUTION/

# Change directory to the SonarGoLC Artifacts
cd golc_v$VERSION\_$LINUX_DISTRIBUTION/golc_v$VERSION\_$LINUX_DISTRIBUTION/ # TODO there is a double directory here

test_Github(){
  SONAR_GOLC_GITHUB_USERNAME="XXXXX"
  SONAR_GOLC_GITHUB_TOKEN="XXXXX"
  SONAR_GOLC_GITHUB_ORGANIZATION="XXXXX"

  # Use jq to modify the desired lines
  jq ".platforms.Github.Users = \"$SONAR_GOLC_GITHUB_USERNAME\" | .platforms.Github.AccessToken = \"$SONAR_GOLC_GITHUB_TOKEN\" | .platforms.Github.Organization = \"$SONAR_GOLC_GITHUB_ORGANIZATION\" | .platforms.Github.Workers = 1 | .platforms.Github.NumberWorkerRepos= 1" $CONFIG_FILE > tmp.$$.json && mv tmp.$$.json $CONFIG_FILE
  
  # Run the SonarGoLC tool on Github
  ./golc -devops Github

  # Check if the TotalLines results are the correct amount for at least one of the repositories
  cd Results/byfile-report/
  JSON_FILE="Result_SonarSource-Demos_opencv_4.x_byfile.json"
  TOTAL_LINES=$(jq -r '.TotalLines' $JSON_FILE)
  EXPECTED_TOTAL_LINES="2940614"
  # return to the root directory
  cd ../../

  # Check if the TotalLines are equal to the expected amount
  if [ "$TOTAL_LINES" -eq "$EXPECTED_TOTAL_LINES" ]; then
    echo "Success: TOTAL_LINES of $JSON_FILE is equal to $EXPECTED_TOTAL_LINES."
    return 0
  else
    echo "Error: TOTAL_LINES of $JSON_FILE is not equal to $EXPECTED_TOTAL_LINES."
    return 1
  fi
}

test_Azure(){
  SONAR_GOLC_AZDO_USERNAME="XXXXX"
  SONAR_GOLC_AZDO_ORGANIZATION="XXXXX"
  SONAR_GOLC_AZDO_TOKEN="XXXXX"

  # Use jq to modify the desired lines
  jq ".platforms.Azure.Users = \"$SONAR_GOLC_AZDO_USERNAME\" | .platforms.Azure.AccessToken = \"$SONAR_GOLC_AZDO_TOKEN\" | .platforms.Azure.Organization = \"$SONAR_GOLC_AZDO_ORGANIZATION\" | .platforms.Azure.Repos = \"$SONAR_GOLC_AZDO_REPO\"" $CONFIG_FILE > tmp.$$.json && mv tmp.$$.json $CONFIG_FILE

  # Run the SonarGoLC tool on Azure
  ./golc -devops Azure

  cd Results/byfile-report/
  JSON_FILE="Result_WebApp.NET_WebApp.NET_master_byfile.json"
  TOTAL_LINES=$(jq -r '.TotalLines' $JSON_FILE)
  EXPECTED_TOTAL_LINES="44696"
  # return to the root directory
  cd ../../

  # Check if the TotalLines are equal to the expected amount
  if [ "$TOTAL_LINES" -eq "$EXPECTED_TOTAL_LINES" ]; then
    echo "Success: TOTAL_LINES of $JSON_FILE is equal to $EXPECTED_TOTAL_LINES."
    return 0
  else
    echo "Error: TOTAL_LINES of $JSON_FILE is not equal to $EXPECTED_TOTAL_LINES."
    return 1
  fi
}

test_Gitlab(){
  SONAR_GOLC_GITLAB_USERNAME="XXXXX"
  SONAR_GOLC_GITLAB_ORGANIZATION="XXXXX"
  SONAR_GOLC_GITLAB_TOKEN="XXXXX" 
  # Use jq to modify the desired lines
  jq ".platforms.Gitlab.Users = \"$SONAR_GOLC_GITLAB_USERNAME\" | .platforms.Gitlab.AccessToken = \"$SONAR_GOLC_GITLAB_TOKEN\" | .platforms.Gitlab.Organization = \"$SONAR_GOLC_GITLAB_ORGANIZATION\"" $CONFIG_FILE > tmp.$$.json && mv tmp.$$.json $CONFIG_FILE

  # Run the SonarGoLC tool on Azure
  ./golc -devops Gitlab

  # cd Results/byfile-report/
  # JSON_FILE="Result_WebApp.NET_WebApp.NET_master_byfile.json"
  # TOTAL_LINES=$(jq -r '.TotalLines' $JSON_FILE)
  # EXPECTED_TOTAL_LINES="44696"
  # # return to the root directory
  # cd ../../

  # if [ "$TOTAL_LINES" -eq "$EXPECTED_TOTAL_LINES" ]; then
  #   echo "Success: TOTAL_LINES of $JSON_FILE is equal to $EXPECTED_TOTAL_LINES."
  #   return 0
  # else
  #   echo "Error: TOTAL_LINES of $JSON_FILE is not equal to $EXPECTED_TOTAL_LINES."
  #   return 1
  # fi
}

test_BitBucket(){
  SONAR_GOLC_BITBUCKET_USERNAME="XXXXX"
  SONAR_GOLC_BITBUCKET_ORGANIZATION="XXXXX"
  SONAR_GOLC_BITBUCKET_TOKEN=""

  # Use jq to modify the desired lines
  jq ".platforms.BitBucket.Users = \"$SONAR_GOLC_BITBUCKET_USERNAME\" | .platforms.BitBucket.AccessToken = \"$SONAR_GOLC_BITBUCKET_TOKEN\" | .platforms.BitBucket.Organization = \"$SONAR_GOLC_BITBUCKET_ORGANIZATION\"" $CONFIG_FILE > tmp.$$.json && mv tmp.$$.json $CONFIG_FILE

  # Run the SonarGoLC tool on Azure
  ./golc -devops BitBucket
}
  

# Run the tests
test_Github
# test_Azure
# test_Gitlab

# Uncomment the following line to keep the container running
# tail -f /dev/null
