<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## Using minio in tests

The type TestBucket provides a Bucket object that is backed by a MinIO container that it starts in Docker.
This allows integration tests to use a object storage without the user having to start a minio separately.

### Initialization

```go
	By("Starting test bucket")
	testBucket, err = buckets.NewTestBucket(ctx, bucketName)
	Expect(err).Should(Succeed())
```

### Cleanup

```go
    By("Stopping object storage")
    Expect(testBucket.Stop(ctx)).Should(Succeed())
```
