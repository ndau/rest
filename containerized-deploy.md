# Deployment with containers

Intent - To provide a CI/CD pipeline for microservices based on ECS and CircleCI.

This design document is a work in progress.

## Infrastructure and services

* Github - hosts code.
* CircleCI
  - runs scripts on github events
  - can store secrets in environment variables
* Docker - Creates images and runs containers
* ECS - Runs docker containers
* ECR - Stores docker images
* bash scripts that are set up by default to build test and dockerize a default hello world.
  * build.sh
  * test.sh
  * dockerize.sh
  * Dockerfile


# Flow


## branch commit

1. Commit and push to github
2. CircleCI
  * test
  * build

## master commit

1. Commit and push to github
2. CircleCI
  * test
  * build
  * dockerize
  * upload to ECR
  * update task definition with ECS
3. ECR saves the image
4. ECS updates the service
5. Integration tests

## -dev commit

1. Commit and push to github
2. CircleCI
  * test
  * build
  * dockerize
  * upload to ECR


## tagged `-staging` commit

1. Push tag to github
2. CircleCI
  * update staging task definition with ECS
3. ECS updates the staging service
4. Integration tests run on staging deployment

## tagged `-prod` commit

1. Push tag to github
2. CircleCI
  * update task definition with ECS
3. ECS updates the service
4. Integration tests run on deployment


# Infrastructure setup

* [done] github and repos
* [done] CircleCI
* Create an ALB
  * ALB doesn't do path rewrites, routes in apps need to start with /basepath/
* Attach ALB to a domain and configure route
* Deploy a hello world service


# Setup per project

* add route to ALB
* create an ECS cluster
* create task definition
* register definition
* create a IAM user for CircleCI to use AWS
* enable repo in circleCI

# ECS and Fargate

* In general, fargate has fewer steps to get started, and
* fargate can be configured to use an ALB so that traffic from the ALB comes to a service created using Fargate. [docs](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_GetStarted.html)
* ECS is docker as a service
* Fargate is ECS as a service lol
  * task definitions aren't exactly the same, but close
  * don't have to select instance types

[How to migrate from ECS to Fargate](https://aws.amazon.com/blogs/compute/migrating-your-amazon-ecs-containers-to-aws-fargate/)
  * Might be deployment delays with fargate https://www.reddit.com/r/aws/comments/7ixf1q/ecs_fargate_bluegreen_deployments/dr26r5j

## Todo

How do updates work for no downtime?
  - https://aws.amazon.com/blogs/compute/bluegreen-deployments-with-amazon-ecs/ This shows how to do blue green with ECS largely doing it automatically.
  - This article makes it look hard. Not sure why  - https://blog.osones.com/en/update-your-ecs-container-instances-with-no-downtime.html


# Debugging

The container can be downloaded from ECR and run on its own with a local machine.
  * A local-docker.sh script could be used to simplify the process and automate the following steps
    * login with ECR / ask for creds
    * select which docker image to pull from ECR
      * latest (delete local copy first)
      * specific sha
      * highest version
    * pull the docker image from ECR
    * Run the docker image with environment variables `docker run --env-file local.env`

The executable can be debugged like normal, though should probably get variables from local.env as well.


# Staging

to provide a staging deploy, each service must
* use templates to generate a secondary task definition


# Observability
  Honeycomb


# CircleCI environment variables
  AWS keys and config
  Honeycomb creds
  Secrets?
  ECR endpoint

# Secrets

CircleCI environment variables are the simplest way to pass config to the running container. When a task definition is updated, it can include environment variables. This way, environment variables live in AWS task definitions, which on circle are transient files, and do not show up in the image. They can also be baked into the image, but for secrets this is less desirable. If no one will ever get access to any of our images in ECR, this baking them in is ok. Otherwise, they live in task definitions.


# Reference

https://circleci.com/docs/2.0/ecs-ecr/

most of what below comes from the link above


# config.yml

aws configure set default.region us-east-1
aws configure set default.output json

## deploy.sh



echo 'export ECR_REPOSITORY_NAME="${AWS_RESOURCE_NAME_PREFIX}"' >> $BASH_ENV
echo 'export ECS_CLUSTER_NAME="${AWS_RESOURCE_NAME_PREFIX}-cluster"' >> $BASH_ENV
echo 'export ECS_SERVICE_NAME="${AWS_RESOURCE_NAME_PREFIX}-service"' >> $BASH_ENV

export ECS_TASK_FAMILY_NAME="${AWS_RESOURCE_NAME_PREFIX}-service"
export ECS_CONTAINER_DEFINITION_NAME="${AWS_RESOURCE_NAME_PREFIX}-service"
export EXECUTION_ROLE_ARN="arn:aws:iam::$AWS_ACCOUNT_ID:role/${AWS_RESOURCE_NAME_PREFIX}-ecs-execution-role"

> TODO Figure out why bash variables don't seem to work right in CircleCI. Current workaround is to `source $BASH_ENV` to re export them.

#!/usr/bin/env bash

# more bash-friendly output for jq
JQ="jq --raw-output --exit-status"

configure_aws_cli(){
	aws --version
	aws configure set default.region us-east-1
	aws configure set default.output json
}

deploy_cluster() {

    family="sample-webapp-task-family"

    make_task_def
    register_definition
    if [[ $(aws ecs update-service --cluster sample-webapp-cluster --service sample-webapp-service --task-definition $revision | \
                   $JQ '.service.taskDefinition') != $revision ]]; then
        echo "Error updating service."
        return 1
    fi

    # wait for older revisions to disappear
    # not really necessary, but nice for demos
    for attempt in {1..30}; do
        if stale=$(aws ecs describe-services --cluster sample-webapp-cluster --services sample-webapp-service | \
                       $JQ ".services[0].deployments | .[] | select(.taskDefinition != \"$revision\") | .taskDefinition"); then
            echo "Waiting for stale deployments:"
            echo "$stale"
            sleep 5
        else
            echo "Deployed!"
            return 0
        fi
    done
    echo "Service update took too long."
    return 1
}

make_task_def(){
	task_template='[
		{
			"name": "go-sample-webapp",
			"image": "%s.dkr.ecr.us-east-1.amazonaws.com/go-sample-webapp:%s",
			"essential": true,
			"memory": 200,
			"cpu": 10,
			"portMappings": [
				{
					"containerPort": 8080,
					"hostPort": 80
				}
			]
		}
	]'

	task_def=$(printf "$task_template" $AWS_ACCOUNT_ID $CIRCLE_SHA1)
}

push_ecr_image(){
	eval $(aws ecr get-login --region us-east-1 --no-include-email)
	docker push $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-sample-webapp:$CIRCLE_SHA1
}

register_definition() {

    if revision=$(aws ecs register-task-definition --container-definitions "$task_def" --family $family | $JQ '.taskDefinition.taskDefinitionArn'); then
        echo "Revision: $revision"
    else
        echo "Failed to register task definition"
        return 1
    fi

}

configure_aws_cli
push_ecr_image
deploy_cluster

# test the deployment

TARGET_GROUP_ARN=$(aws ecs describe-services --cluster $ECS_CLUSTER_NAME --services $ECS_SERVICE_NAME | jq -r '.services[0].loadBalancers[0].targetGroupArn')
ELB_ARN=$(aws elbv2 describe-target-groups --target-group-arns $TARGET_GROUP_ARN | jq -r '.TargetGroups[0].LoadBalancerArns[0]')
ELB_DNS_NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns $ELB_ARN | jq -r '.LoadBalancers[0].DNSName')
curl http://$ELB_DNS_NAME | grep "Hello World!"


# other thoughts for later

going from kubernetes to X
  considerations
    * volumes
    * networking
    * initialization
    * managing configuration

* write a doc on how to go from localnet to one big container
  - make .localnet directory, share that directory as a volume
  - [one or the other]
    - ./bin/setup.sh N
    - Copy already setup files to the shared .localnet directory.
  - copy everything in commands
  - reset, run


* write a doc on how to get a nodegroup up with minikube
  - install dependencies
  - give creds to user with access to ECR
  - replace genesis.json with cannonical one in up.sh
  - provide deploy script
  - run minikube
  - run up.sh

~~~
* write a doc on how one might go about moving from kubernetes to docker compose
  - Convert all commands run in pods to a script
  - init containers are now image dependencies
  - environment variables map
  - one big volume is shared with specific directories for each service
  - internal service names are
~~~

