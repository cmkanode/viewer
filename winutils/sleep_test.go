package winutils

import "testing"

func TestUpdateSleepPreventionState(t *testing.T) {
	var prevented int
	var allowed int

	prevent := func() { prevented++ }
	allow := func() { allowed++ }

	state := false
	state = UpdateSleepPreventionState(state, true, prevent, allow)
	if !state {
		t.Fatal("expected sleep prevention to be enabled when the app is focused")
	}
	if prevented != 1 || allowed != 0 {
		t.Fatalf("unexpected call counts after enabling: prevented=%d allowed=%d", prevented, allowed)
	}

	state = UpdateSleepPreventionState(state, false, prevent, allow)
	if state {
		t.Fatal("expected sleep prevention to be disabled when the app is no longer focused")
	}
	if prevented != 1 || allowed != 1 {
		t.Fatalf("unexpected call counts after disabling: prevented=%d allowed=%d", prevented, allowed)
	}

	state = UpdateSleepPreventionState(state, false, prevent, allow)
	if state {
		t.Fatal("expected state to remain disabled without repeated toggles")
	}
	if prevented != 1 || allowed != 1 {
		t.Fatalf("expected no extra calls when state is unchanged: prevented=%d allowed=%d", prevented, allowed)
	}
}
