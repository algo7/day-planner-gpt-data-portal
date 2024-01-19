FROM gcr.io/distroless/static-debian12

WORKDIR /go/src/app

# The application has to be built first with CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w" main.go outside of docker
COPY main .

EXPOSE 3000

CMD [ "./main"]