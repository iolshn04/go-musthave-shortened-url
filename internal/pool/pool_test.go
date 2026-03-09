package pool

import "testing"

type testObject struct {
	value int
}

func (t *testObject) Reset() {
	t.value = 0
}

func TestPoolReset(t *testing.T) {

	p := New(func() *testObject {
		return &testObject{}
	})

	obj := p.Get()

	obj.value = 100

	p.Put(obj)

	obj2 := p.Get()

	if obj2.value != 0 {
		t.Fatalf("expected value 0 after reset, got %d", obj2.value)
	}
}
