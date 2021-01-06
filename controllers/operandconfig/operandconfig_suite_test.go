//
// Copyright 2021 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package operandconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	etcdv1beta2 "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	nssv1 "github.com/IBM/ibm-namespace-scope-operator/api/v1"

	apiv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	"github.com/IBM/operand-deployment-lifecycle-manager/controllers/deploy"
	"github.com/IBM/operand-deployment-lifecycle-manager/controllers/operandregistry"
	"github.com/IBM/operand-deployment-lifecycle-manager/controllers/operandrequest"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const useExistingCluster = "USE_EXISTING_CLUSTER"

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	// scheme    = runtime.NewScheme()

	timeout  = time.Second * 300
	interval = time.Second * 5
)

func TestOperandConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"OperandConfig Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: UseExistingCluster(),
		CRDDirectoryPaths:  []string{filepath.Join("../..", "config", "crd", "bases"), filepath.Join("../..", "testbin", "crds")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = apiv1alpha1.AddToScheme(clientgoscheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	err = nssv1.AddToScheme(clientgoscheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = olmv1alpha1.AddToScheme(clientgoscheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = olmv1.AddToScheme(clientgoscheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = etcdv1beta2.AddToScheme(clientgoscheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: clientgoscheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// Start your controllers test logic
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             clientgoscheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	odlmManager := deploy.NewODLMManager(k8sManager)
	// Setup Manager with OperandRegistry Controller
	err = (&operandregistry.Reconciler{
		ODLMManager: odlmManager,
		Recorder:    k8sManager.GetEventRecorderFor("OperandRegistry"),
		Scheme:      k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())
	// Setup Manager with OperandConfig Controller
	err = (&Reconciler{
		ODLMManager: odlmManager,
		Recorder:    k8sManager.GetEventRecorderFor("OperandConfig"),
		Scheme:      k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())
	// Setup Manager with OperandRequest Controller
	err = (&operandrequest.Reconciler{
		ODLMManager: odlmManager,
		Recorder:    k8sManager.GetEventRecorderFor("OperandRequest"),
		Scheme:      k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	// End your controllers test logic

	close(done)
}, 600)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func UseExistingCluster() *bool {
	use := false
	if os.Getenv(useExistingCluster) != "" && os.Getenv(useExistingCluster) == "true" {
		use = true
	}
	return &use
}