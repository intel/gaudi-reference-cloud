<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# MineralRiver library for Golang

<p align="center">
 <img src="https://github.com/intel-innersource/applications.analytics.elastic-search.mineral-river/blob/master/Capture.PNG?raw=true" alt="Mineral River" /></a>
</p>

This repo contains the package that needs to be imported for instrumentation of http traces for golang services
The package makes use of OTEL protocol to send traces data to OTEL collector hosted on IDC


 #### Downloading the package

* set env variable GOPRIVATE
set GOPRIVATE=*

* go get github.com/intel-innersource/applications.analytics.elastic-search.mineral-river

#### Proxy settings if go get doesnot  work from Intel network

set https_proxy=http://internal-placeholder.com:912


 #### Using  the package
 
import "github.com/intel-innersource/applications.analytics.elastic-search.mineral-river"

#### Setting config variables 

The following configuration is required to be set via environment variables or in a config.yml file in the parent directory.


* module: The name of the module which is instrumented
* loglevel: default to debug

Create a config file in your working directory of the golang service
Sample content in config file: 

```yml

module: "test-shruti-golang-service"
loglevel: "debug"
```
environment variables REQUIRED TO BE SET

**  Mandatory **
OTELEXPORTEROTLPENDPOINT is a mandatory field TO BE SET WITH environment variable

#### OTEL environment variables for windows
* set OTELEXPORTEROTLPENDPOINT=internal-placeholder.com:80
* set OTEL_SERVICE_NAME=idc-golang
* set OTEL_EXPORTER_OTLP_CERTIFICATE=/pathToCertFile"


#### OTEL environment variables for linux
* export OTELEXPORTEROTLPENDPOINT=internal-placeholder.com
* export OTEL_SERVICE_NAME=idc-golang
* export OTEL_EXPORTER_OTLP_CERTIFICATE=/pathToCertFile"






#### Usage in gin [link to sample](https://github.com/intel-innersource/applications.analytics.elastic-search.telemetry.mineral-river-examples/tree/main/golang/http)

* r.Use(otelgin.Middleware(serviceName))
This will start instrumenting traces for all API handlers

* r.Use(mineralriver.TraceMiddleware(authorizedUserUuidHeader))
This will add user userHeader attribute for every trace automatically
authorizedUserUuidHeader is the header that needs to be read for extracting user information

```go
package main


import (
  "context"
  "net/http"
  "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
  "github.com/intel-innersource/applications.analytics.elastic-search.mineral-river"
  "github.com/gin-gonic/gin"
  "log"
)
const authorizedUserUuidHeader = "x-jwt-user"
var serviceName string = "BookService"
func main() {

        mr := mineralriver.New()
        tp := mr.InitTracer()	
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

        r := gin.Default()
        r.Use(otelgin.Middleware(serviceName))
        r.Use(mineralriver.TraceMiddleware(authorizedUserUuidHeader))
        r.GET("/books", FindBooks)
        r.Run(":8090")
}
func FindBooks(c *gin.Context) {
        // Get model if exist
        // var book models.Book
        // if err := models.DB.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).First(&book).Error; err != nil {
        //      c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
        //      return
        // }

        c.JSON(http.StatusOK, gin.H{"data": "book"})
}
```


#### For setting custom resource options (Optional)


```go
	
	//optional to set custom resources
	customResourceOptions := make(map[string]string)
	customResourceOptions["custom1"]="custom1"
	customResourceOptions["custom2"]="custom2"
	

	mr := mineralriver.New(
			mineralriver.SetCustomResourceOptions(customResourceOptions),
	)

```


#### Adding custom intrumenation attributes/event at span level

```go
package kaas

import (
    "go.opentelemetry.io/otel/trace"
	"fmt"
	"time"
	"context"
	
)
var tracer trace.Tracer
// Init configures an OpenTelemetry exporter and trace provider
func Callkaas(ctx context.Context, tracer trace.Tracer) (int, error) {
	
	fmt.Println("K8 called")
	ctx, span := tracer.Start(ctx, "k8parent")
	time.Sleep(20 * time.Second)
	fmt.Println("K8 called done")
	
	defer span.End()
	return 1, nil
}



```
#### Usage in grpc [link to sample](https://github.com/intel-innersource/applications.analytics.elastic-search.telemetry.mineral-river-examples/tree/main/golang/grpc)


```
import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"github.com/intel-innersource/applications.analytics.elastic-search.mineral-river"
	"google.golang.org/grpc"
	.....
	.....
	
	
)

func main() {
	// main function code

	mr := mineralriver.New()
    	cleanup := mr.InitTracer()
    	defer cleanup(context.Background())

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	employeepc.RegisterEmployeeServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	
	
	
	//other lines of code
	
	}
	
	```


