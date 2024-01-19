# day-planner-gpt-data-portal
Data integration solution designed to seamlessly connect data sources such as emails, news feeds, and calendar information for use with OpenAI's GPTs

## Table of Contents
- [day-planner-gpt-data-portal](#day-planner-gpt-data-portal)
  - [Table of Contents](#table-of-contents)
  - [How to run it](#how-to-run-it)
    - [Locally](#locally)
    - [Docker](#docker)
  - [Documentation](#documentation)


## How to run it
### Locally
1. Make sure the `credentials` folder exists in the root directory and contains the following files:
    - `google_credentials.json` - Google OAuth2 Configuration
    - `outlook_credentials.json` - Outlook OAuth2 Configuration
  - You can find the examples of these files in the `credentials` folder
2. Make sure you have an Redis instance running
    - If your Redis instance is not running on `localhost:6379`, you need to pass in the `REDIS_HOST` (with port) and `REDIS_PASS` environment variables
3. Run `go run main.go`

### Docker
1. Make sure the `credentials` folder exists in the root directory and contains the following files:
    - `google_credentials.json` - Google OAuth2 Configuration
    - `outlook_credentials.json` - Outlook OAuth2 Configuration
  - You can find the examples of these files in the `credentials` folder
2. Run `docker compose build`
3. Run `docker compose up`. Redis will be automatically started as a dependency

## Documentation
You can find the Swagger documentation on http://localhost:3000/docs

The `/outlook` and the `/google` routes are protected by the API key, which needs to be sent in the header as `X-API-KEY`. To obtain the initial API key, you need to first visit the `/apikey` endpoint in the browser and enter the initial password in the form to obtain the API key. The initial password is can be found in the startup logs of the application. The initial password is randomly generated on each startup. The initial password will expire the moment you obtain the API key. Subsequent visit to the `/apikey` endpoint will redirect you to the `/` or the homepage of the application. To call the protected endpoints listed above, you will need something like Postman to send the API key in the header.
