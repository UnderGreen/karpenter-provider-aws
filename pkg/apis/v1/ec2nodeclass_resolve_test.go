/*
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

package v1_test

import (
	corev1 "k8s.io/api/core/v1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/scheduling"

	"github.com/aws/aws-sdk-go-v2/aws"

	v1 "github.com/aws/karpenter-provider-aws/pkg/apis/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveBlockDeviceMappings", func() {
	defaultBDM := []*v1.BlockDeviceMapping{
		{DeviceName: aws.String("default-device")},
	}
	gpuOverrideBDM := []*v1.BlockDeviceMapping{
		{DeviceName: aws.String("gpu-device")},
	}
	catchAllOverrideBDM := []*v1.BlockDeviceMapping{
		{DeviceName: aws.String("catch-all-device")},
	}
	gpuRequirements := scheduling.NewLabelRequirements(map[string]string{
		"karpenter.k8s.aws/instance-category": "g",
	})
	cpuRequirements := scheduling.NewLabelRequirements(map[string]string{
		"karpenter.k8s.aws/instance-category": "c",
	})
	gpuOverride := v1.BlockDeviceMappingOverride{
		Requirements: []karpv1.NodeSelectorRequirementWithMinValues{
			{Key: "karpenter.k8s.aws/instance-category", Operator: corev1.NodeSelectorOpIn, Values: []string{"g", "p"}},
		},
		BlockDeviceMappings: gpuOverrideBDM,
	}
	catchAllOverride := v1.BlockDeviceMappingOverride{
		BlockDeviceMappings: catchAllOverrideBDM,
	}

	It("should return the first matching override, ignoring later matches", func() {
		result := v1.ResolveBlockDeviceMappings(gpuRequirements, defaultBDM, []v1.BlockDeviceMappingOverride{gpuOverride, catchAllOverride})
		Expect(result).To(Equal(gpuOverrideBDM))
	})
	It("should fall through to a later override if an earlier one doesn't match", func() {
		result := v1.ResolveBlockDeviceMappings(cpuRequirements, defaultBDM, []v1.BlockDeviceMappingOverride{gpuOverride, catchAllOverride})
		Expect(result).To(Equal(catchAllOverrideBDM))
	})
	It("should treat an empty requirements override as a catch-all", func() {
		result := v1.ResolveBlockDeviceMappings(gpuRequirements, defaultBDM, []v1.BlockDeviceMappingOverride{catchAllOverride})
		Expect(result).To(Equal(catchAllOverrideBDM))
	})
	It("should fall back to the top-level blockDeviceMappings when no override matches", func() {
		result := v1.ResolveBlockDeviceMappings(cpuRequirements, defaultBDM, []v1.BlockDeviceMappingOverride{gpuOverride})
		Expect(result).To(Equal(defaultBDM))
	})
	It("should fall back to nil when neither an override nor blockDeviceMappings are set", func() {
		result := v1.ResolveBlockDeviceMappings(cpuRequirements, nil, []v1.BlockDeviceMappingOverride{gpuOverride})
		Expect(result).To(BeNil())
	})
	It("should resolve through the EC2NodeClass method the same way as the free function", func() {
		nc := &v1.EC2NodeClass{
			Spec: v1.EC2NodeClassSpec{
				BlockDeviceMappings:         defaultBDM,
				BlockDeviceMappingOverrides: []v1.BlockDeviceMappingOverride{gpuOverride, catchAllOverride},
			},
		}
		Expect(nc.ResolveBlockDeviceMappings(gpuRequirements)).To(Equal(gpuOverrideBDM))
		Expect(nc.ResolveBlockDeviceMappings(cpuRequirements)).To(Equal(catchAllOverrideBDM))
	})
})
