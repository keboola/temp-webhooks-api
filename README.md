# Webhooks API

## Dev Image
```
docker-compose build dev
docker-compose run --rm -u "$(id -u):$(id -g)" dev bash
```

### Generate API
Generate API from `api/webhooks/design.go`. Run in `dev image`:
```
make generate-api
```

## Api Image

Build image:
```
docker-compose build api
```

Start server:
```
docker-compose run --rm -u "$(id -u):$(id -g)" -e KBC_STORAGE_API_HOST="connection.keboola.com" --service-ports api
```

Open:
`http://localhost:8888`
