/*
Copyright 2022 The Kubernetes Authors.

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

package v1alpha1

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kueuev1alpha1 "sigs.k8s.io/kueue/apis/kueue/v1alpha1"
	"sigs.k8s.io/kueue/test/integration/framework"
	// +kubebuilder:scaffold:imports
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	fwk       *framework.Framework
	ctx       context.Context
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t,
		"Webhook Suite",
	)
}

var _ = ginkgo.BeforeSuite(func() {
	fwk = &framework.Framework{
		CRDPath:     filepath.Join("..", "..", "..", "..", "config", "crd", "bases"),
		WebhookPath: filepath.Join("..", "..", "..", "..", "config", "webhook"),
		ManagerSetup: func(mgr manager.Manager, ctx context.Context) {
			err := (&kueuev1alpha1.Workload{}).SetupWebhookWithManager(mgr)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		},
	}
	ctx, cfg, k8sClient = fwk.Setup()
})

var _ = ginkgo.AfterSuite(func() {
	fwk.Teardown()
})
