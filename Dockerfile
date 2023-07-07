FROM golang:1.20 as build
WORKDIR /golang-lambda
# Copy dependencies list
COPY go.mod go.sum ./
# build
COPY src/*.go .
RUN go build -o main main.go
# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /golang-lambda/main /main
ENTRYPOINT [ "/main" ]