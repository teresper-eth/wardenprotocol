#!/usr/bin/env bash

set -e

# Version control
commit_hash=$(git rev-parse HEAD)
commit_hash_short=$(git rev-parse --short HEAD)

# Set ARCH variable based on the architecture
architecture=$(uname -m)
if [ -z "$architecture" ]; then
    export ARCH="x86_64" # Linux, Windows (default)
elif [ "$architecture" == "x86_64" ]; then
    export ARCH="x86_64" # Linux, Windows
else
    export ARCH="aarch64" # Mac
fi

docker build \
       --build-arg ARCH="$ARCH" \
       --build-arg BUILD_DATE="$(git show -s --format=%ci "$commit_hash")"\
       --build-arg SERVICE=fusionkms \
       --build-arg GIT_SHA="$commit_hash" \
       -t "${ECR}"fusionkms:latest  \
       -t "${ECR}"fusionkms:"$commit_hash_short"  \
       -f Dockerfile-fusionkms ..

# must login with 'aws ecr get-login-password  --region eu-west-1 | docker login --username AWS --password-stdin 532153175488.dkr.ecr.eu-west-1.amazonaws.com'
docker tag fusionkms 532153175488.dkr.ecr.eu-west-1.amazonaws.com/qredo/production/fusionkms:latest
docker tag fusionkms 532153175488.dkr.ecr.eu-west-1.amazonaws.com/qredo/production/fusionkms:${commit_hash_short}
docker push 532153175488.dkr.ecr.eu-west-1.amazonaws.com/qredo/production/fusionkms:latest
docker push 532153175488.dkr.ecr.eu-west-1.amazonaws.com/qredo/production/fusionkms:${commit_hash_short}