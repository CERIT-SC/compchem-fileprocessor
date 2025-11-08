# Compchem fileprocessor

This API is an implementation for the layer between compchem repository and argo workflows, it is written in go and connects to a postgresql database. The entire purpose of this API is to be configurable and generate workflow definitions from requests that come from compchem.

To run full application with the database use:
```
docker compose --profile run-backend up
```

To run just the db for development purposes use:
```
docker compose up
go run main.go
```

To configure workflows the application will make available use the workflows property in server-config.yaml:
```
workflows:
  - name: simulation-annotation
    mimetype: application/octet-stream
    extension: tpr
    processing-templates:
      - name: simulation-annotation-template
        template: simulation-annotation
```

For a file to be eligible for a workflow its extension must exactly match whats specified in the worklow AND the mimetype in files metadata needs to match.

A single workflow may have any amount of processing templates they are all run in parallel.

The API can be tested by using the commands below, some tests require docker to spin up postgres:
```
go test ./...
```
