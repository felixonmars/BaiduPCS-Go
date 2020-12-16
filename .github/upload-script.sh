#!/bin/bash
#
# Upload binary artifacts when a new release is made.
#
# Ensure that the GITHUB_TOKEN secret is included


# versionTag='latest'
# GITHUB_REF="refs/tags/v1.1"
# # TAG=`echo "${GITHUB_REF}" | grep "refs/tags" | tr -d 'refs/tags' `
# if [[ "$GITHUB_REF" = refs/tags/* ]];
# then
#     versionTag=`echo "${GITHUB_REF}" | grep "refs/tags" | tr -d 'refs/tags' `
# fi
# echo "tag ## ${GITHUB_REF} ## ${versionTag}"

if [[ -z "$GITHUB_TOKEN" ]]; then
  echo "Set the GITHUB_TOKEN env variable."
  exit 1
fi

# Ensure that there is a pattern specified.
if [[ -z "$1" ]]; then
    echo "Missing file (pattern) to upload."
    exit 1
fi


#
# In the past we invoked a build-script to generate the artifacts
# prior to uploading.
#
# Now we no longer do so, they must exist before they are uploaded.
#
# Test for them here.
#

# Have we found any artifacts?
found=
for file in $*; do
  echo "file:>>${file}<<"
    if [ -e "${file}" ]; then
        found=1
    fi
done

#
# Abort if missing.
#
if [ -z "${found}" ]; then

    echo "*****************************************************************"
    echo " "
    echo " Artifacts are missing, and this action no longer invokes the "
    echo " legacy-build script."
    echo " "
    echo " Please see the README.md file for github-action-publish-binaries"
    echo " which demonstrates how to build AND upload artifacts."
    echo " "
    echo "*****************************************************************"
    echo "allparams : $* "
    today=$(date +"%Y_%m_%d-%H_%M_%S")
    echo "today=>>${today}<< sha=>>${GITHUB_SHA}<<"
    pwd
    ls -lah

    exit 1
fi

# Prepare the headers for our curl-command.
AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"


echo "用于存储用户数据的 GitHub 主目录路径。 例如 /github/home。   HOME=${HOME}"
echo "工作流程的名称。   GITHUB_WORKFLOW=${GITHUB_WORKFLOW}"
echo "操作唯一的标识符 (id)。   GITHUB_ACTION=${GITHUB_ACTION}"
echo "Always set to true when GitHub 操作 is running the workflow. You can use this variable to differentiate when tests are being run locally or by GitHub 操作.   GITHUB_ACTIONS=${GITHUB_ACTIONS}"
echo "发起工作流程的个人或应用程序的名称。 例如 octocat。   GITHUB_ACTOR=${GITHUB_ACTOR}"
echo "所有者和仓库名称。 例如 octocat/Hello-World。   GITHUB_REPOSITORY=${GITHUB_REPOSITORY}"
echo "触发工作流程的 web 挂钩事件的名称。   GITHUB_EVENT_NAME=${GITHUB_EVENT_NAME}"
echo "具有完整 web 挂钩事件有效负载的文件路径。 例如 /github/workflow/event.json。   GITHUB_EVENT_PATH=${GITHUB_EVENT_PATH}"
cat " GITHUB_EVENT_PATH=${GITHUB_EVENT_PATH}"
echo "============="
echo "GitHub 工作空间目录路径。 如果您的工作流程使用 actions/checkout 操作，工作空间目录将包含存储仓库副本的子目录。 如果不使用 actions/checkout 操作，该目录将为空。 例如 /home/runner/work/my-repo-name/my-repo-name。   GITHUB_WORKSPACE=${GITHUB_WORKSPACE}"
echo "触发工作流程的提交 SHA。 例如 ffac537e6cbbf934b08745a378932722df287a53。   GITHUB_SHA=${GITHUB_SHA}"
echo "触发工作流程的分支或标记参考。 例如 refs/heads/feature-branch-1。 如果分支或标记都不适用于事件类型，则变量不会存在。   GITHUB_REF=${GITHUB_REF}"
echo "仅为复刻的仓库设置。 头部仓库的分支。   GITHUB_HEAD_REF=${GITHUB_HEAD_REF}"
echo "仅为复刻的仓库设置。 基础仓库的分支。   GITHUB_BASE_REF=${GITHUB_BASE_REF}"
echo "\$*=$*"
echo "\$#=$#"

repository_name=$(jq --raw-output '.repository.name' "$GITHUB_EVENT_PATH")
echo "repository_name=${repository_name}"

versionTag='latest'
# "ref": "refs/tags/v1.1",
# TAG=`echo "${GITHUB_REF}" | grep "refs/tags" | tr -d 'refs/tags' `
if [[ "$GITHUB_REF" = refs/tags/* ]];
then
    versionTag=`echo "${GITHUB_REF}" | grep "refs/tags" | tr -d 'refs/tags' `
fi
echo "tag ## ${GITHUB_REF} ## ${versionTag}"

RELEASE_ID=`.github/github_release.sh list_releases "${GITHUB_ACTOR}" "${repository_name}" | grep "${versionTag}" | awk '{print $3}'`

if [ "${RELEASE_ID}" = "" ];
then
    .github/github_release.sh create_release "${GITHUB_ACTOR}" "${repository_name}" "${versionTag}"
    RELEASE_ID=`.github/github_release.sh list_releases "${GITHUB_ACTOR}" "${repository_name}" | grep "${versionTag}" | awk '{print $3}'`
fi

# Create the correct Upload URL.
# RELEASE_ID=$(jq --raw-output '.release.id' "$GITHUB_EVENT_PATH")

# For each matching file..
for file in $*; do

    echo "Processing file ${file}"

    if [ ! -e "$file" ]; then
        echo "***************************"
        echo " file not found - skipping."
        echo "***************************"
        continue
    fi

    if [ ! -s "$file" ]; then
        echo "**************************"
        echo " file is empty - skipping."
        echo "**************************"
        continue
    fi


    FILENAME=$(basename "${file}")

    UPLOAD_URL="https://uploads.github.com/repos/${GITHUB_REPOSITORY}/releases/${RELEASE_ID}/assets?name=${FILENAME}"
    echo "Upload URL is ${UPLOAD_URL}"

    # Generate a temporary file.
    tmp=$(mktemp)

    # Upload the artifact - capturing HTTP response-code in our output file.
    response=$(curl \
        -sSL \
        -XPOST \
        -H "${AUTH_HEADER}" \
        --upload-file "${file}" \
        --header "Content-Type:application/octet-stream" \
        --write-out "%{http_code}" \
        --output $tmp \
        "${UPLOAD_URL}")

    # If the curl-command returned a non-zero response we must abort
    if [ "$?" -ne 0 ]; then
        echo "**********************************"
        echo " curl command did not return zero."
        echo " Aborting"
        echo "**********************************"
        cat $tmp
        rm $tmp
        exit 1
    fi

    # If upload is not successful, we must abort
    if [ $response -ge 400 ]; then
        echo "***************************"
        echo " upload was not successful."
        echo " Aborting"
        echo " HTTP status is $response"
        echo "**********************************"
        cat $tmp
        rm $tmp
        exit 1
    fi

    # Show pretty output, since we already have jq
    cat $tmp | jq .
    rm $tmp

done
