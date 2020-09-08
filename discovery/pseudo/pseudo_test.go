//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package pseudo

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestCheckHTTP(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.Global().Discovery.Config = pseudoConfig{
		Instances: []string{"127.0.0.1:1355"},
		Probe:     "http",
		Path:      "",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "panoptes") })
	go http.ListenAndServe("127.0.0.1:1355", nil)

	d, err := New(cfg)
	assert.NoError(t, err)

	time.Sleep(6 * time.Second)

	instances, err := d.GetInstances()
	assert.NoError(t, err)

	hostname, err := os.Hostname()
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, hostname, instances[0].Address)
	assert.Equal(t, "passing", instances[0].Status)

}
