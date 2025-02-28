module github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go

go 1.23

toolchain go1.23.2

require (
	cloud.google.com/go v0.116.0
	cloud.google.com/go/gaming v1.10.1
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/AppsFlyer/go-sundheit v0.6.0
	github.com/AppsFlyer/wrk3 v0.0.0-20210816075316-aa3ea7d90a33
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/IguteChung/casbin-psql-watcher v1.0.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/alron/ginlogr v0.0.4
	github.com/amacneil/dbmate v1.16.0
	github.com/apenella/go-ansible v1.2.0
	github.com/aristanetworks/goeapi v1.0.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/aws/aws-sdk-go v1.49.13
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.36
	github.com/aws/aws-sdk-go-v2/credentials v1.17.34
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.5.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.33.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.29.4
	github.com/aws/aws-sdk-go-v2/service/sqs v1.31.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.31.0
	github.com/aws/smithy-go v1.21.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/bougou/go-ipmi v0.0.0-20230702095700-b7dd37559417
	github.com/bramvdbogaerde/go-scp v1.2.1
	github.com/casbin/casbin/v2 v2.84.0
	github.com/casbin/gorm-adapter/v3 v3.23.1
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cert-manager/cert-manager v1.13.6
	github.com/chzyer/logex v1.2.0
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23
	github.com/danjacques/gofslock v0.0.0-20240212154529-d899e02bfe22
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0
	github.com/edsrzf/mmap-go v1.0.0
	github.com/elazarl/goproxy v0.0.0-20230731152917-f99041a5c027
	github.com/envoyproxy/protoc-gen-validate v1.0.4
	github.com/exaring/otelpgx v0.4.1
	github.com/fatih/structs v1.1.0
	github.com/flosch/pongo2 v0.0.0-20200913210552-0d938eb266f3
	github.com/flowchartsman/swaggerui v0.0.0-20221017034628-909ed4f3701b
	github.com/friendsofgo/errors v0.9.2
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gin-contrib/zap v0.1.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v1.4.2
	github.com/go-openapi/errors v0.20.2
	github.com/go-openapi/runtime v0.25.0
	github.com/go-openapi/strfmt v0.21.2
	github.com/go-openapi/swag v0.23.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-test/deep v1.1.0
	github.com/gobuffalo/logger v1.0.6
	github.com/gobuffalo/packd v1.0.1
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/gocarina/gocsv v0.0.0-20230616125104-99d496ca653d
	github.com/gofrs/uuid v4.3.1+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.14.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/gofuzz v1.2.0
	github.com/google/renameio v1.0.1
	github.com/google/s2a-go v0.1.8
	github.com/google/uuid v1.6.0
	github.com/googleapis/gax-go/v2 v2.13.0
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0
	github.com/hashicorp/go-getter v1.7.6
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/hashicorp/vault/api v1.15.0
	github.com/hashicorp/vault/api/auth/approle v0.3.0
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgx/v5 v5.5.4
	github.com/jarcoal/httpmock v1.3.0
	github.com/jedib0t/go-pretty/v6 v6.5.9
	github.com/jinzhu/copier v0.4.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.6.0
	github.com/karrick/godirwalk v1.16.1
	github.com/knadh/koanf v1.4.4
	github.com/lestrrat-go/blackmagic v1.0.1
	github.com/lestrrat-go/iter v1.0.2
	github.com/lestrrat-go/option v1.0.1
	github.com/lib/pq v1.10.9
	github.com/markbates/oncer v1.0.0
	github.com/metal3-io/baremetal-operator/apis v0.3.1
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.2.0
	github.com/miekg/dns v1.1.62
	github.com/minio/minio-go/v7 v7.0.78
	github.com/minio/sha256-simd v1.0.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mittwald/go-helm-client v0.12.6
	github.com/muonsoft/validation v0.16.0
	github.com/netbox-community/go-netbox v0.0.0-20230225105939-fe852c86b3d6
	github.com/netbox-community/go-netbox/v3 v3.7.1-alpha.0
	github.com/nirasan/go-oauth-pkce-code-verifier v0.0.0-20220510032225-4f9f17eaec4c
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.19.0
	github.com/onsi/gomega v1.33.1
	github.com/open-policy-agent/opa v0.47.3
	github.com/open-policy-agent/opa-envoy-plugin v0.47.3-envoy
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/ovn-org/libovsdb v0.6.1-0.20240125124854-03f787b1a892
	github.com/ovn-org/ovn-kubernetes/go-controller v0.0.0-20240520202957-f71cdcb4c186
	github.com/pborman/uuid v1.2.1
	github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.5
	github.com/playwright-community/playwright-go v0.4201.1
	github.com/praserx/ipconv v1.2.1
	github.com/prometheus/client_golang v1.18.0
	github.com/rancher/lasso v0.0.0-20240424194130-d87ec407d941
	github.com/rancher/norman v0.0.0-20240517190939-17053b35e056
	github.com/rancher/rancher/pkg/apis v0.0.0-20230113190314-623addd2b1f0
	github.com/rancher/rancher/pkg/client v0.0.0-20230323194012-41b2cf3ab63f
	github.com/rancher/wrangler v1.1.1
	github.com/rogpeppe/go-internal v1.12.0
	github.com/sergi/go-diff v1.2.0
	github.com/sethvargo/go-password v0.2.0
	github.com/sethvargo/go-retry v0.2.4
	github.com/sirupsen/logrus v1.9.3
	github.com/sony/gobreaker v0.5.0
	github.com/spdx/tools-golang v0.5.5
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.16.0
	github.com/stmcginnis/gofish v0.14.0
	github.com/stretchr/testify v1.9.0
	github.com/testcontainers/testcontainers-go v0.32.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.32.0
	github.com/thediveo/enumflag/v2 v2.0.5
	github.com/tidwall/gjson v1.17.0
	github.com/vishvananda/netns v0.0.4
	github.com/xdg-go/stringprep v1.0.4
	github.com/zitadel/oidc v1.13.0
	go.etcd.io/etcd/api/v3 v3.5.14
	go.etcd.io/etcd/client/v3 v3.5.14
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.45.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.20.0
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.31.0
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	golang.org/x/mod v0.20.0
	golang.org/x/oauth2 v0.23.0
	golang.org/x/sync v0.10.0
	golang.org/x/text v0.21.0
	golang.org/x/time v0.7.0
	google.golang.org/api v0.198.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.35.1
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/warnings.v0 v0.1.2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.9
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.5.1
	helm.sh/helm/v3 v3.15.0-rc.1
	istio.io/api v1.22.6
	istio.io/client-go v1.22.6
	k8s.io/api v0.29.9
	k8s.io/apimachinery v0.29.9
	k8s.io/apiserver v0.29.9
	k8s.io/cli-runtime v0.29.9
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.29.9
	k8s.io/component-helpers v0.29.9
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.29.9
	k8s.io/utils v0.0.0-20240921022957-49e7df575cb6
	kubevirt.io/api v1.0.0-alpha.0
	oras.land/oras-go v1.2.5
	sigs.k8s.io/cluster-api v1.5.3
	sigs.k8s.io/controller-runtime v0.17.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/alexflint/go-filemutex v1.2.0 // indirect
	github.com/cenkalti/hub v1.0.1 // indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect
	github.com/containerd/errdefs v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/containernetworking/cni v1.1.2 // indirect
	github.com/containernetworking/plugins v1.2.0 // indirect
	github.com/coreos/go-iptables v0.6.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/safchain/ethtool v0.3.1-0.20231027162144-83e5e0097c91 // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/urfave/cli/v2 v2.20.2 // indirect
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20231024175852-77df5d35f725 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	sigs.k8s.io/gateway-api v1.0.0-rc2 // indirect
)

require (
	cloud.google.com/go/longrunning v0.6.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.11.5 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/anchore/go-struct-converter v0.0.0-20221118182256-c68fdcfa2092 // indirect
	github.com/apenella/go-common-utils/data v0.0.0-20220913191136-86daaa87e7df // indirect
	github.com/apenella/go-common-utils/error v0.0.0-20220913191136-86daaa87e7df // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.23.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.27.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytecodealliance/wasmtime-go/v3 v3.0.2 // indirect
	github.com/bytedance/sonic v1.11.9 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/casbin/govaluate v1.1.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cncf/xds/go v0.0.0-20240423153145-555b57ec207b // indirect
	github.com/containerd/containerd v1.7.18 // indirect
	github.com/containerd/continuity v0.4.2 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.4 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v25.0.1+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v27.1.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/envoyproxy/go-control-plane v0.12.1-0.20240621013728-1eb8caab5155 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/glebarez/go-sqlite v1.20.3 // indirect
	github.com/glebarez/sqlite v1.7.0 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/analysis v0.21.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.21.1 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/validate v0.21.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gobuffalo/flect v1.0.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/glog v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.8 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jerryzhen01/go-netbox/v4 v4.0.10-0.20241015162424-ba1d557364c7
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/microsoft/go-mssqldb v1.6.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/spdystream v0.4.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.1.12 // indirect
	github.com/openshift/api v0.0.0-20231120222239-b86761094ee3 // indirect
	github.com/openshift/custom-resource-status v1.1.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/peterh/liner v1.2.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rancher/wrangler/v2 v2.2.0-rc6 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rs/cors v1.8.3 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/rubenv/sql-migrate v1.7.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/yashtewari/glob-intersection v0.1.0 // indirect
	github.com/zitadel/logging v0.3.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.14 // indirect
	go.mongodb.org/mongo-driver v1.8.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.starlark.net v0.0.0-20231016134836-22325403fcb3 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/automaxprocs v1.5.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	gonum.org/v1/gonum v0.13.1-0.20230729095443-194082cf5ba1 // indirect
	google.golang.org/genproto v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1
	gopkg.in/evanphx/json-patch.v5 v5.6.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/sqlserver v1.5.3 // indirect
	gorm.io/plugin/dbresolver v1.3.0 // indirect
	k8s.io/apiextensions-apiserver v0.29.9 // indirect
	k8s.io/code-generator v0.30.0-alpha.3 // indirect
	k8s.io/gengo v0.0.0-20240310015720-9cff6334dab4 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	kubevirt.io/containerized-data-importer-api v1.57.0-alpha1
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.0.0-20220329064328-f3cc58c6ed90 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.20.3 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.16.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.16.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.29.9
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340
)
