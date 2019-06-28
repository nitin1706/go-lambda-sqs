# go-lambda-sqs



To Package and upload the artifact to s3 bucket named golang-poc:
>> sam package   --template-file template.yaml   --output-template-file package.yml --s3-bucket golang-poc




To create a stack of resources:
>> aws cloudformation deploy --template-file package.yml --stack-name sam-test-deployment --capabilities CAPABILITY_NAMED_IAM
