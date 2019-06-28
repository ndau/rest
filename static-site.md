This document is intended to serve as a guide and plan to create a static site that is automatically deployed and hosted with S3. Replace instances of `blockchain-explorer` with your site.

## IAM setup

  - create IAM user in AWS `blockchain-explorer-ci-cd` with no permissions
  - [ ] should be able to execute `aws iam list-users`
  - upload credentials to 1password

## S3 setup

  - create a bucket `blockchain-explorer.service.ndau.tech`
    - upload dummy `index.html` for testing
    - give read/write permissions to `blockchain-explorer-ci-cd`
      - [ ] should be able to use blockchain-explorer-ci-cd credentials and aws-cli to upload to s3
    - give public read permissions
    ```
      {
          "Version": "2008-10-17",
          "Id": "Policy139763252196",
          "Statement": [
              {
                  "Sid": "Stmt1397633323327",
                  "Effect": "Allow",
                  "Principal": {
                      "AWS": "*"
                  },
                  "Action": "s3:GetObject",
                  "Resource": "arn:aws:s3:::blockchain-explorer.service.ndau.tech/*"
              },
              {
                  "Sid": "Stmt1397633323327",
                  "Effect": "Allow",
                  "Principal": {
                      "AWS": "*"
                  },
                  "Action": "s3:ListBucket*",
                  "Resource": [
                      "arn:aws:s3:::blockchain-explorer.service.ndau.tech",
                      "arn:aws:s3:::blockchain-explorer.service.ndau.tech/*"
                  ]
              },
              {
                  "Sid": "Stmt1397633323327",
                  "Effect": "Allow",
                  "Principal": {
                      "AWS": "arn:aws:iam::578681496768:user/blockchain-explorer-ci-cd"
                  },
                  "Action": "s3:*",
                  "Resource": [
                      "arn:aws:s3:::blockchain-explorer.service.ndau.tech",
                      "arn:aws:s3:::blockchain-explorer.service.ndau.tech/*"
                  ]
              }
          ]
      }
    ```
      - [ ] should be able to hit the url https://s3.amazonaws.com/blockchain-explorer.service.ndau.tech/index.html
  - configure bucket for static website hosting: Bucket | Properties | Static website hosting. Add index.html.
  - [optional] Turn on CloudTrail for object logging
  - [optional] configure any redirects forom other domains by adding additional buckets and setting them up to redirect to the real target bucket.

Turning on cloudfront for the S3 bucket would transfer the content to Amazon's CDN. Local users wouldn't notice much of a difference but global users would. Redirects are possible so multiple buckets wouldn't be required for `www.`. It also supports CORS. https://i.stack.imgur.com/eEDGb.png

## cloudfront

Create a distribution for each bucket directory you plan to host. For example, if you use staging and prod, this would require two distributions, each responding to a different domain name.

- [ ] Set the default root object to `index.html` or whatever is appropriate.
- [ ] The origin domain name and path should be the bucket and directory where index.html is.
- [ ] For react apps with push navigation routing, create a Custom Error Response for 404's that will route to index.html and respond with 200.

### SSL

docker run -it -v $(pwd):/etc/letsencrypt certbot/certbot certonly --manual --preferred-challenges dns --email=josh.anderson@oneiro.io -d explorer.service.ndau.tech -d explorer.staging.ndau.tech

If will ask you to agree to the terms and then it will ask you to set some TXT records on any domain names requested.

The domains are stored in /etc/letsencrypt/archive.

Go to AWS ACM and import the certs.

## Route53

  - Create an A record for the domain `blockchain-explorer.service.ndau.tech`.
    - paste the cloudfront domain name into the
  - Wait for dns to propagate and test
  - [ ] should be able to visit http://blockchain-explorer.service.ndau.tech/ and see the dummy index.html

## Circle CI

- Turn on blockchain-explorer repo in circleci
- Get the credentials csv file for IAM user `blockchain-explorer-ci-cd`.
  - and save them to 1password
  - add them as environment variables
    - `AWS_ACCESS_KEY_ID` - from csv
    - `AWS_SECRET_ACCESS_KEY` - from csv
    - `AWS_REGION` - us-east-1
  - [ ] When the environment variables are set in the command line `aws iam list-users` should work
- Environment variables for configuration can be stored in
    https://circleci.com/gh/oneiro-ndev/{repo}/edit#env-vars
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
