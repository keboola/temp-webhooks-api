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

Copy dependencies to the `/vendor` directory:
```
make fix
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

## Deployment

The service is deployed to Azure Container Instances to subscription `Keboola 2022-03 Hackathon` and resource group `zeleni_webhooks`.

The built image has to be pushed to repository `keboolawebhooks.azurecr.io`.

The deployment was created manually by command: `az container create --resource-group zeleni_webhooks --file deploy-aci.yaml`.

The update can be done by updating the docker image in the repository and restarting the service by command `az container restart --name keboolawebhooks --resource-group zeleni_webhooks`
