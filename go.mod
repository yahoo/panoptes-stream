module git.vzbuilders.com/marshadrad/panoptes

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/consul/api v1.5.0
	github.com/hashicorp/vault v1.4.3
	github.com/hashicorp/vault-plugin-secrets-kv v0.5.5
	github.com/hashicorp/vault/api v1.0.5-0.20200317185738-82f498082f02
	github.com/hashicorp/vault/sdk v0.1.14-0.20200702114606-96dd7d6e10db
	github.com/influxdata/influxdb v1.8.1
	github.com/influxdata/influxdb-client-go v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/openconfig/gnmi v0.0.0-20200617225440-d2b4e6a45802
	github.com/openconfig/ygot v0.8.1
	github.com/prometheus/client_golang v1.7.1
	github.com/segmentio/kafka-go v0.3.7
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.2.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200520232829-54ba9589114f
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.23.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	software.sslmate.com/src/go-pkcs12 v0.0.0-20200619203921-c9ed90bd32dc
)

replace (
	github.com/coreos/bbolt => github.com/coreos/bbolt v1.3.2
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.10+incompatible
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200520232829-54ba9589114f // 54ba9589114f is the SHA for git tag v3.4.9
)
