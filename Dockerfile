FROM --platform=amd64 gcr.io/distroless/static-debian12
# Dedebug Image
# FROM --platform=amd64 golang:1.21.4


WORKDIR /go/src/app

# The application has to be built first with CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w" main.go outside of docker
COPY main swagger.json ./
COPY assets ./assets

EXPOSE 3000

CMD [ "./main"]