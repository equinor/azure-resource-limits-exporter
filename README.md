# azure-resource-limits-exporter

Uses [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go)

Uses Azure API to export current compute, storage and network resource
quota usage and limits and exposes data in Prometheus format.

## Create an Azure Service Principal

    az ad sp create-for-rbac -n "go-collect-azure-resource-metrics"

## Deploy on Kubernetes with Helm version 3

Use the output from above as well as `az account list --output table` to populate the fields below. AZURE_CLIENT_ID=appId and AZURE_CLIENT_SECRET=password from `az`.

    helm repo add azure-resource-limits-exporter https://equinor.github.io/azure-resource-limits-exporter/charts/
    helm repo update

    helm upgrade --install azure-limits azure-resource-limits-exporter/azure-resource-limits-exporter \
        --set location=northeurope \
        --set azureCredentials.tenantId=xx \
        --set azureCredentials.clientId=xx \
        --set azureCredentials.clientSecret=xx \
        --set azureCredentials.subscriptionId=xx

## Run locally

    docker run -p 5000:5000 -e LOCATION=northeurope -e AZURE_TENANT_ID=xx -e AZURE_CLIENT_ID=xx -e AZURE_CLIENT_SECRET=xx -e SUBSCRIPTION_ID=xx stianovrevage/azure-resource-limits-exporter

## Development

### Build docker image

    docker build -t azure-resource-limits-exporter .

### Build locally

    go get -u github.com/Azure/azure-sdk-for-go/...
    go get github.com/gorilla/handlers
    go build main.go

### Package helm charts

    helm3 package charts/azure-resource-limits-exporter --destination ./charts
    helm3 repo index ./charts
