# day-planner-gpt-data-portal
Data integration solution designed to seamlessly connect data sources such as emails, news feeds, and calendar information to plan your day by using OpenAI's GPTs, which are custom versions of ChatGPT (read more here: https://openai.com/blog/introducing-gpts).

At this moment, the application only supports Google and Outlook integration. The API fetches the latest unread emails (past 2 days) from the user's inbox and returns the emails in JSON format that can be used in the custom "Actions" of the GPTs.

[![Build and Push Docker Image](https://github.com/algo7/day-planner-gpt-data-portal/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/algo7/day-planner-gpt-data-portal/actions/workflows/build.yml)

[![CodeQL](https://github.com/algo7/day-planner-gpt-data-portal/actions/workflows/codeql.yml/badge.svg)](https://github.com/algo7/day-planner-gpt-data-portal/actions/workflows/codeql.yml)
## Table of Contents
- [day-planner-gpt-data-portal](#day-planner-gpt-data-portal)
  - [Table of Contents](#table-of-contents)
  - [TODO](#todo)
  - [Requirements](#requirements)
  - [How to run it](#how-to-run-it)
    - [Locally](#locally)
    - [Docker](#docker)
    - [Makefile](#makefile)
  - [Documentation](#documentation)
  - [Note on the API Key](#note-on-the-api-key)
    - [Revoking the API Key](#revoking-the-api-key)
    - [Obtaining a new API Key after Expiration or Revocation](#obtaining-a-new-api-key-after-expiration-or-revocation)
  - [How to Interact with the API](#how-to-interact-with-the-api)
  - [Limitations](#limitations)


## TODO
- [ ] Add validation to `google_credentials.json` and `outlook_credentials.json` files
- [x] Declutter and version the API endpoints
- [ ] Implement OAuth Device Flow
- [ ] Add calendar integration
- [ ] Add news feed integration
- [ ] Write tests
- [ ] Kubernetes manifest

## Requirements
1. Go 1.21.4+
2. A running Redis instance
3. An API client of your choice (Postman, Insomnia, Thunder Client etc.)
4. A browser
5. Docker (optional)
6. Make (optional)

## How to run it
It is important that you follow the prerequisites here before running the application; otherwise, the application will not work.
1. Create a Google Cloud Project and Enable the Gmail API
    - Follow the instructions here: https://developers.google.com/gmail/api/quickstart/go all the way to the `Authorize credentials for a desktop application` section
    - Save the downloaded JSON file as `google_credentials.json` in the `credentials` folder
    - You can find the example of the file in the `credentials` folder as `google_credentials.example.json`

2. Register an app with the Microsoft Identity Platform
   - Follow the instructions here: https://docs.microsoft.com/en-us/graph/auth-register-app-v2
   - Configure the correct redirect URL in the `Authentication` section of the app registration
   - Create a client secret in the `Certificates & secrets` section of the app registration
   - Copy the client secret and ID then save them in the `outlook_credentials.json` file in the `credentials` folder
   - You can find the example of the file in the `credentials` folder as `outlook_credentials.example.json`

### Locally
1. Make sure the `credentials` folder exists in the root directory and contains the following files:
    - `google_credentials.json` - Google OAuth2 Configuration
    - `outlook_credentials.json` - Outlook OAuth2 Configuration
2. Make sure you have an Redis instance running
    - If your Redis instance is not running on `localhost:6379`, you need to pass in the `REDIS_HOST` (with port) and `REDIS_PASS` environment variables
3. Run `go run main.go`

### Docker
1. Make sure the `credentials` folder exists in the root directory and contains the following files:
    - `google_credentials.json` - Google OAuth2 Configuration
    - `outlook_credentials.json` - Outlook OAuth2 Configuration
2. Run `docker compose build`
3. Run `docker compose up`. Redis will be automatically started as a dependency

### Makefile
If you have `make` installed, you can simply run `make start` to build + run the application locally, or `make docker` to build + run the application in Docker.

## Documentation
You can find the Swagger documentation on http://localhost:3000/docs

## Note on the API Key
The `/v1/email/outlook` and the `/v1/email/google` routes are protected by the API key, which needs to be sent in the header as `X-API-KEY`. To obtain the initial API key, you need to first visit the `/v1/auth/internal/apikey` endpoint in the browser and enter the initial password in the form to obtain the API key. The initial password can be found in the startup logs of the application. The initial password is randomly generated on each startup, if and only if it has not been set. The initial password will get set to an empty string the moment you obtain the API key. Subsequent visit to the `/v1/auth/internal/apikey` endpoint will redirect you to the `/` or the homepage of the application. To call the protected endpoints listed above, you will need something like Postman to send the API key in the header.

### Revoking the API Key
The API key is stored in Redis and the TTL will get extended by 7 days everytime you call an protected endpoint. It will expire after 7 days of inactivity. If you want to revoke your active API key, you will have to manually delete it from Redis.

### Obtaining a new API Key after Expiration or Revocation
Since the initial password has been set to an empty string the 1st time you generated the API Key, to obtain a new one, you have 2 options:
1. Delete the `initial_password` key in Redis, restart the application (a new initial password will be generated), then go to the `/v1/auth/internal/apikey` endpoint in the browser to obtain a new API key
2. Set the `initial_password` key in Redis to a non-empty string and then go to the `/v1/auth/internal/apikey` endpoint in the browser to obtain a new API key using the new password

## How to Interact with the API
1. Start the application
2. Check the startup logs for the initial password
3. Visit the `/v1/auth/internal/apikey` endpoint in the browser and enter the initial password to obtain the API key
4. Call the `/v1/auth/oauth/auth` endpoint using an API client using the query parameter `provider` with the value `google` or `outlook`
   - The endpoint will present you with a link to the OAuth2 provider to complete the authentication flow 
   - Complete the authentication flow
5. Call the `/v1/email/outlook` using using an API client and send the API key in the header as `X-API-KEY` to get the latest unread emails from Outlook
6. Call the `/v1/email/google` using using an API client and send the API key in the header as `X-API-KEY` to get the latest unread emails from Gmail
7. Call the `/v1/auth/oauth/refresh` using an API client using the query parameter with the value `google` or `outlook` to refresh the token.
   - The endpoint will replace the token object in Redis with the new token object
   - The endpoint effectively revokes the old token and replaces it with a new one

## Limitations
The application will most likely not work with work or school accounts unless 2 requirements are met:
1. For Microsoft: Become a verified publisher
     - https://learn.microsoft.com/en-us/entra/identity-platform/publisher-verification-overview
2. For Google: Get your OAuth App verified
     -  https://support.google.com/cloud/answer/13463073?hl=en
