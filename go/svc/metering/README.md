<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## IDC Metering Service 
Metering is one the foundational service of any cloud offering. The core functional objectives behind metering service are to record the lifecycle events (create/update/delete) for cloud resources, persist those records with HA and provides an API interface to query these records. These records can be used for billing, analytics, capacity management, or archiving. 

IDC Metering Service provides reliable and highly available APIs to storing and quering metering records


### Local Dev Setup

To run the metering service outside of kind, first start a postgres database in docker:
```
docker run -d --name metering-db -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:latest
```

Now run the metering service:
```
go run cmd/metering/main.go --config dev-config.yaml
```

### Testing

You can use `grpcurl` to test out the grpc apis

These commands work with the "Local Dev Setup". To test the metering service running in kind, first setup port forwarding for the metering service:
```
$ kubectl -n idcs-system port-forward svc/metering 50051:80
```

1. Create some metering records

```
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d", "cloudAccountId": "123456789000",  "transactionId":"2022-12-16T15", "timestamp":"2022-12-16T15:34:00.000Z", "properties":{"instance":"small"} }' -plaintext localhost:50051 proto.MeteringService/Create
```
```
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d", "cloudAccountId": "123456789000",  "transactionId":"2022-12-16T16", "timestamp":"2022-12-16T16:34:00.000Z", "properties":{"instance":"small"} }' -plaintext localhost:50051 proto.MeteringService/Create
```
```
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d", "cloudAccountId": "123456789000",  "transactionId":"2022-12-16T17", "timestamp":"2022-12-16T17:34:00.000Z", "properties":{"instance":"small"} }' -plaintext localhost:50051 proto.MeteringService/Create
```

>> You can manually check in the database that the metering records are being created

2. Search metering records

For searching, you can pass multiple (optional) filtering properties as below:
```
transaction_id
cloud_account_id
resource_id
start_time
end_time
reported 
```

E.g.:
```
// To search records for resource_id=`3bc52389-da79-4947-a562-ab7a88c38e1d` 
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d"}' -plaintext localhost:50051 proto.MeteringService/Search


// To search records for resource_id=`3bc52389-da79-4947-a562-ab7a88c38e1d` and transaction_id=`2022-12-16T17`
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d", "transactionId":"2022-12-16T17"}' -plaintext localhost:50051 proto.MeteringService/Search
```

3. Find previous record:

This query would return the last inserted record for a given resource_id. The last insertion is currently determined by the timestamp in the record itself. So there could be a race condition, where some records with later timestamp are inserted before. And this case needs to be handled.

```
$ grpcurl -d '{"resourceId": "3bc52389-da79-4947-a562-ab7a88c38e1d", "id": "2"}' -plaintext localhost:50051 proto.MeteringService/FindPrevious
```
 
4. Update reported state for list of records

Update the reported state of records in the input list of id(s)

```
$ grpcurl -d '{"id":[2], "reported":true}' -plaintext localhost:50051 proto.MeteringService/Update

$ grpcurl -d '{"id":[1,3], "reported":true}' -plaintext localhost:50051 proto.MeteringService/Update
```