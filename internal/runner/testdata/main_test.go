package testdata

import "testing"

func TestSimplePass(t *testing.T) {
	t.Log("this test should pass")
}

func TestSimpleFail(t *testing.T) {
	t.Log("this test should fail")
	t.Fail()
}
