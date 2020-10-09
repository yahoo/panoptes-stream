//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package k8s

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/yahoo/panoptes-stream/config"
)

func TestRegisterGetInstances(t *testing.T) {
	os.Setenv("PANOPTES_TEST", "true")

	cfg := config.NewMockConfig()
	k := &k8s{
		cfg:       cfg,
		logger:    cfg.Logger(),
		namespace: "default",
		clientset: fake.NewSimpleClientset(),
	}

	cfg.Global().Shards.Enabled = true
	cfg.Global().Shards.NumberOfNodes = 3

	k.clientset.CoreV1().Pods("default").Create(context.Background(), pod("default"), metav1.CreateOptions{})

	err := k.Register()
	assert.NoError(t, err)

	podName := "panoptes-stream-0"
	assert.NoError(t, err)

	pod, err := k.clientset.CoreV1().Pods("default").Get(context.Background(), podName, metav1.GetOptions{})
	assert.NoError(t, err)

	assert.Equal(t, "true", pod.Annotations["shards_enabled"])
	assert.Equal(t, "3", pod.Annotations["shards_nodes"])
	assert.Equal(t, "0.0.0", pod.Annotations["version"])

	instances, err := k.GetInstances()
	assert.NoError(t, err)

	assert.Len(t, instances, 1)
	assert.Equal(t, "panoptes-stream-0", instances[0].Address)
}

func TestNew(t *testing.T) {
	cfg := config.NewMockConfig()
	_, err := New(cfg)
	assert.Contains(t, err.Error(), "unable to load in-cluster configuration")
}

func TestConvPhaseToStatus(t *testing.T) {
	assert.Equal(t, "passing", convPhaseToStatus("Running"))
	assert.Equal(t, "warning", convPhaseToStatus("NonRunning"))
}

func TestDeregister(t *testing.T) {
	k := &k8s{}
	k.Deregister()
}

func pod(namespace string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "panoptes-stream-0",
			Labels: map[string]string{
				"app": "panoptes-stream",
			},
		},
		Spec: v1.PodSpec{
			Hostname:   "panoptes-stream-0",
			Containers: []v1.Container{{Image: "panoptes-stream"}},
		},
	}
}
