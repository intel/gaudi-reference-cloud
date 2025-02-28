## Configure NetBox Webhook for BMaaS Enrollment
For demonstration purposes, we are going to use a kind cluster.

### Deploy the Enrollment API Service
```shell
make kind-deploy-enrollment
```

### Deploy NetBox
```shell
make kind-deploy-netbox
```

### Configure the Webhook
Log in to NetBox at http://localhost:30001, and navigate to `Other -> Integrations -> Webhooks` to create a new webhook integration:
 - Select a model or device and the events to be received by the enrollment API service.
 - Add the enrollment API endpoint:
    ```
    http://bmaas-enrollment-kind.idcs-enrollment.svc.cluster.local:8970/api/v1/enroll
    ```
 - Set the body template to 
   ```json
   { "id": {{ data.id }}, "name": "{{ data.name }}"}
   ```
 - Uncheck the SSL verification.
   
You may also create the above webhook and a sample device object using this command:
```
make kind-populate-netbox-samples
```

### Verify the webhook configuration
Once you create or update a device, check the API service's logs for a POST request. A K8s Job should be created as a result to perform Ansible tasks.

### Enrollment API service request/response samples

```
Enroll a new node
-----------------

curl -v -X POST http://10.165.56.144:8970/api/v1/enroll -H 'content-type: application/json' -d '{ "id": 12345, "name": "test"}'
* Connected to 10.165.56.144 (10.165.56.144) port 8970 (#0)
> POST /api/v1/enroll HTTP/1.1
> Host: 10.165.56.144:8970
> User-Agent: curl/7.81.0
> Accept: */*
> content-type: application/json
> Content-Length: 30
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 201 Created
< Date: Thu, 15 Dec 2022 16:54:42 GMT
< Content-Length: 0

Enrollment job is already running
---------------------------------
curl -v -X POST http://10.165.56.144:8970/api/v1/enroll -H 'content-type: application/json' -d '{ "id": 12345, "name": "test"}'
* Connected to 10.165.56.144 (10.165.56.144) port 8970 (#0)
> POST /api/v1/enroll HTTP/1.1
> Host: 10.165.56.144:8970
> User-Agent: curl/7.81.0
> Accept: */*
> content-type: application/json
> Content-Length: 30
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 409 Conflict
< Date: Thu, 15 Dec 2022 16:54:45 GMT
< Content-Length: 0

Enrollment request doesn't have all the necessary data
------------------------------------------------------
curl -v -X POST http://10.165.56.144:8970/api/v1/enroll -H 'content-type: application/json' -d '{ "id": 12345 }'
* Connected to 10.165.56.144 (10.165.56.144) port 8970 (#0)
> POST /api/v1/enroll HTTP/1.1
> Host: 10.165.56.144:8970
> User-Agent: curl/7.81.0
> Accept: */*
> content-type: application/json
> Content-Length: 15
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 400 Bad Request
< Content-Type: application/json; charset=utf-8
< Date: Thu, 15 Dec 2022 16:54:56 GMT
< Content-Length: 118
<
* Connection #0 to host 10.165.56.144 left intact
{"error":"Key: 'BMaaSEnrollmentData.DeviceName' Error:Field validation for 'DeviceName' failed on the 'required' tag"}
```
