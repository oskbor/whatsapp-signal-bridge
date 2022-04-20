#!/bin/bash

set -e
# modified from 
function increase_version (){
    current_version=$1
    mode=$2
    regex="([0-9]+).([0-9]+).([0-9]+)"
    if [[ $current_version =~ $regex ]]; then
      major="${BASH_REMATCH[1]}"
      minor="${BASH_REMATCH[2]}"
      patch="${BASH_REMATCH[3]}"

      if [[ "$mode" == "major" ]]; then
        major=$((major+1))
        minor=0
        patch=0
      elif [[ "$2" == "minor" ]]; then
        minor=$((minor+1))
        patch=0
      elif [[ "$2" == "patch" ]]; then
        patch=$((patch+1))
      fi
      
      echo "${major}.${minor}.${patch}"
     
    else
      >&2 echo "wrong version number $current_version. "
      >&2 echo "major.minor.patch (e.g. 1.0.56) expected"
      exit 0
    fi
}

    
TAG=$(git describe --tags --abbrev=0)
CURRENT_VERSION=${TAG#"v"}
echo "current version is $CURRENT_VERSION"

if [[ `git status --porcelain` ]]; then
  echo "You have uncommitted changes. Please commit or stash them before releasing."
  exit 0
else
  echo "No uncommitted changes. Good to go."
fi

case $1 in
     -h|--help)
       echo "Generate a new version in the same way of npx version"
       echo "Write the new version in the .version file and "
       echo "push the new version on the git repo, generating a tag"
       echo "Usage: "
       echo "$0 -m|--minor"
       echo "$0 -p|--patch"
       echo "$0 -M|--major"
       exit 0
       ;;
     -m|--minor) 
      mode=minor
      ;;
     -p|--patch)
      mode=patch
      ;;
     -M|--major)
      mode=major
      ;;
      *)
      echo "Invalid mode $1. Allowed values are "
      echo "-m|--minor"
      echo "-p|--patch"
      echo "-M|--major"
      exit 0
      ;;
esac
echo "mode is $mode"
NEW_VERSION=$(increase_version "$CURRENT_VERSION" "$mode")
echo "New version is $NEW_VERSION"

./build.sh 
docker tag oskbor/whatsapp-signal-bridge:latest oskbor/whatsapp-signal-bridge:v$NEW_VERSION 
docker push oskbor/whatsapp-signal-bridge:v$NEW_VERSION
docker push oskbor/whatsapp-signal-bridge:latest
git commit -m"v$NEW_VERSION"  --allow-empty && git tag v"$NEW_VERSION" && git push origin --tags
