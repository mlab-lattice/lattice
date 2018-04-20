#!/usr/bin/env bash

set -e

if [[ -z ${REPOSITORY} ]]; then
    REPOSITORY=gcr.io/$(gcloud config get-value project)
fi

set -u

PROJECT=$(gcloud config get-value project)
IMAGES=$(gcloud container images list --repository ${REPOSITORY} --filter="name:${FILTER}")
printf "will delete the following images: ${IMAGES}"

echo && echo
read -p "Are you sure you want to delete these images in ${REPOSITORY} [y/N]? " -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
    while [[ 1 ]]
    do
        IMAGE=$(gcloud container images list --repository ${REPOSITORY} --filter="name:${FILTER}" --limit 1 2>/dev/null | tail -n 1)
        if [[ -z ${IMAGE} ]]; then
            break
         fi

        echo deleting ${IMAGE}...

        while [[ 1 ]]
        do
            DIGESTS=$(gcloud container images list-tags ${IMAGE} --format='get(digest)')
            if [[ -z ${DIGESTS} ]]; then
                break
            fi

            IMAGES=""
            for digest in ${DIGESTS}; do
                IMAGES="${IMAGES} ${IMAGE}@${digest}"
            done
            gcloud container images delete ${IMAGES} --force-delete-tags --quiet
        done
        echo
    done
fi
