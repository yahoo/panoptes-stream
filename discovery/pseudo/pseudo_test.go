//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package pseudo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestCheckHTTP(t *testing.T) {
	cfg := config.NewMockConfig()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	cfg.Global().Discovery.Config = pseudoConfig{
		Instances: []string{ts.Listener.Addr().String(), "127.0.0.2:1355"},
		Probe:     "http",
		Path:      "",
		MaxRetry:  1,
	}

	d, err := New(cfg)
	assert.NoError(t, err)

	time.Sleep((2 + 2) * time.Second)

	instances, err := d.GetInstances()
	assert.NoError(t, err)

	hostname, err := os.Hostname()
	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	assert.Equal(t, hostname, instances[0].Address)
	assert.Equal(t, "passing", instances[0].Status)
	assert.Equal(t, "failure", instances[1].Status)
}

func TestDeepCopy(t *testing.T) {
	si := []*instance{{hostname: "test"}}
	di := make([]*instance, len(si))
	deepCopy(di, si)
	assert.NotEqual(t, fmt.Sprintf("%p", di[0]), fmt.Sprintf("%p", si[0]))
}

func TestWatcher(t *testing.T) {
	cfg := config.NewMockConfig()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	cfg.Global().Discovery.Config = pseudoConfig{
		Instances: []string{ts.Listener.Addr().String()},
		Probe:     "http",
		Path:      "",
	}

	d, err := New(cfg)
	assert.NoError(t, err)

	ch := make(chan struct{}, 1)
	go d.Watch(ch)

	triggered := false
L:
	for i := 0; i < 6; i++ {
		time.Sleep(1 * time.Second)
		select {
		case <-ch:
			triggered = true
			break L
		default:
		}
	}

	assert.True(t, triggered)
}

func TestRegister(t *testing.T) {
	// not available
	p := pseudo{}
	p.Register()
}

func TestDeregister(t *testing.T) {
	// not available
	p := pseudo{}
	p.Deregister()
}
