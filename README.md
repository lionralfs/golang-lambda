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
