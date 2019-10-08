/*
 * azure-resource-limits-exporter v0.1
 * 2019-10-08 Omnia Aurora Team / Stian Øvrevåge
 *
 * Uses Azure API to export current compute, storage and network resource 
 * quota usage and limits and exposes data in Prometheus format.
 *
 * Requirements:
 *   A Service Principal for accessing Azure API. Create one with:
 *     $ az ad sp create-for-rbac -n "go-collect-azure-resource-metrics"
 * 
 *   Use data to populate these environment variables:
 *     AZURE_TENANT_ID
 *     AZURE_CLIENT_ID (appId)
 *     AZURE_CLIENT_SECRET (password)
 * 
 *   Use `az account list --output table` to get Subscription ID:
 *     SUBSCRIPTION_ID
 * 
 *   Optionally set LISTEN_PORT, default is 5000
 * 
 */

package main

import (
  "context"
  "os"
  "log"
  "fmt"
  "net/http"
  "strings"
  "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
  "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
  "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-07-01/network"

  "github.com/Azure/go-autorest/autorest/azure/auth"
  
  "github.com/gorilla/handlers"
)

const (
	Version = "0.1"
	// ListenPort Default port for server to listen on unless specified in environment variable
	ListenPort = "5000"
)

// Simple counters for application metrics
var requestCount int64
var errorCount int64

func main() {
  fmt.Printf("Starting azure-resource-limits-exporter version %s\n", Version)

  http.HandleFunc("/metrics", Metrics)

	// See if listen_port environment variable is set
	port := os.Getenv("LISTEN_PORT")

	// Default port if none given
	if port == "" {
		port = ListenPort
	}

	fmt.Printf("Starting server on port %v\n", port)

	// Start server. Exit fatally on error
	log.Fatal(http.ListenAndServe(":"+port, handlers.CompressHandler(handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))))

}

func Metrics(w http.ResponseWriter, r *http.Request) {
  requestCount++

  hostname, _ := os.Hostname()

  // Valid label names: [a-zA-Z_][a-zA-Z0-9_]*
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	labels := map[string]interface{}{
		"host":      hostname,
		"pid":       os.Getpid(),
		"component": "azure-resource-limits-exporter",
		"version":   Version,
  }
  
  var labelsStr string

  for labelName, labelValue := range labels {
		labelsStr += fmt.Sprintf(`%s="%v",`, labelName, labelValue)
	}
	labelsStr = strings.Trim(labelsStr, ",")

	appMetrics := map[string]interface{}{
		"requests_total": requestCount,
		"errors_total":   errorCount,
	}

  for metric, value := range appMetrics {
		fmt.Fprintf(w, "%s{%s} %v\n", metric, labelsStr, value)
  }

  computeUsageClient := compute.NewUsageClient(os.Getenv("SUBSCRIPTION_ID"))
  storageUsageClient := storage.NewUsagesClient(os.Getenv("SUBSCRIPTION_ID"))
  networkUsageClient := network.NewUsagesClient(os.Getenv("SUBSCRIPTION_ID"))

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
    log.Print("Error while getting authorizer: ", err)
    errorCount++
    return
  }

  computeUsageClient.Authorizer = authorizer
  storageUsageClient.Authorizer = authorizer
  networkUsageClient.Authorizer = authorizer

  // Get Compute resources usage and limits
  for usage, err := computeUsageClient.ListComplete(context.Background(), os.Getenv("LOCATION")); usage.NotDone(); err = usage.NextWithContext(context.Background()) {
    if err != nil {
        log.Print("Error while traversing compute resource list: ", err)
        errorCount++
        return
    }
    i := usage.Value()
    fmt.Fprintf(w, "azure_resources_compute_%s_CurrentValue{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.CurrentValue)
    fmt.Fprintf(w, "azure_resources_compute_%s_Limit{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.Limit)
  }

  // Get Network resources usage and limits
  for usage, err := networkUsageClient.ListComplete(context.Background(), os.Getenv("LOCATION")); usage.NotDone(); err = usage.NextWithContext(context.Background()) {
    if err != nil {
        log.Print("Error while traversing network resource list: ", err)
        errorCount++
        return
    }
    i := usage.Value()
    fmt.Fprintf(w, "azure_resources_network_%s_CurrentValue{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.CurrentValue)
    fmt.Fprintf(w, "azure_resources_network_%s_Limit{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.Limit)
  }

  // Get Storage resources usage and limits
  storageUsageList, err := storageUsageClient.ListByLocation(context.Background(), os.Getenv("LOCATION"))
  if err != nil {
      log.Print("Error while traversing storage resource list: ", err)
      errorCount++
      return
  }

  for _, i := range *storageUsageList.Value {
    fmt.Fprintf(w, "azure_resources_storage_%s_CurrentValue{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.CurrentValue)
    fmt.Fprintf(w, "azure_resources_storage_%s_Limit{%s,location=\"%s\"} %v\n", *i.Name.Value, labelsStr, os.Getenv("LOCATION"), *i.Limit)
  }
  

}