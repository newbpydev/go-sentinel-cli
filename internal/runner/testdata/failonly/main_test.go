package failonly

import "testing"

func TestAlwaysFail(t *testing.T) {
	t.Log("this test always fails")
	t.Fail()
}
