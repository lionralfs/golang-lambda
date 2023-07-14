# local

```sh
docker-compose -f docker-compose-local.yaml up -d
```

```sh
go run main.go
```

```sh
docker build -t docker-image:test .
```

```sh
docker run -d -v ~/.aws-lambda-rie:/aws-lambda -p 9000:8080 --entrypoint /aws-lambda/aws-lambda-rie docker-image:test /main
```

```sh
curl "http://localhost:9000/2015-03-31/functions/function/invocations" --data-binary "@test-input.json"
```

# screw it, deploy

```sh
./deployment/deploy.sh <AWS-Account-ID>
```

# invoke via SNS

```sh
aws sns publish --topic-arn arn:aws:sns:eu-central-1:<AWS_ACCOUNT_ID>:topic-for-golang-lambda --message '{"id":"1"}' --message-attributes '{"MyCustomAttribute" : { "DataType":"String", "StringValue":"some-value"}}' --no-cli-pager
```