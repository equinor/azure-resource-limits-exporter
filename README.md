# azure-resource-limits-exporter

Uses [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go)

Uses Azure API to export current compute, storage and network resource 
quota usage and limits and exposes data in Prometheus format.

## Create an Azure Service Principal

    az ad sp create-for-rbac -n "go-collect-azure-resource-metrics"

## Deploy on Kubernetes

Use the output from above as well as `az account list --output table` to populate the fields below. AZURE_CLIENT_ID=appId and AZURE_CLIENT_SECRET=password from `az`.

    kubectl create secret generic azure-credentials \
        --from-literal=AZURE_TENANT_ID=xx \
        --from-literal=AZURE_CLIENT_ID=xx \
        --from-literal=AZURE_CLIENT_SECRET=xx \
        --from-literal=SUBSCRIPTION_ID=xx

Change the `LOCATION` environment variable in `kubernetes/deployment.yaml` if necessary.

    kubectl apply -f kubernetes/deployment.yaml

## Run locally

    docker run -p 5000:5000 -e LOCATION=northeurope -e AZURE_TENANT_ID=xx -e AZURE_CLIENT_ID=xx -e AZURE_CLIENT_SECRET=xx -e SUBSCRIPTION_ID=xx stianovrevage/azure-resource-limits-exporter

## Development

### Build docker image

    docker build -t azure-resource-limits-exporter .

### Build locally

    go get -u github.com/Azure/azure-sdk-for-go/...
    go get github.com/gorilla/handlers
    go build main.go

