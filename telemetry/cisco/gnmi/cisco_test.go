package gnmi

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/mock"
)

func TestGetPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()
	prefix, labels, output := g.getPrefix(buf, md.Prefix)

	assert.Equal(t, "/interfaces/interface/state/counters", prefix)
	assert.Equal(t, "out::out", output)
	assert.Equal(t, map[string]string{"name": "GigabitEthernet0/0/0/0"}, labels)
}

func TestGetPrefixWithKey(t *testing.T) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface[name=GigabitEthernet0/0/0/0]/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()
	prefix, labels, output := g.getPrefix(buf, md.Prefix)

	assert.Equal(t, "/interfaces/interface/state/counters", prefix)
	assert.Equal(t, "out::out", output)
	assert.Equal(t, map[string]string{"name": "GigabitEthernet0/0/0/0"}, labels)
}

func TestGetKeys(t *testing.T) {
	buf := new(bytes.Buffer)
	md := mock.CiscoXRInterface()
	key, labels := telemetry.GetKey(buf, md.Update[0].Path.Elem)
	assert.Equal(t, "in-octets", key)
	assert.Equal(t, map[string]string{}, labels)
}

func BenchmarkGetPrefix(b *testing.B) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()

	for i := 0; i < b.N; i++ {
		g.getPrefix(buf, md.Prefix)
	}

}

func BenchmarkGetPrefixWithKey(b *testing.B) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface[name=GigabitEthernet0/0/0/0]/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()

	for i := 0; i < b.N; i++ {
		g.getPrefix(buf, md.Prefix)
	}
}
