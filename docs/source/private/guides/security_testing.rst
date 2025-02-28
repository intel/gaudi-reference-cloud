.. _security_testing:

Security Testing
################

This document describes some methods for testing TLS with IDC.

TLS 1.3 Validation
*******************

Enable Debug Tools Pod
======================

Ensure that debugTools is enabled with the following steps.

#.  In the environments/dev-flex.yaml.gotmpl or environments/kind-singlecluster.yaml.gotmpl,
    set global.debugTools.enabled=true.

#.  In universe_deployer/environments/kind-singlecluster.json or other applicable file,
    add the debugTools component to the global section (not inside "region").

Deploy IDC
==========

Run deploy-all-in-kind-v2 by performing the steps in
:ref:`deploy_idc_core_services_in_local_kind_cluster`.

Create Client Certificate
=========================

Generate a client certificate and copy it to the debug-tools pod.

.. code:: bash

   eval `make show-export`
   make generate-vault-pki-test-cert
   kubectl cp ${SECRETS_DIR}/pki/testclient1.tgz -n idcs-system debug-tools:/root/testclient1.tgz

Extract client certificate inside the debug-tools pod.

.. code:: bash

   debug-tools:~#
   tar -xzvf testclient1.tgz

Test compute-api-server server
==============================

Ensure TLS 1.3 connection to compute-api-server with curl.

.. code:: bash

   debug-tools:~#
   curl -v \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   https://us-dev-1-compute-api-server.idcs-system.svc.cluster.local:8443

.. code:: console

   *   Trying 10.96.50.125:8443...
   * Connected to us-dev-1-compute-api-server.idcs-system.svc.cluster.local (10.96.50.125) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   *  CAfile: ca.pem
   *  CApath: none
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, CERT verify (15):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
   * ALPN: server accepted h2
   * Server certificate:
   *  subject: OU=us-dev-1-compute-api-server; CN=us-dev-1-compute-api-server.idcs-system.svc.cluster.local
   *  start date: Sep  4 20:39:26 2024 GMT
   *  expire date: Sep  5 20:39:54 2024 GMT
   *  subjectAltName: host "us-dev-1-compute-api-server.idcs-system.svc.cluster.local" matched cert's "us-dev-1-compute-api-server.idcs-system.svc.cluster.local"
   *  issuer: CN=Intel IDC CA 9c3fe80f us-dev-1-ca
   *  SSL certificate verify ok.
   * using HTTP/2
   * h2h3 [:method: GET]
   * h2h3 [:path: /]
   * h2h3 [:scheme: https]
   * h2h3 [:authority: us-dev-1-compute-api-server.idcs-system.svc.cluster.local:8443]
   * h2h3 [user-agent: curl/8.0.1]
   * h2h3 [accept: */*]
   * Using Stream ID: 1 (easy handle 0x7f69ff8dea90)
   > GET / HTTP/2
   > Host: us-dev-1-compute-api-server.idcs-system.svc.cluster.local:8443
   > user-agent: curl/8.0.1
   > accept: */*
   > 
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   < HTTP/2 415 
   < content-type: application/grpc
   < grpc-status: 3
   < grpc-message: invalid gRPC request content-type ""
   < 
   * Connection #0 to host us-dev-1-compute-api-server.idcs-system.svc.cluster.local left intact

Ensure failure of TLS 1.2 connection to compute-api-server with curl.

.. code:: bash

   debug-tools:~#
   curl -v \
   --tls-max 1.2 \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   https://us-dev-1-compute-api-server.idcs-system.svc.cluster.local:8443

.. code:: console

   *   Trying 10.96.50.125:8443...
   * Connected to us-dev-1-compute-api-server.idcs-system.svc.cluster.local (10.96.50.125) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.2 (OUT), TLS handshake, Client hello (1):
   *  CAfile: ca.pem
   *  CApath: none
   * TLSv1.2 (IN), TLS alert, protocol version (582):
   * OpenSSL/3.1.0: error:0A00042E:SSL routines::tlsv1 alert protocol version
   * Closing connection 0
   curl: (35) OpenSSL/3.1.0: error:0A00042E:SSL routines::tlsv1 alert protocol version

Test grpc-proxy-internal server
================================

Ensure TLS 1.3 connection to grpc-proxy-internal with curl.
We must use ``-connect-to`` because the certificate CN is different from the service name.

.. code:: bash

   debug-tools:~#
   curl -v \
   --tlsv1.3 \
   --tls-max 1.3 \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   --connect-to dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:8443:us-dev-1-grpc-proxy-internal.idcs-system.svc.cluster.local:8443 \
   https://dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:8443

.. code:: console

   * Connecting to hostname: us-dev-1-grpc-proxy-internal.idcs-system.svc.cluster.local
   * Connecting to port: 8443
   *   Trying 10.96.181.46:8443...
   * Connected to (nil) (10.96.181.46) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   *  CAfile: ca.pem
   *  CApath: none
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, CERT verify (15):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
   * ALPN: server did not agree on a protocol. Uses default.
   * Server certificate:
   *  subject: OU=us-dev-1-grpc-proxy-internal; CN=dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local
   *  start date: Sep  5 03:19:02 2024 GMT
   *  expire date: Sep  6 03:19:32 2024 GMT
   *  subjectAltName: host "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local" matched cert's "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local"
   *  issuer: CN=Intel IDC CA 9c3fe80f us-dev-1-ca
   *  SSL certificate verify ok.
   * using HTTP/1.x
   > GET / HTTP/1.1
   > Host: dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:8443
   > User-Agent: curl/8.0.1
   > Accept: */*
   > 
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * old SSL session ID is stale, removing
   < HTTP/1.1 404 Not Found
   < date: Thu, 05 Sep 2024 03:37:26 GMT
   < server: envoy
   < content-length: 0
   < 
   * Connection #0 to host (nil) left intact

Ensure failure of TLS 1.2 connection to grpc-proxy-internal with curl.

.. code:: bash

   debug-tools:~#
   curl -v \
   --tlsv1.2 \
   --tls-max 1.2 \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   --connect-to dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:8443:us-dev-1-grpc-proxy-internal.idcs-system.svc.cluster.local:8443 \
   https://dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:8443

.. code:: console

   * Connecting to hostname: us-dev-1-grpc-proxy-internal.idcs-system.svc.cluster.local
   * Connecting to port: 8443
   *   Trying 10.96.181.46:8443...
   * Connected to (nil) (10.96.181.46) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.2 (OUT), TLS handshake, Client hello (1):
   *  CAfile: ca.pem
   *  CApath: none
   * TLSv1.2 (IN), TLS alert, protocol version (582):
   * OpenSSL/3.1.0: error:0A00042E:SSL routines::tlsv1 alert protocol version
   * Closing connection 0
   curl: (35) OpenSSL/3.1.0: error:0A00042E:SSL routines::tlsv1 alert protocol version

Ensure grpcurl to grpc-proxy-internal via ingress controller.

.. code:: bash

   debug-tools:~#
   grpcurl \
   -v \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:443 \
   list

.. code:: none

   grpc.reflection.v1.ServerReflection
   grpc.reflection.v1alpha.ServerReflection
   proto.BucketLifecyclePrivateService
   proto.BucketUserPrivateService
   proto.DpaiAirflowConfService   
   ...

Kind Multicluster Testing
*************************

Test preparation
================

.. code:: bash

   export IDC_ENV=kind-multicluster
   make deploy-all-in-kind
   make generate-vault-pki-test-cert

Attempt to access global grpc-proxy-internal without client certificate succeeds
=================================================================================

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   no_proxy=.local curl -v --cacert local/secrets/pki/testclient1/ca.pem \
   https://dev.grpcapi.cloud.intel.com.kind.local

.. code:: none

   * Uses proxy env variable no_proxy == '.local'
   *   Trying 127.0.2.2:443...
   * TCP_NODELAY set
   * Connected to dev.grpcapi.cloud.intel.com.kind.local (127.0.2.2) port 443 (#0)
   * ALPN, offering h2
   * ALPN, offering http/1.1
   * successfully set certificate verify locations:
   *   CAfile: local/secrets/pki/kind-multicluster-root-ca/ca.pem
     CApath: /etc/ssl/certs
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
   * ALPN, server accepted to use h2
   * Server certificate:
   *  subject: OU=grpc-proxy-internal; CN=dev.grpcapi.cloud.intel.com.kind.local
   *  start date: Jun  5 22:37:43 2023 GMT
   *  expire date: Jun  5 23:38:12 2023 GMT
   *  subjectAltName: host "dev.grpcapi.cloud.intel.com.kind.local" matched cert's "dev.grpcapi.cloud.intel.com.kind.local"
   *  issuer: CN=idcs-system.svc.cluster.local
   *  SSL certificate verify ok.
   * Using HTTP2, server supports multi-use
   * Connection state changed (HTTP/2 confirmed)
   * Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
   * Using Stream ID: 1 (easy handle 0x55a1464cf300)
   > GET / HTTP/2
   > Host: dev.grpcapi.cloud.intel.com.kind.local
   > user-agent: curl/7.68.0
   > accept: */*
   >
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * old SSL session ID is stale, removing
   * Connection state changed (MAX_CONCURRENT_STREAMS == 128)!
   < HTTP/2 404
   < date: Mon, 05 Jun 2023 22:52:01 GMT
   < content-length: 0
   < strict-transport-security: max-age=15724800; includeSubDomains
   <
   * Connection #0 to host dev.grpcapi.cloud.intel.com.kind.local left intact

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   grpcurl --cacert local/secrets/pki/testclient1/ca.pem \
   dev.grpcapi.cloud.intel.com.kind.local:443 list

.. code:: console

   grpc.reflection.v1alpha.ServerReflection
   proto.BillingAccountService
   proto.BillingCouponService
   proto.BillingCreditService
   proto.BillingDriverUsageService
   proto.BillingInvoiceService
   proto.BillingOptionService
   proto.BillingProductCatalogSyncService
   proto.BillingRateService
   proto.BillingUsageService
   proto.CloudAccountEnrollService
   proto.CloudAccountMemberService
   proto.CloudAccountService
   proto.ConsoleInvoiceService
   proto.MeteringService
   proto.PaymentService
   proto.ProductCatalogService
   proto.ProductVendorService

Attempt to access regional us-dev-1-grpc-proxy-internal without client certificate fails
=========================================================================================

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   no_proxy=.local curl -v --cacert local/secrets/pki/testclient1/ca.pem \
   https://dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443

.. code:: none

   * Uses proxy env variable no_proxy == '.local'
   *   Trying 127.0.2.2:10443...
   * TCP_NODELAY set
   * Connected to dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local (127.0.2.2) port 10443 (#0)
   * ALPN, offering h2
   * ALPN, offering http/1.1
   * successfully set certificate verify locations:
   *   CAfile: local/secrets/pki/testclient1/ca.pem
     CApath: /etc/ssl/certs
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
   * ALPN, server did not agree to a protocol
   * Server certificate:
   *  subject: OU=us-dev-1-grpc-proxy-internal; CN=dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local
   *  start date: Jun  6 00:02:42 2023 GMT
   *  expire date: Jun  6 01:03:12 2023 GMT
   *  subjectAltName: host "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local" matched cert's "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local"
   *  issuer: CN=idcs-system.svc.cluster.local
   *  SSL certificate verify ok.
   > GET / HTTP/1.1
   > Host: dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443
   > User-Agent: curl/7.68.0
   > Accept: */*
   >
   * TLSv1.3 (IN), TLS alert, unknown (628):
   * OpenSSL SSL_read: error:1409445C:SSL routines:ssl3_read_bytes:tlsv13 alert certificate required, errno 0
   * Closing connection 0
   curl: (56) OpenSSL SSL_read: error:1409445C:SSL routines:ssl3_read_bytes:tlsv13 alert certificate required, errno 0

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   grpcurl --cacert local/secrets/pki/testclient1/ca.pem \
   dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443 list

.. code:: none

   Failed to dial target host "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443": context deadline exceeded

Attempt to access regional us-dev-1-grpc-proxy-internal with client certificate succeeds
========================================================================================

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   no_proxy=.local curl -v \
   --cacert local/secrets/pki/testclient1/ca.pem \
   --cert local/secrets/pki/testclient1/cert.pem \
   --key local/secrets/pki/testclient1/cert.key \
   https://dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443

.. code:: console

   > --cacert local/secrets/pki/testclient1/ca.pem \
   > --cert local/secrets/pki/testclient1/cert.pem \
   > --key local/secrets/pki/testclient1/cert.key \
   > https://dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443
   * Uses proxy env variable no_proxy == '.local'
   *   Trying 127.0.2.2:10443...
   * TCP_NODELAY set
   * Connected to dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local (127.0.2.2) port 10443 (#0)
   * ALPN, offering h2
   * ALPN, offering http/1.1
   * successfully set certificate verify locations:
   *   CAfile: local/secrets/pki/testclient1/ca.pem
     CApath: /etc/ssl/certs
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, CERT verify (15):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
   * ALPN, server did not agree to a protocol
   * Server certificate:
   *  subject: OU=us-dev-1-grpc-proxy-internal; CN=dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local
   *  start date: Jun  6 01:01:52 2023 GMT
   *  expire date: Jun  6 02:02:22 2023 GMT
   *  subjectAltName: host "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local" matched cert's "dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local"
   *  issuer: CN=idcs-system.svc.cluster.local
   *  SSL certificate verify ok.
   > GET / HTTP/1.1
   > Host: dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443
   > User-Agent: curl/7.68.0
   > Accept: */*
   >
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   * old SSL session ID is stale, removing
   * Mark bundle as not supporting multiuse
   < HTTP/1.1 404 Not Found
   < date: Tue, 06 Jun 2023 01:21:09 GMT
   < server: envoy
   < content-length: 0
   <
   * Connection #0 to host dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local left intact

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   grpcurl \
   --cacert local/secrets/pki/testclient1/ca.pem \
   --cert local/secrets/pki/testclient1/cert.pem \
   --key local/secrets/pki/testclient1/cert.key \
   dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:10443 list

.. code:: console

   grpc.reflection.v1alpha.ServerReflection
   proto.InstancePrivateService
   proto.InstanceSchedulingService
   proto.InstanceService
   proto.InstanceTypeService
   proto.IpResourceManagerService
   proto.MachineImageService
   proto.SshPublicKeyService
   proto.VNetPrivateService
   proto.VNetService

Attempt to access AZ us-dev3-1a-vm-instance-scheduler with client certificate succeeds
=======================================================================================

.. code:: bash

   claudiof@claudiof-ws:~/frameworks.cloud.devcloud.services.idc$
   make generate-vault-pki-test-cert
   kubectl cp local/secrets/${IDC_ENV}/pki/testclient1.tgz -n idcs-system debug-tools:/root/testclient1.tgz

.. code:: bash

   debug-tools:~#
   tar -xzvf testclient1.tgz
   curl -vk \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   https://us-dev3-1a-vm-instance-scheduler:8443

.. code:: console

   *   Trying 10.43.209.117:8443...
   * Connected to us-dev3-1a-vm-instance-scheduler (10.43.209.117) port 8443 (#0)

.. code:: bash

   debug-tools:~#
   tar -xzvf testclient1.tgz
   echo "10.43.209.117  internal-placeholder.com" >> /etc/hosts
   curl -vk \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   https://internal-placeholder.com:8443

.. code:: console

   *   Trying 10.43.209.117:8443...
   * Connected to internal-placeholder.com (10.43.209.117) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, CERT verify (15):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
   * ALPN: server accepted h2
   * Server certificate:
   *  subject: OU=us-dev3-1a-vm-instance-scheduler; CN=internal-placeholder.com
   *  start date: Jun  7 01:19:58 2023 GMT
   *  expire date: Jun  7 02:20:28 2023 GMT
   *  issuer: CN=idcs-system.svc.cluster.local
   *  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
   * using HTTP/2
   * h2h3 [:method: GET]
   * h2h3 [:path: /]
   * h2h3 [:scheme: https]
   * h2h3 [:authority: internal-placeholder.com:8443]
   * h2h3 [user-agent: curl/8.0.1]
   * h2h3 [accept: */*]
   * Using Stream ID: 1 (easy handle 0x7facd2ba9a90)
   > GET / HTTP/2
   > Host: internal-placeholder.com:8443
   > user-agent: curl/8.0.1
   > accept: */*
   >
   * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
   < HTTP/2 415
   < content-type: application/grpc
   < grpc-status: 3
   < grpc-message: invalid gRPC request content-type ""
   <
   * Connection #0 to host internal-placeholder.com left intact

.. code:: console

   us-dev3-1a-debug-tools:~# openssl s_client -CAfile ca.pem internal-placeholder.com:8443
   CONNECTED(00000003)
   depth=2 CN = Intel IDC Private CA 0157a516
   verify return:1
   depth=1 CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
   verify return:1
   depth=0 OU = us-dev3-1a-vm-instance-scheduler, CN = internal-placeholder.com
   verify return:1
   ---
   Certificate chain
    0 s:OU = us-dev3-1a-vm-instance-scheduler, CN = internal-placeholder.com
      i:CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
      a:PKEY: rsaEncryption, 2048 (bit); sigalg: RSA-SHA256
      v:NotBefore: Jun  7 02:20:10 2023 GMT; NotAfter: Jun  7 03:20:40 2023 GMT
    1 s:CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
      i:CN = Intel IDC Private CA 0157a516
      a:PKEY: rsaEncryption, 2048 (bit); sigalg: RSA-SHA256
      v:NotBefore: Jun  7 02:18:16 2023 GMT; NotAfter: Jun  6 02:18:46 2024 GMT
   ---
   Server certificate
   -----BEGIN CERTIFICATE-----
   MIIEhDCCA2ygAwIBAgIUeOM/VoBDayz/EgDG6nmlv+5JEI4wDQYJKoZIhvcNAQEL
   BQAwNjE0MDIGA1UEAxMrSW50ZWwgSURDIFByaXZhdGUgQ0EgMDE1N2E1MTYgdXMt
   ZGV2My0xYS1jYTAeFw0yMzA2MDcwMjIwMTBaFw0yMzA2MDcwMzIwNDBaMGoxKTAn
   BgNVBAsTIHVzLWRldjMtMWEtdm0taW5zdGFuY2Utc2NoZWR1bGVyMT0wOwYDVQQD
   EzRkZXYzLWNvbXB1dGUtdXMtZGV2My0xYS1ncnBjYXBpLWNsb3VkLmVnbGIuaW50
   ZWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwjcP1quM75ej
   EIjSDscC+xpIT4Kj6PvKQQcTcUB3wtxy77k6HnRwLUraPQvL+wat/nhig/oq9HI+
   Wpr6miXnPmZ+jyKtedj6OVHfQjuB78HxvBY+oPdmFyCreesLhydjSRXv4JCPco8L
   2JieolnflZP1zmqgdQufkJFRhiW86gUTSDIKpEqxn7ADMhxN+hYZo2nK6HIabJlR
   MHA/3+vKbsHFujWAL934X7yOUXPv1GW8+B/rHrEOp0Kb6QKdUqVuZRRV3Q8YkMJz
   i+KaMqbgDGBwBMGEgH++SIzYYk8E6PuYwIdNcjCDcT4LnaxKlWrNVuNZrNbLMTjE
   Lrk1lf+tEQIDAQABo4IBVDCCAVAwDgYDVR0PAQH/BAQDAgOoMB0GA1UdJQQWMBQG
   CCsGAQUFBwMBBggrBgEFBQcDAjAdBgNVHQ4EFgQUfbar9ngyv/aawqkSCfxDFL08
   vmAwHwYDVR0jBBgwFoAUYuLQAplGYiv23rs/MmU83iAdtnUwUwYIKwYBBQUHAQEE
   RzBFMEMGCCsGAQUFBzAChjdodHRwczovL2lkY3ZhdWx0ZGV2LmVnbGIuaW50ZWwu
   Y29tLy92MS91cy1kZXYzLTFhLWNhL2NhMD8GA1UdEQQ4MDaCNGRldjMtY29tcHV0
   ZS11cy1kZXYzLTFhLWdycGNhcGktY2xvdWQuZWdsYi5pbnRlbC5jb20wSQYDVR0f
   BEIwQDA+oDygOoY4aHR0cHM6Ly9pZGN2YXVsdGRldi5lZ2xiLmludGVsLmNvbS8v
   djEvdXMtZGV2My0xYS1jYS9jcmwwDQYJKoZIhvcNAQELBQADggEBAHnjYdJ8zA/b
   QPlk626jw2RQH4Jn/D2SXju2Zb7IqTQBAykptm4dV6U1cFUYgtWkLjH9FmXalgYY
   ZHZkeHu+ZI81b8pGW5Hsj/yLQzefBq1GyDl++q+2FOvJ5C5qJL7aEIoS2hXKW0hI
   yuHDoc6NWt6DmlZJhSknJQWu7jk66wt5gk8sG2Wn/UpZanQK7KLiI+1v/fWSwuTe
   qZctcPvlj8FiTzlMWl2XDyGX3d42+3GU6eJLL1r58j9wbJSKDb7WBamThQ32oLT1
   qfmw1kIvqN2Hu+g3iStxFjN8ZX0BzrrT2Gtxrd4bSHvFO65tOZVg9FwUoQe0fM5c
   UlXsrPfYWs0=
   -----END CERTIFICATE-----
   subject=OU = us-dev3-1a-vm-instance-scheduler, CN = internal-placeholder.com
   issuer=CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
   ---
   Acceptable client certificate CA names
   CN = Intel IDC Private CA 0157a516
   Requested Signature Algorithms: RSA-PSS+SHA256:ECDSA+SHA256:Ed25519:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA+SHA256:RSA+SHA384:RSA+SHA512:ECDSA+SHA384:ECDSA+SHA512:RSA+SHA1:ECDSA+SHA1
   Shared Requested Signature Algorithms: RSA-PSS+SHA256:ECDSA+SHA256:Ed25519:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA+SHA256:RSA+SHA384:RSA+SHA512:ECDSA+SHA384:ECDSA+SHA512
   Peer signing digest: SHA256
   Peer signature type: RSA-PSS
   Server Temp Key: X25519, 253 bits
   ---
   SSL handshake has read 2834 bytes and written 448 bytes
   Verification: OK
   ---
   New, TLSv1.3, Cipher is TLS_AES_128_GCM_SHA256
   Server public key is 2048 bit
   This TLS version forbids renegotiation.
   No ALPN negotiated
   Early data was not sent
   Verify return code: 0 (ok)
   ---
   488B30D71C7F0000:error:0A000412:SSL routines:ssl3_read_bytes:sslv3 alert bad certificate:ssl/record/rec_layer_s3.c:1586:SSL alert number 42

.. code:: bash

   us-dev3-1a-debug-tools:~# \
   ping internal-placeholder.com
   PING internal-placeholder.com (10.43.209.117) 56(84) bytes of data.
   openssl s_client \
   -connect internal-placeholder.com:8443 \
   -state -quiet \
   -CAfile ca.pem
   48DB6568E37F0000:error:0A000412:SSL routines:ssl3_read_bytes:sslv3 alert bad certificate:ssl/record/rec_layer_s3.c:1586:SSL alert number 42

.. code:: bash

   us-dev3-1a-debug-tools:~# \
   cat cert.pem >> ca.pem
   # Remove testclient1 leaf cert from ca.pem. It should have root + intermediate only.
   openssl s_client \
   -connect internal-placeholder.com:8443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

Connection successful.

.. code:: console

   SSL_connect:before SSL initialization
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS read server hello
   SSL_connect:TLSv1.3 read encrypted extensions
   SSL_connect:SSLv3/TLS read server certificate request
   depth=2 CN = Intel IDC Private CA 0157a516
   verify return:1
   depth=1 CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
   verify return:1
   depth=0 OU = us-dev3-1a-vm-instance-scheduler, CN = internal-placeholder.com
   verify return:1
   SSL_connect:SSLv3/TLS read server certificate
   SSL_connect:TLSv1.3 read server certificate verify
   SSL_connect:SSLv3/TLS read finished
   SSL_connect:SSLv3/TLS write change cipher spec
   SSL_connect:SSLv3/TLS write client certificate
   SSL_connect:SSLv3/TLS write certificate verify
   SSL_connect:SSLv3/TLS write finished
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSLv3/TLS read server session ticket
   @

Now connect through EGLB.

.. code:: bash

   us-dev3-1a-debug-tools:~# \
   # Remove line from /etc/hosts
   ping internal-placeholder.com
   PING internal-placeholder.com (100.64.16.101) 56(84) bytes of data.
   openssl s_client \
   -connect internal-placeholder.com:443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

Connection *TO NGINX INGRESS* successful but this did not connect to
vm-instance-scheduler. The problem is that NGINX is not in SSL
passthrough mode!

.. code:: console

   SSL_connect:before SSL initialization
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS read server hello
   depth=2 C = US, ST = New Jersey, L = Jersey City, O = The USERTRUST Network, CN = USERTrust RSA Certification Authority
   verify error:num=19:self-signed certificate in certificate chain
   verify return:1
   depth=2 C = US, ST = New Jersey, L = Jersey City, O = The USERTRUST Network, CN = USERTrust RSA Certification Authority
   verify return:1
   depth=1 C = GB, ST = Greater Manchester, L = Salford, O = Sectigo Limited, CN = Sectigo RSA Organization Validation Secure Server CA
   verify return:1
   depth=0 C = US, ST = California, O = Intel Corporation, CN = *.eglb.intel.com
   verify return:1
   SSL_connect:SSLv3/TLS read server certificate
   SSL_connect:SSLv3/TLS read server key exchange
   SSL_connect:SSLv3/TLS read server done
   SSL_connect:SSLv3/TLS write client key exchange
   SSL_connect:SSLv3/TLS write change cipher spec
   SSL_connect:SSLv3/TLS write finished
   SSL_connect:SSLv3/TLS write finished
   SSL_connect:SSLv3/TLS read change cipher spec
   SSL_connect:SSLv3/TLS read finished

Fix NGINX SSL passthrough.

.. code:: console

   kubectl apply -f deployment/rke2/root/var/lib/rancher/rke2/server/manifests/rke2-ingress-nginx-config.yaml

.. code:: bash

   openssl s_client \
   -connect internal-placeholder.com:443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

Connection successful.

.. code:: console

   SSL_connect:before SSL initialization
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS read server hello
   SSL_connect:TLSv1.3 read encrypted extensions
   SSL_connect:SSLv3/TLS read server certificate request
   depth=2 CN = Intel IDC Private CA 0157a516
   verify return:1
   depth=1 CN = Intel IDC Private CA 0157a516 us-dev3-1a-ca
   verify return:1
   depth=0 OU = us-dev3-1a-vm-instance-scheduler, CN = internal-placeholder.com
   verify return:1
   SSL_connect:SSLv3/TLS read server certificate
   SSL_connect:TLSv1.3 read server certificate verify
   SSL_connect:SSLv3/TLS read finished
   SSL_connect:SSLv3/TLS write change cipher spec
   SSL_connect:SSLv3/TLS write client certificate
   SSL_connect:SSLv3/TLS write certificate verify
   SSL_connect:SSLv3/TLS write finished
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSLv3/TLS read server session ticket

.. code:: bash

   us-dev3-1a-debug-tools:~# \
   openssl s_client \
   -connect internal-placeholder.com:443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

Misc. Troubleshooting
*********************

.. code:: bash

   claudiof@claudiof-ws:~/frameworks.cloud.devcloud.services.idc$
   make generate-vault-pki-test-cert
   kubectl cp local/secrets/pki/testclient1.tgz -n idcs-system debug-tools:/root/testclient1.tgz

.. code:: bash

   debug-tools:~#
   tar -xzvf testclient1.tgz
   openssl s_client \
   -connect us-dev-1a-vm-instance-scheduler:8443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

.. code:: bash

   ~/frameworks.cloud.devcloud.services.idc$
   no_proxy=.local curl -v \
   --cacert local/secrets/pki/testclient1/ca.pem \
   --cert local/secrets/pki/testclient1/cert.pem \
   --key local/secrets/pki/testclient1/cert.key \
   https://dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:20443

.. code:: bash

   debug-tools:~#
   curl -v \
   --cacert ca.pem \
   https://dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:443
   curl: (56) OpenSSL SSL_read: OpenSSL/3.1.0: error:0A000412:SSL routines::sslv3 alert bad certificate, errno 0

.. code:: bash

   debug-tools:~#
   curl -v \
   --cacert ca.pem \
   --cert cert.pem \
   --key cert.key \
   https://dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:443
   HTTP/2 415

.. code:: bash

   debug-tools:~#
   openssl s_client \
   -connect dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key
   489B5C19F97F0000:error:0A000412:SSL routines:ssl3_read_bytes:sslv3 alert bad certificate:ssl/record/rec_layer_s3.c:1586:SSL alert number 42

.. code:: bash

   debug-tools:~#
   cat cert.pem >> ca.pem
   # Remove testclient1 leaf cert from ca.pem. It should have root + intermediate only.
   openssl s_client \
   -connect dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:443 \
   -state -quiet \
   -CAfile ca.pem \
   -cert cert.pem \
   -key cert.key

TLS connection successful.

.. code:: console

   SSL_connect:before SSL initialization
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS write client hello
   SSL_connect:SSLv3/TLS read server hello
   SSL_connect:TLSv1.3 read encrypted extensions
   SSL_connect:SSLv3/TLS read server certificate request
   depth=2 CN = Intel IDC Private CA 83fc6367
   verify return:1
   depth=1 CN = Intel IDC Private CA 83fc6367 us-dev-1a-ca
   verify return:1
   depth=0 OU = us-dev-1a-vm-instance-scheduler, CN = dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local
   verify return:1
   SSL_connect:SSLv3/TLS read server certificate
   SSL_connect:TLSv1.3 read server certificate verify
   SSL_connect:SSLv3/TLS read finished
   SSL_connect:SSLv3/TLS write change cipher spec
   SSL_connect:SSLv3/TLS write client certificate
   SSL_connect:SSLv3/TLS write certificate verify
   SSL_connect:SSLv3/TLS write finished
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSL negotiation finished successfully
   SSL_connect:SSLv3/TLS read server session ticket
   @

Test Staging Environment
************************

.. code:: bash

   debug-tools:~#
   curl -vk \
   https://us-staging-1-compute-api-server.idcs-system.svc.cluster.local:8443

.. code:: console

   *   Trying 100.83.168.59:8443...
   * Connected to us-staging-1-compute-api-server.idcs-system.svc.cluster.local (100.83.168.59) port 8443 (#0)
   * ALPN: offers h2,http/1.1
   * TLSv1.3 (OUT), TLS handshake, Client hello (1):
   * TLSv1.3 (IN), TLS handshake, Server hello (2):
   * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
   * TLSv1.3 (IN), TLS handshake, Request CERT (13):
   * TLSv1.3 (IN), TLS handshake, Certificate (11):
   * TLSv1.3 (IN), TLS handshake, CERT verify (15):
   * TLSv1.3 (IN), TLS handshake, Finished (20):
   * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
   * TLSv1.3 (OUT), TLS handshake, Certificate (11):
   * TLSv1.3 (OUT), TLS handshake, Finished (20):
   * SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
   * ALPN: server accepted h2
   * Server certificate:
   *  subject: OU=us-staging-1-compute-api-server; CN=us-staging-1-compute-api-server.idcs-system.svc.cluster.local
   *  start date: Sep  4 15:09:50 2024 GMT
   *  expire date: Sep  5 15:10:20 2024 GMT
   *  issuer: CN=Intel IDC CA dc5f4b4c us-staging-1-ca
   *  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
   * using HTTP/2
   * getpeername() failed with errno 107: Socket not connected
   * getpeername() failed with errno 107: Socket not connected
   * h2h3 [:method: GET]
   * h2h3 [:path: /]
   * h2h3 [:scheme: https]
   * h2h3 [:authority: us-staging-1-compute-api-server.idcs-system.svc.cluster.local:8443]
   * h2h3 [user-agent: curl/8.0.1]
   * h2h3 [accept: */*]
   * Using Stream ID: 1 (easy handle 0x7f294e815a90)
   * Send failure: Broken pipe
   * OpenSSL SSL_write: Broken pipe, errno 32
   * Failed sending HTTP request
   * Connection #0 to host us-staging-1-compute-api-server.idcs-system.svc.cluster.local left intact
   curl: (55) getpeername() failed with errno 107: Socket not connected
