# day-planner-gpt-data-portal
Data integration solution designed to seamlessly connect data sources such as emails, news feeds, and calendar information to plan your day by using OpenAI's GPTs, which are custom versions of ChatGPT (read more here: https://openai.com/blog/introducing-gpts).

At this moment, the application only supports Google and Outlook integration. The API fetches the latest unread emails (past 2 days) from the user's inbox and returns the emails in JSON format that can be used in the custom "Actions" of the GPTs.

## Table of Contents
- [day-planner-gpt-data-portal](#day-planner-gpt-data-portal)
  - [Table of Contents](#table-of-contents)
  - [TODO](#todo)
  - [How to run it](#how-to-run-it)
    - [Locally](#locally)
    - [Docker](#docker)
    - [Makefile](#makefile)
  - [Documentation](#documentation)
  - [How to Interact with the API](#how-to-interact-with-the-api)
  - [Limitations](#limitations)


## TODO
- [ ] Declutter and version the API endpoints
- [ ] Add calendar integration
- [ ] Add news feed integration
- [ ] Write tests
- [ ] Kubernetes manifest



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

The `/outlook` and the `/google` routes are protected by the API key, which needs to be sent in the header as `X-API-KEY`. To obtain the initial API key, you need to first visit the `/apikey` endpoint in the browser and enter the initial password in the form to obtain the API key. The initial password is can be found in the startup logs of the application. The initial password is randomly generated on each startup. The initial password will expire the moment you obtain the API key. Subsequent visit to the `/apikey` endpoint will redirect you to the `/` or the homepage of the application. To call the protected endpoints listed above, you will need something like Postman to send the API key in the header.

## How to Interact with the API
1. Start the application
2. Check the startup logs for the initial password
3. Visit the `/apikey` endpoint in the browser and enter the initial password to obtain the API key
4. Visit the `/outlook/auth` endpoint in the browser to start the Outlook OAuth2 flow
5. Visit the `/google/auth` endpoint in the browser to start the Google OAuth2 flow
6. Visit the `/outlook` using Postman or any other API client and send the API key in the header as `X-API-KEY`
7. Visit the `/google` using Postman or any other API client and send the API key in the header as `X-API-KEY`

## Limitations
The application most likely not work with work or school accounts unless 2 requirements are met:
1. For Microsoft: Become a verified publisher
     - https://learn.microsoft.com/en-us/entra/identity-platform/publisher-verification-overview
2. For Google: Get your Oauth App verified
     -  https://support.google.com/cloud/answer/13463073?hl=en