//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/discovery"
)

type k8s struct {
	cfg       config.Config
	logger    *zap.Logger
	clientset kubernetes.Interface
	namespace string
}

type k8sConfig struct {
	Namespace string
}

// New constructs new k8s service discovery.
func New(cfg config.Config) (discovery.Discovery, error) {
	var (
		err error
		k8s = &k8s{cfg: cfg, logger: cfg.Logger()}
	)

	config, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_discovery_k8s"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	if config.Namespace != "" {
		k8s.namespace = config.Namespace
	} else {
		k8s.namespace = "default"
	}

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	k8s.clientset, err = kubernetes.NewForConfig(clusterConfig)

	return k8s, err
}

// Register doesn't register the nodes as they're
// already registered at kubernetes. it patches the
// Panoptes pods with specific annotations.
func (k *k8s) Register() error {
	var annotations = map[string]string{}

	hostname, err := getHostname()
	if err != nil {
		return err
	}

	name := strings.Split(hostname, "-")
	if len(name) < 3 {
		return errors.New("statefulset pod name is invalid")
	}

	if k.cfg.Global().Shards.Enabled {
		annotations["shards_enabled"] = "true"
	} else {
		annotations["shards_enabled"] = "false"
	}

	annotations["version"] = k.cfg.Global().Version
	annotations["shards_nodes"] = strconv.Itoa(k.cfg.Global().Shards.NumberOfNodes)

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	}

	b, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = k.clientset.CoreV1().Pods(k.namespace).Patch(
		context.TODO(),
		hostname,
		types.StrategicMergePatchType,
		b,
		metav1.PatchOptions{})

	if err != nil {
		return err
	}

	k.logger.Info("k8s", zap.String("event", "register"), zap.String("ns", k.namespace), zap.String("id", name[len(name)-1]))

	return nil
}

// Deregister satisfies discovery interface.
// k8s doesn't have deregistry service.
func (*k8s) Deregister() error {
	// not available
	return nil
}

// GetInstances returns all instances.
func (k *k8s) GetInstances() ([]discovery.Instance, error) {
	var instances []discovery.Instance

	ctx := context.Background()
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=panoptes-stream",
	})
	if err != nil {
		return nil, err
	}

	for _, item := range pods.Items {
		name := strings.Split(item.Name, "-")
		if len(name) < 3 {
			panic("statefulset pod name is invalid")
		}

		instances = append(instances, discovery.Instance{
			ID:      name[len(name)-1],
			Meta:    item.Annotations,
			Address: item.Name,
			Status:  convPhaseToStatus(item.Status.Phase),
		})

	}

	return instances, nil
}

// Watch monitors status of instances and notify through the channel.
func (k *k8s) Watch(ch chan<- struct{}) {
	optionsModifer := func(options *metav1.ListOptions) {
		options.LabelSelector = "app=panoptes-stream"
	}

	listWatcher := cache.NewFilteredListWatchFromClient(
		k.clientset.CoreV1().RESTClient(),
		"pods",
		k.namespace,
		optionsModifer,
	)

	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&v1.Pod{},
		time.Second*5,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			select {
			case ch <- struct{}{}:
			default:
				k.logger.Debug("k8s", zap.String("event", "watcher.response.drop"))
			}
		},
		DeleteFunc: func(obj interface{}) {
			select {
			case ch <- struct{}{}:
			default:
				k.logger.Debug("k8s", zap.String("event", "watcher.response.drop"))
			}
		},
	})

	stop := make(chan struct{})
	informer.Run(stop)

	select {}
}

func convPhaseToStatus(p v1.PodPhase) string {
	if p == v1.PodRunning {
		return "passing"
	}

	return "warning"
}

func getConfig(cfg config.Config) (*k8sConfig, error) {
	k8sConfig := new(k8sConfig)
	b, err := json.Marshal(cfg.Global().Discovery.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, k8sConfig)

	return k8sConfig, err
}

func getHostname() (string, error) {
	if os.Getenv("PANOPTES_TEST") != "" {
		return "panoptes-stream-0", nil
	}

	return os.Hostname()
}
