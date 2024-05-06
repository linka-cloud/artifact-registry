module go.linka.cloud/artifact-registry

go 1.22.0

toolchain go1.22.2

replace code.gitea.io/gitea => github.com/go-gitea/gitea v1.20.4

require (
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb
	github.com/distribution/distribution/v3 v3.0.0-20231012071349-fe98c998600a
	github.com/dustin/go-humanize v1.0.1
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.2.2
	github.com/klauspost/compress v1.17.8
	github.com/lithammer/dedent v1.1.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/sassoftware/go-rpmutils v0.3.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/afero v1.11.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	github.com/ulikunitz/xz v0.5.12
	go.linka.cloud/env v0.0.0-20230215124643-494dd1a27d3c
	go.linka.cloud/grpc-toolkit v0.4.4-0.20231026145832-5d6b16a2c2a0
	go.linka.cloud/printer v0.0.0-20240221170110-7ea9393f148f
	golang.org/x/crypto v0.23.0
	golang.org/x/sync v0.7.0
	golang.org/x/term v0.20.0
	helm.sh/helm/v3 v3.14.4
	oras.land/oras-go/v2 v2.5.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Microsoft/hcsshim v0.12.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bombsimon/logrusr/v4 v4.1.0 // indirect
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/containerd v1.7.16 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v26.1.1+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v26.1.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.1 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20150114040149-fa567046d9b1 // indirect
	github.com/emicklei/go-restful/v3 v3.12.0 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru/arc/v2 v2.0.5 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.5 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.14.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.0.5 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.0.5 // indirect
	github.com/redis/go-redis/v9 v9.1.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.51.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	go.starlark.net v0.0.0-20240411212711-9b43f0afd521 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240429193739-8cf5692501f6 // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.30.0 // indirect
	k8s.io/apimachinery v0.30.0 // indirect
	k8s.io/cli-runtime v0.30.0 // indirect
	k8s.io/client-go v0.30.0 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240430033511-f0e62f92d13f // indirect
	k8s.io/utils v0.0.0-20240502163921-fe8a2dddb1d0 // indirect
	oras.land/oras-go v1.2.5 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.17.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.17.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)
