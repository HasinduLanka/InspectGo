# Inspect Go

Web page inspection tool

## Project structure

This project is a stateless API and a static file web server written in Golang.

- `main.go` is the entry point of the program, providing both API and static file server for web app.
- `api` directory contains API endpoints.
- `pkg` directory contains exportable application logic decoupled from API.
- `frontend` directory contains a single page web app to consume the API and present the result to the user.

## Running the application

The server is written in platform-independent pure Go. It will run on all platforms consistently.

1. Install Golang compiler ( https://go.dev/doc/install ).
2. Clone this repository.
3. Install dependencies `go mod tidy`
4. Run the webserver `go run .`
5. Open the web page http://localhost:20000

## Development

Standard Go development environment is used.

- Visual Studio Code is preferred as the IDE.
- Code formatting is done using inbuilt `go fmt`
- Frontend is a SvelteJS web app written in TypeScript, which isn't covered in this documentation.
- NodeJS is required only if you do frontend development.
- Fork this repository, create a feature branch, and make a pull request to contribute.

## Testing

Go tests are configured for each functionality. They are used to test the library under `pkg` directly, decoupled from API endpoints. When making changes, you must update the tests accordingly.

Test all using:

```
go test -v ./...
```

## Deployment

The project is targeted to run on serverless platforms such as AWS Lambda, Google Cloud Functions, and Vercel. All these platforms support continuous deployment via a Git repository.

Deployment as a monolithic web server or a docker container is also supported out of the box. Direct source code deployment is recommended instead of pre-compiled binaries.

Currently, the project is deployed to Vercel with Go serverless functions: https://inspect-go.vercel.app/

However, this specific deployment has some limitations :

- Vercel Go functions do not support response streaming yet.
- Maximum request execution time is 10 seconds, because of the hobby plan.

## API endpoints

```
/api/inspect
```

- Usage 1

  - Method: POST
  - Request body: JSON with URL to be inspected `{url: "url"}`
  - Response: Single JSON that returns the analysis report.
  - The response could take more than 3 minutes when inspecting large web pages.

- Usage 2

  - Method: POST
  - Request header: `inspector-response-streamable = true`
  - Request body: JSON with URL to be inspected `{url: "url"}`
  - Response: Multiple responses streamed every 20 seconds, each containing a JSON that represents the analysis report state at that time.
  - The first report containing basic information about the webpage is returned immediately and presentable to the user. This is further explained [here](Task.md#link-analysis-taking-too-long).

- Response status codes:
  - 200: Success
  - 400: Bad request
  - 500: Internal server error
- Structure of the report object can be found [in `inspector.go` (Go)](pkg/inspector/inspector.go) and [`Types.ts` (TypeScript)](frontend/src/Types.ts)

## Task and challenges

The objective of the task and the challenges I faced while working on the project are [explained on Task.md](Task.md)
