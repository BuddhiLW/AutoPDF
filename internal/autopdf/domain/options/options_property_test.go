// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"slices"
	"testing"
	"testing/quick"
	"time"
)

func TestPropertyEnabledOptionsAreCompleteProjection(t *testing.T) {
	property := func(flags [5]bool) bool {
		buildOptions := NewBuildOptions()
		if flags[0] {
			buildOptions.EnableClean("build")
		}
		if flags[1] {
			buildOptions.EnableVerbose(3)
		}
		if flags[2] {
			buildOptions.EnableDebug("stderr")
		}
		if flags[3] {
			buildOptions.EnableForce(true)
		}
		if flags[4] {
			buildOptions.EnableWatch(time.Second)
		}

		allNames := [...]string{"clean", "verbose", "debug", "force", "watch"}
		want := make([]string, 0, len(allNames))
		for index, enabled := range flags {
			if enabled {
				want = append(want, allNames[index])
			}
		}

		got := buildOptions.GetEnabledOptions()
		return slices.Equal(got, want) && buildOptions.HasAnyEnabled() == (len(want) > 0)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}
