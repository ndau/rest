This document is intended to serve as a guide and plan to create a static site that is automatically deployed and hosted with S3. Replace instances of `blockchain-explorer` with your site.

## IAM setup

  - create IAM user in AWS `blockchain-explorer-ci-cd` with no permissions

## S3 setup

  - create a bucket `blockchain-explorer.service.ndau.tech`
    - upload dummy `index.html` for testing
    - give read/write permissions to `blockchain-explorer-ci-cd`
      - [ ] should be able to use blockchain-explorer-ci-cd credentials and aws-cli to upload to s3
    - give public read permissions
    ```
      {
        "Version":"2012-10-17",
        "Statement":[{
          "Sid":"PublicReadGetObject",
          "Effect":"Allow",
          "Principal": "*",
          "Action":["s3:GetObject"],
          "Resource":["arn:aws:s3:::blockchain-explorer.service.ndau.tech/*"]
        }]
      }
      ```
      - [ ] should be able to hit the url https://s3.amazonaws.com/blockchain-explorer.service.ndau.tech/index.html
  - configure bucket for static website hosting: Bucket | Properties | Static website hosting. Add index.html.
  - [optional] configure any redirects forom other domains by adding additional buckets and setting them up to redirect to the real target bucket.

Turning on cloudfront for the S3 bucket would transfer the content to Amazon's CDN. Local users wouldn't notice much of a difference but global users would. Redirects are possible so multiple buckets wouldn't be required for `www.`. It also supports CORS. https://i.stack.imgur.com/eEDGb.png

## Route53

  - Create an A record for the domain `blockchain-explorer.service.ndau.tech`.
    - configure it to do simple aliasing to the s3 bucket
  - Wait for dns to propagate and test
  - [ ] should be able to visit http://blockchain-explorer.service.ndau.tech/ and see the dummy index.html

## Circle CI

- Turn on blockchain-explorer repo in circleci
- Get credentials for `blockchain-explorer-ci-cd` and add them as environment variables
  - `AWS_ACCESS_KEY_ID` - from csv
  - `AWS_SECRET_ACCESS_KEY` - from csv
  - `AWS_DEFAULT_REGION` - us-east-1
  - [ ] `aws iam list-users` should work
- Environment variables for configuration can be stored in
    https://circleci.com/gh/oneiro-ndev/commands/edit#env-vars
- Add config.yaml to blockchain repo
  - config.yaml should
    - [ ] start with a container with aws-cli
    - [ ] run tests and pass/fail the build
    - [ ] upload static contents using `aws s3 mv` for master commits
        - [optional] upload to staging server for other commits?
    - backup of old contents not necessary because git will keep them
- connect git repo to show branch test failures
  - [ ] should see an ✗ or a ✓ for branches and PRs
- add badge to readme
  - [ ] should show status of master build
