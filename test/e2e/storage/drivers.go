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
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/testsuites/testlib/driverlib"
	"k8s.io/kubernetes/test/e2e/storage/testsuites/testlib/patterns"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

type sampleCSIDriver struct {
	cleanup    func()
	driverInfo driverlib.DriverInfo
}

var _ driverlib.TestDriver = &sampleCSIDriver{}
var _ driverlib.DynamicPVTestDriver = &sampleCSIDriver{}

// InitSampleCSIDriver returns sampleCSIDriver that implements driverlib.TestDriver interface
func InitSampleCSIDriver() driverlib.TestDriver {
	return &sampleCSIDriver{
		driverInfo: driverlib.DriverInfo{
			Name:        "csi-hostpath",
			FeatureTag:  "",
			MaxFileSize: patterns.FileSizeMedium,
			SupportedFsType: sets.NewString(
				"", // Default fsType
			),
			IsPersistent:       true,
			IsFsGroupSupported: false,
			IsBlockSupported:   false,
		},
	}
}

func (h *sampleCSIDriver) GetDriverInfo() *driverlib.DriverInfo {
	return &h.driverInfo
}

func (h *sampleCSIDriver) SkipUnsupportedTest(pattern patterns.TestPattern) {
}

func (h *sampleCSIDriver) GetDynamicProvisionStorageClass(fsType string) *storagev1.StorageClass {
	provisioner := driverlib.GetUniqueDriverName(h)
	parameters := map[string]string{}
	ns := h.driverInfo.Framework.Namespace.Name
	suffix := fmt.Sprintf("%s-sc", provisioner)

	return driverlib.GetStorageClass(provisioner, parameters, nil, ns, suffix)
}

func (h *sampleCSIDriver) CreateDriver() {
	By("deploying csi sample driver")
	f := h.driverInfo.Framework
	cs := f.ClientSet

	// pods should be scheduled on the node
	nodes := framework.GetReadySchedulableNodesOrDie(cs)
	node := nodes.Items[rand.Intn(len(nodes.Items))]
	h.driverInfo.Config.ClientNodeName = node.Name
	h.driverInfo.Config.ServerNodeName = node.Name

	// TODO (?): the storage.csi.image.version and storage.csi.image.registry
	// settings are ignored for this test. We could patch the image definitions.
	o := utils.PatchCSIOptions{
		OldDriverName:            h.driverInfo.Name,
		NewDriverName:            driverlib.GetUniqueDriverName(h),
		DriverContainerName:      "hostpath",
		ProvisionerContainerName: "csi-provisioner",
		NodeName:                 h.driverInfo.Config.ServerNodeName,
	}
	cleanup, err := h.driverInfo.Framework.CreateFromManifests(func(item interface{}) error {
		return utils.PatchCSIDeployment(h.driverInfo.Framework, o, item)
	},
		"test/e2e/storage/manifests/driver-registrar/rbac.yaml",
		"test/e2e/storage/manifests/external-attacher/rbac.yaml",
		"test/e2e/storage/manifests/external-provisioner/rbac.yaml",
		"test/e2e/storage/manifests/hostpath/hostpath/csi-hostpath-attacher.yaml",
		"test/e2e/storage/manifests/hostpath/hostpath/csi-hostpath-provisioner.yaml",
		"test/e2e/storage/manifests/hostpath/hostpath/csi-hostpathplugin.yaml",
		"test/e2e/storage/manifests/hostpath/hostpath/e2e-test-rbac.yaml",
	)
	h.cleanup = cleanup
	if err != nil {
		framework.Failf("deploying csi sample driver: %v", err)
	}
}

func (h *sampleCSIDriver) CleanupDriver() {
	if h.cleanup != nil {
		By("uninstalling csi sample driver")
		h.cleanup()
	}
}
