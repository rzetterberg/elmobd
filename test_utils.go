package elmobd

import (
	"fmt"
	"testing"
)

/*==============================================================================
 * Utils
 */

func assert(t *testing.T, assertion bool, msg string) {
	if assertion {
		return
	}

	t.Fatal(fmt.Sprintf("Assertion %s failed", msg))
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	assert(
		t,
		a == b,
		fmt.Sprintf("'%v == %v'", a, b),
	)
}

func assertSuccess(t *testing.T, err error) {
	assert(
		t,
		err == nil,
		fmt.Sprintf("'err == nil' (%v)", err),
	)
}

func assertOBDParseSuccess(t *testing.T, command OBDCommand, outputs []string) {
	_, err := parseOBDResponse(
		command,
		outputs,
	)

	assertSuccess(t, err)
}
