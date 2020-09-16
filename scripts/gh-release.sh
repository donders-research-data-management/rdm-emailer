#!/bin/bash
#
########################################################################
# This is the script used to publish a release on the Github repository.
# It does the following steps:
#
# - make a new release tag on GitHub based on the given version number,
# - build RPM packages, and
# - upload RPMs as assessts of the release tag.
########################################################################

function get_script_dir() {
    ## resolve the base directory of this executable
    local SOURCE=$1
    while [ -h "$SOURCE" ]; do
        # resolve $SOURCE until the file is no longer a symlink
        DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
        SOURCE="$(readlink "$SOURCE")"

        # if $SOURCE was a relative symlink,
        # we need to resolve it relative to the path
        # where the symlink file was located

        [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
    done

    echo "$( cd -P "$( dirname "$SOURCE" )" && pwd )"
}

function new_release_post_data() {
    t=1
    p=2
    cat <<EOF
{
    "tag_name": "${tag}",
    "tag_commitish": "master",
    "name": "${tag}",
    "body": "Release ${tag}",
    "draft": false,
    "prerelease": $2
}
EOF
}

if [ $# -ne 2 ]; then
    echo "$0 <release_tag> <is_prerelease>"
    exit 1
fi

tag=$1
pre=$2
gh_token=""

GH_ORG="donders-research-data-management"
GH_REPO_NAME="rdr-emailer"

GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$GH_ORG/$GH_REPO_NAME"
GH_RELE="$GH_REPO/releases"
GH_TAG="$GH_REPO/releases/tags/$tag"
GH_REPO_ASSET_PREFIX="https://uploads.github.com/repos/$GH_ORG/$GH_REPO_NAME/releases"

# check if version tag already exists
response=$(curl -X GET $GH_TAG 2>/dev/null)
eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
if [ "$id" ]; then
    read -p "release tag already exists: ${tag}, continue? y/[n]: " cnt
    if [ "${cnt,,}" != "y" ]; then
        exit 1
    fi
fi

while [ "$gh_token" == "" ]; do
    read -s -p "github personal access token: " gh_token
done

# create a new tag with current master branch
# if the $id of the release is not available.
if [ ! "$id" ]; then
    response=$(curl -H "Authorization: token $gh_token" -X POST --data "$(new_release_post_data ${tag} ${pre})" $GH_RELE)
    eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
    [ "$id" ] || { echo "release tag not created successfully: ${tag}"; exit 1; }
fi

# copy over id to rid (release id)
rid=$id

# resolve to absolute directory where the rdr-emailer binaries are produced
mydir=$( get_script_dir $0 )
bindir=$( realpath ${mydir}/.. )

## get list of rdr-emailer.* executables to be uploaded as release assets 
bins=( $( find ${bindir} -type f -name "rdr-emailer.*" -exec file -i '{}' \; | grep 'x-executable; charset=binary' | awk -F ':' '{print $1}' ) )

## upload binaries as release assets
if [ ${#bins[@]} -gt 0 ]; then
    upload="y"
    read -p "upload ${#bins[@]} binaries as release assets?, continue? [y]/n: " upload
    for f in ${bins[@]}; do
        echo ${f}
        if [ "${upload,,}" == "y" ]; then
            fname=$( basename $f )
            # check if the asset with the same name already exists
            id=""
            eval $(echo "$response" | grep -C3 "name.:.\+${fname}" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
            if [ "$id" != "" ]; then
                # delete existing asset
                echo "deleting asset: ${id} ..."
                curl -H "Authorization: token $gh_token" -X DELETE "${GH_RELE}/assets/${id}"
            fi
            # post new asset
            echo "uploading ${f} ..."
            GH_ASSET="${GH_REPO_ASSET_PREFIX}/${rid}/assets?name=$(basename $f)"
            resp_upload=$( curl --data-binary @${f} \
                                -H "Content-Type: application/octet-stream" \
				-H "Authorization: token $gh_token" $GH_ASSET )
        fi
    done
fi
