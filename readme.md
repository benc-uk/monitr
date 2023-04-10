# NanoMon - Monitoring Tool

NanoMon is a lightweight network and HTTP monitoring system, designed to be self hosted with Kubernetes (or other container based system). It is written in Go and based on the microservices pattern, as such it is decomposed into several discreet but interlinked components.

It also serves as a reference & learning app for microservices and is used by my Kubernetes workshop as the workload & application deployed in order to demonstrate Kubernetes concepts.

In a hurry? - Jump to the sections [running locally quick start](#local-dev-quick-start) or [deploying with Helm](#deploy-to-kubernetes-using-helm)

## Architecture

The architecture is fairly simple consisting of four application components and a database.

![architecture diagram](./etc/architecture.drawio.png)

- **API** - API provides the main interface for the frontend and any custom clients. It is RESTful and runs over HTTP(S). It connects directly to the MongoDB database.
- **Runner** - Monitor runs are executed from here (see [concepts](#concepts) below). It connects directly to the MongoDB database, and reads monitor configuration data, and saves back & stores result data.
- **Frontend** - The web interface is a SPA (single page application), consisting of a static set of HTML, JS etc which executes from the user's browser. It connects directly to the API, and is [developed using Alpine.js](https://alpinejs.dev/)
- **Frontend Host** - The static content host for the frontend app, which contains no business logic. This simply serves frontend application files HTML, JS and CSS files over HTTP. In addition it exposes a small configuration endpoint.
- **MongoDB** - Backend data store, this is a vanilla instance of MongoDB v4. External services which provide MongoDB compatibility (e.g. Azure Cosmos DB) will also work

## Concepts

NanoMon executes monitoring calls remotely over the network using standard protocols, it does this periodically on a set interval per monitor. The results & execution of a "run" is validated to determine the status or success. There are currently three statuses:

- **OK** &ndash; Indicates no problems, e.g. got a HTTP valid response.
- **Error** &ndash; Partial success as one or more rules failed, e.g. HTTP status code wasn't the expected value. See rules below.
- **Failed** &ndash; The monitor failed to run entirely e.g. connection, network or DNS failure.

### Monitor

A _monitor_ represents an instance of a given monitor _type_ (see below) with it's associated configuration. Common properties of all monitors include the interval on which they are run, and the target. The target is _type_ dependant but typically is a hostname or URL.

### Result

When a _monitor_ runs it generates a _result_. The _result_ as the name implies, holds the results of a run of a monitor, such as the timestamp, status, message and a value. The value of a _result_ is dependant on the type of _monitor_ however it most commonly represents the duration of the network request in milliseconds.

### Monitor Types

There are three types of monitor currently supported:

- **HTTP** &ndash; Makes HTTP(S) requests to a given URL and measures the response time.
- **Ping** &ndash; Carries out an ICMP ping to the target hostname or IP address.
- **TCP** &ndash; Attempts to create a TCP socket connection to the given hostname and port.

For more details see the complete monitor reference

## Repo Index

```text
📂
├── api             - API reference and spec, using TypeSpec
├── build           - Dockerfiles and supporting build artifacts
├── deploy  
│   ├── azure       - Deploy to Azure using Bicep
│   ├── helm        - Helm chart to deploy NanoMon
│   └── kubernetes  - Example Kubernetes manifests (No Helm)
├── etc             - Misc stuff :)
├── frontend        - The HTML/JS source for the frontend app
├── scripts         - Supporting helper bash scripts
├── services
│   ├── api         - Go source for the API service
│   ├── common      - Shared internal Go code
│   ├── frontend    - Go source for the frontend host server
│   └── runner      - Go source for the runner
└── tests           - Integration and performance tests
```

## Getting Started

This section provides options for quickly getting started running locally, or deploying to the cloud or Kubernetes.

### Local Dev Quick Start

This runs all the components directly on your dev machine. You will need to be using a Linux compatible system (e.g. WSL or a MacOS) with bash, make, Go, Docker & Node.js installed. You can try the provided [devcontainer](https://containers.dev/) if you don't have these pre-reqs.

- Run `make install-tools`
- Run `make run-db` (Note. Needs Docker)
- Open another terminal, run `make run-api`
- Open another terminal, run `make run-runner`
- Open another terminal, run `make run-frontend`
- The frontend should automatically open in your browser.

### Run Standalone Image

If you just want to try the app out, you can start the standalone image using Docker. This doesn't require you to have Go, Node.js etc

```bash
docker pull ghcr.io/benc-uk/nanomon-standalone:latest
docker run --rm -it -p 8000:8000 -p 8001:8001 ghcr.io/benc-uk/nanomon-standalone:latest
```

Then open the following URL http://localhost:8001/

### Deploy to Kubernetes using Helm 

See [Helm & Helm chart docs](./deploy/helm/)

### Deploy to Azure Container Apps with Bicep

See [Azure & Bicep docs](./deploy/azure/)

## Components & Services

### Runner

- Written in Go, [source code - /services/api](./services/api/)
- The runner requires a connection to MongoDB in order to start, it will exit if the connection fails.
- It will keep in sync with the `monitors` collection in the DB, it does this one of two ways:
  - Watch the collection using MongoDB change stream.
  - If change stream isn't supported, then it will poll the database and look for changes.
- The runner doesn't listen for network traffic or bind to any ports.

## Makefile reference

```text
help                 💬 This help message :)
install-tools        🔮 Install dev tools into project bin directory
lint                 🔍 Lint & format check only, sets exit code on error for CI
lint-fix             📝 Lint & format, attempts to fix errors & modify code
build                🔨 Build all binaries into project bin directory
images               📦 Build all container images
image-standalone     📦 Build the standalone image
push                 📤 Push all container images
run-api              🎯 Run API service locally with hot-reload
run-runner           🏃 Run monitor runner locally with hot-reload
run-frontend         🌐 Run frontend with dev HTTP server & hot-reload
run-db               🍃 Run MongoDB in container (needs Docker)
test                 🧪 Run all unit tests
test-api             🧪 Run API integration tests
generate             🤖 Generate OpenAPI spec using TypeSpec
clean                🧹 Clean up, remove dev data and files
```

## Configuration Reference

All three components (API, runner and frontend host) expect their configuration in the form of environmental variables. When running locally this is done via a `.env` file. Note. The `.env` file is not used when deploying or running the app elsewhere

### Variables used only by the frontend host:

| _Name_       | _Description_                                    | _Default_ |
| ------------ | ------------------------------------------------ | --------- |
| API_ENDPOINT | Instructs the frontend SPA where to find the API | /api      |

### Variables used by both API service and runner:

| _Name_        | _Description_                     | _Default_                 |
| ------------- | --------------------------------- | ------------------------- |
| MONGO_URI     | Connection string for MongoDB     | mongodb://localhost:27017 |
| MONGO_DB      | Database name to use              | nanomon                   |
| MONGO_TIMEOUT | Timeout for connecting to MongoDB | 30s                       |

### Variables used by both the API and frontend host:

| _Name_         | _Description_                                                   | _Default_   |
| -------------- | --------------------------------------------------------------- | ----------- |
| PORT           | TCP port for service to listen on                               | 8000 & 8001 |
| AUTH_CLIENT_ID | Used to enable authentication with given Azure AD app client ID. See auth section | _blank_     |
| AUTH_TENANT    | Set to Azure AD tenant ID if not using common                   | common      |

### Variables used only by the runner:

| _Name_              | _Description_                                                                                             | _Default_      |
| ------------------- | --------------------------------------------------------------------------------------------------------- | -------------- |
| ALERT_SMTP_PASSWORD | For alerting, the password for mail server                                                                | _blank_        |
| ALERT_SMTP_FROM     | From address for alerts, also used as the username                                                        | _blank_        |
| ALERT_SMTP_TO       | Address alert emails are sent to                                                                          | _blank_        |
| ALERT_SMTP_HOST     | SMTP hostname                                                                                             | smtp.gmail.com |
| ALERT_SMTP_PORT     | SMTP port                                                                                                 | 587            |
| ALERT_FAIL_COUNT    | How many time a monitor needs to fail in a row to trigger an alert                                        | 3              |
| POLLING_INTERVAL    | Only used when in polling mode                                                                            | 10s            |
| USE_POLLING         | Force polling mode, by default MongoDB change streams will be tried, and polling mode used if that fails. | false          |

## Scratch Notes Area

Using Cosmos DB
Add index for the `date` field to the results collection
`az cosmosdb mongodb collection update -a $COSMOS_ACCOUNT -g $COSMOS_RG -d nanomon -n results --idx '[{"key":{"keys":["_id"]}},{"key":{"keys":["date"]}}]'`

HELM REPO
helm repo add nanomon 'https://raw.githubusercontent.com/benc-uk/nanomon/main/deploy/helm'
