# Webhooks API

## Dev Image
```
docker-compose build dev
docker-compose run --rm -u "$(id -u):$(id -g)" dev bash
```

### Generate API
Generate API from `api/webhooks/design.go`:
```
make generate-api
```

## Api Image
```
docker-compose build api
```
