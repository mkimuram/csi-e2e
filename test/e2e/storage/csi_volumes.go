/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package storage

import (
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
	"k8s.io/kubernetes/test/e2e/storage/testsuites/testlib"
	"k8s.io/kubernetes/test/e2e/storage/testsuites/testlib/driverlib"
	"k8s.io/kubernetes/test/e2e/storage/testsuites/testlib/patterns"

	. "github.com/onsi/ginkgo"
)

var testDrivers = []func() driverlib.TestDriver{
	InitSampleCSIDriver,
}

// List of testsuites to be executed in below loop
var testSuites = []func() testlib.TestSuite{
	testsuites.InitVolumesTestSuite,
	testsuites.InitVolumeIOTestSuite,
	testsuites.InitVolumeModeTestSuite,
	testsuites.InitSubPathTestSuite,
	testsuites.InitProvisioningTestSuite,
	// Customized tests can be added like below. Also see customtest_sample.go.
	//InitCustomVolumesTestSuite,
}

func tunePattern(p []patterns.TestPattern) []patterns.TestPattern {
	tunedPatterns := []patterns.TestPattern{}

	for _, pattern := range p {
		// Skip inline volume and pre-provsioned PV tests for csi drivers
		if pattern.VolType == patterns.InlineVolume || pattern.VolType == patterns.PreprovisionedPV {
			continue
		}
		tunedPatterns = append(tunedPatterns, pattern)
	}

	return p
}

var _ = Describe("CSI Volumes", testlib.GetStorageTestFunc("csi-out-of-tree", "csi", testDrivers, testSuites, tunePattern))
