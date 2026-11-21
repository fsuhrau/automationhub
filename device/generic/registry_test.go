package generic_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func ids(devices []*testkit.FakeDevice) []string {
	out := make([]string, len(devices))
	for i, d := range devices {
		out[i] = d.DeviceID()
	}
	return out
}

func TestRegistryAddGetHasRemove(t *testing.T) {
	var r generic.Registry[*testkit.FakeDevice]
	a := testkit.NewFakeDevice("a")
	if r.Has("a") {
		t.Fatal("empty registry has a")
	}
	if r.Remove("a") {
		t.Error("Remove on an empty registry reports true")
	}
	r.Add(a)
	if got, ok := r.Get("a"); !ok || got != a {
		t.Errorf("Get(a) = %v, %v", got, ok)
	}
	if _, ok := r.Get("b"); ok {
		t.Error("Get(b) found a device")
	}
	if !r.HasDevice(a) || r.HasDevice(testkit.NewFakeDevice("b")) {
		t.Error("HasDevice wrong")
	}
	if !r.Remove("a") || r.Has("a") {
		t.Error("Remove did not drop a")
	}
}

func TestRegistryListIsSortedAndAddReplaces(t *testing.T) {
	var r generic.Registry[*testkit.FakeDevice]
	r.Add(testkit.NewFakeDevice("b"))
	r.Add(testkit.NewFakeDevice("a"))
	r.Add(testkit.NewFakeDevice("b"))
	if got := ids(r.List()); fmt.Sprint(got) != "[a b]" {
		t.Errorf("List() = %v", got)
	}
	devices, err := r.GetDevices()
	if err != nil || len(devices) != 2 || devices[0].DeviceID() != "a" {
		t.Errorf("GetDevices() = %v, %v", devices, err)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	var r generic.Registry[*testkit.FakeDevice]
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			r.Add(testkit.NewFakeDevice(id))
			r.Has(id)
			_ = r.List()
			r.Remove(id)
		}(fmt.Sprint(i))
	}
	wg.Wait()
	if n := len(r.List()); n != 0 {
		t.Errorf("%d devices left", n)
	}
}
