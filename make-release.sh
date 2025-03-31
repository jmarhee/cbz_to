#!/bin/bash

TAG=$TAG
TAG_COMMIT=$(git rev-parse --short ${TAG})

REPO_NAME=$(git remote -v | grep push | awk '{print $2}' | sed -e 's|git@github.com:||g' -e 's|.git||g')

AUTH_TOKEN=$GITHUB_TOKEN

curl -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO_NAME/releases" \
  -d "{
    \"tag_name\": \"$TAG\",
    \"name\": \"Release $TAG\",
    \"body\": \"Release notes for $TAG_COMMIT\",
    \"draft\": false,
    \"prerelease\": false
  }"
