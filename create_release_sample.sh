#!/bin/bash

export TAG="V1.0.7" # Release TAG in GitHub
export Release1="v1.0.7" # Release Number
export buildpath="xxxx/Release/"  # Replace with the path where the release files are located

GITHUB_TOKEN="XXXXXXXXX" # Replace with your token
GITHUB_ORG="xXXXXXXXXX"    # Replace with your organization name
GITHUB_REPO="XXXXXXXXX"   # Replace with the name of your GitHub repository

# Set a description for the release
RELEASE_DESCRIPTION="Add Zip parameter improves performance if you have very large repositories; to enable it, you need to set it to true in config.json file.\n\"


CMD=`PWD`

# Function create release
create_release() {
  local tag="$1"
  local name="$2"
  local body="$3"

  echo "Création d'une nouvelle release avec le tag '$tag'..."
  curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/json" \
    "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases" -d "{
      \"tag_name\": \"$tag\",
      \"target_commitish\": \"main\",
      \"name\": \"$name\",
      \"body\": \"$body\",
      \"draft\": false,
      \"prerelease\": false
    }" > /dev/null

  echo "Release created."
}

# Function to retrieve the ID of the existing asset
find_asset_id() {
  local asset_name="$1"
  local release_id="$2"
  
  echo $(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO=/releases/$release_id/assets" | \
    jq -r ".[] | select(.name == \"$asset_name\") | .id")
}

# Function to delete an existing asset
delete_asset() {
  local asset_id="$1"

  echo "Suppression de l'asset existant..."
  curl -s -X DELETE -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases/assets/$asset_id"
  echo "Asset deleted."
}

# Function to upload a file to the release
upload_asset() {
  local upload_url="$1"
  local file_path="$2"

  echo "Uploader le fichier : $(basename "$file_path")..."
  curl -s -X POST "$upload_url?name=$(basename "$file_path")" \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/zip" \
    --data-binary @"$file_path"
  echo "File uploaded to release successfully."
}


# Function update description
update_release_description() {
  local release_id="$1"
  local new_body="$2"

  echo "Updated release description..."
  curl -s -X PATCH -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/json" \
    "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases/$release_id" -d "{
      \"body\": \"$new_body\"
    }"
  
  echo "Description updated."
}

#----------------------- Begin Build --------------------------------#

# Buil arm64 Darwin

export GOARCH=arm64
export GOOS=darwin
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST


go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

# Buil arm64 Linux

export GOOS=linux
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST

go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

# Buil arm64 Windows

export GOOS=windows
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST

go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

# Buil amd64 Darwin

export GOARCH=amd64
export GOOS=darwin
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST

go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

# Buil amd64 Linux

export GOOS=linux
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST

go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

# Buil amd64 Windows

export GOOS=windows
export DEST=${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}
export FILE_DEST=golc_${Release1}_${GOOS}_${GOARCH}

mkdir -p $DEST

go build -ldflags "-X main.version=${TAG}" -o ${DEST}/golc golc.go
go build -o ${DEST}/ResultsAll ResultsAll.go
cp README.md  ${DEST}/
cp -r imgs ${DEST}/
cp -r dist ${DEST}/
cp config_sample.json ${DEST}/config.json
cd ${buildpath}${Release1}/${GOARCH}/${GOOS}/
zip -r ${FILE_DEST}.zip ${FILE_DEST}
cd $CMD

#------------------------------ End Build ------------------------------------#

# Begin to push Releae in GitHub Repository

# Retrieve information from existing release
RELEASE_RESPONSE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases/tags/$TAG")


if [[ $(echo "$RELEASE_RESPONSE" | jq -r '.message') == "Not Found" ]]; then
  echo "The release for tag '$TAG' does not exist. Creating release..."
  create_release "$TAG" "$TAG" "$RELEASE_DESCRIPTION"
  # Retrieve information from the newly created release
  RELEASE_RESPONSE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases/tags/$TAG")
fi

# Retrieve the upload URL and release ID
UPLOAD_URL=$(echo "$RELEASE_RESPONSE" | jq -r '.upload_url' | sed "s/{?name,label}//")
RELEASE_ID=$(echo "$RELEASE_RESPONSE" | jq -r '.id')

# Description update
update_release_description "$RELEASE_ID" "$RELEASE_DESCRIPTION"

# Retrieve the list of files from the release
ASSETS_RESPONSE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/$GITHUB_ORG/$GITHUB_REPO/releases/$RELEASE_ID/assets")


declare -a GOARCH_VALUES=("amd64" "arm64")
declare -a GOOS_VALUES=("linux" "windows" "darwin")

# Upload zip files to each directory for each combination of GOARCH and GOOS
for GOARCH in "${GOARCH_VALUES[@]}"; do
    for GOOS in "${GOOS_VALUES[@]}"; do
        zip_file="${buildpath}${Release1}/${GOARCH}/${GOOS}/golc_${Release1}_${GOOS}_${GOARCH}.zip"
        
       # Find the ID of the existing asset with the same name
  
        EXISTING_ASSET_ID=$(echo "$ASSETS_RESPONSE" | jq -r ".[] | select(.name == \"$(basename $zip_file)\") | .id")

        # Delete existing asset, if found
        if [ ! -z "$EXISTING_ASSET_ID" ]; then
            delete_asset "$EXISTING_ASSET_ID"
        fi
        upload_asset "$UPLOAD_URL" "$zip_file"
    done
done

cd $CMD


