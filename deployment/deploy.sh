#!/bin/bash

export AWS_PAGER=""

AWS_ACCOUNT_ID=$1
STACK_NAME=golang-lambda
REPO_NAME=golang-lambda

# login to ecr
aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.eu-central-1.amazonaws.com

# create repo if it doesn't exist already
aws ecr describe-repositories --repository-names ${REPO_NAME} || aws ecr create-repository --image-tag-mutability MUTABLE --repository-name ${REPO_NAME}

# build the image locally
docker build -t docker-image:test .

# tag the image
docker tag docker-image:test $AWS_ACCOUNT_ID.dkr.ecr.eu-central-1.amazonaws.com/$REPO_NAME:latest

# push image to ecr
docker push $AWS_ACCOUNT_ID.dkr.ecr.eu-central-1.amazonaws.com/$REPO_NAME:latest

# deploy stack
aws cloudformation deploy --template-file ./deployment/stack.yaml --stack-name $STACK_NAME --region eu-central-1 --capabilities CAPABILITY_NAMED_IAM --parameter-overrides lambdaImageUri=$AWS_ACCOUNT_ID.dkr.ecr.eu-central-1.amazonaws.com/$REPO_NAME:latest
