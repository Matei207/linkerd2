package servicetopology

import (
	"github.com/linkerd/linkerd2/testutil"
	"os"
	"testing"
)

var (
	TestHelper *testutil.TestHelper
)

func TestMain(m *testing.M) {
	TestHelper = testutil.NewTestHelper()
	os.Exit(testutil.Run(m, TestHelper))
}

func TestServiceTopology(t *testing.T) {

	// First, upgrade linkerd
	// if it does not exist, install it instead
	// verify that the destination service is up n runnin ready to be gunnin
	// after that think we'll need to install sum stuff
	// ~do checks for namespace too seems to be the practice :thumbs_up me~


	testNamespace := TestHelper.GetTestNamespace("service-topology-test")
	controlPlaneNs := TestHelper.GetLinkerdNamespace()
	err := TestHelper.CreateDataPlaneNamespaceIfNotExists(testNamespace, nil)
	if err != nil {
		testutil.AnnotatedFatalf(t, "failed to create namespace", "failed to create %s namespace: %s", testNamespace, err)
	}

	// Upgrade Linkerd with slices enabled
	exec := []string{"upgrade", "--enable-endpoint-slices"}
	out, stderr, err := TestHelper.LinkerdRun(exec...)
	if err != nil {
		testutil.AnnotatedFatalf(t, "'linkerd upgrade' command failed",
			"'linkerd upgrade' command failed: \n%s\n%s", out, stderr)
	}
	out, err = TestHelper.KubectlApply(out, controlPlaneNs)
	if err != nil {
		testutil.AnnotatedFatalf(t, "kubectl apply command failed", "kubectl apply command failed\n%s", out)
	}

	if err := TestHelper.CheckPods(controlPlaneNs, "linkerd-destination", 1); err != nil {
		if rce, ok := err.(*testutil.RestartCountError); ok {
			testutil.AnnotatedWarn(t, "CheckPods timed-out", rce)
		} else {
			testutil.AnnotatedError(t, "CheckPods timed-out", err)
		}
	}

	if err := TestHelper.CheckDeployment(controlPlaneNs, "linkerd-destination", 1); err != nil {
		testutil.AnnotatedErrorf(t, "CheckDeployment timed-out", "Error validating deployment [%s]:\n%s", "terminus", err)
	}


	// Inject & install emojivoto app
	out, stderr, err = TestHelper.LinkerdRun("inject", "--manual", "testdata/traffic_split_application.yaml")
	if err != nil {
		testutil.AnnotatedFatalf(t, "'linkerd inject' command failed",
			"'linkerd inject' command failed\n%s\n%s", out, stderr)
	}
	out, err = TestHelper.KubectlApply(out, testNamespace)
	if err != nil {
		testutil.AnnotatedFatalf(t, "'kubectl apply' command failed",
			"'kubectl apply' command failed\n%s", out)
	}

	for _, deploy := range []string{"web", "emoji", "vote-bot", "voting"} {
		if err := TestHelper.CheckPods(testNamespace, deploy, 1); err != nil {
			if rce, ok := err.(*testutil.RestartCountError); ok {
				testutil.AnnotatedWarn(t, "CheckPods timed-out", rce)
			} else {
				testutil.AnnotatedError(t, "CheckPods timed-out", err)
			}
		}

		if err := TestHelper.CheckDeployment(testNamespace, deploy, 1); err != nil {
			testutil.AnnotatedErrorf(t, "CheckDeployment timed-out", "Error validating deployment [%s]:\n%s", "terminus", err)
		}
	}

	// Construct test cases with each test case being a different service with different locality/pref, I'd say 2 is enough -- one local, one zonal where zonal won't work
	// get the slices before we do all of that, we get all of the endpointslices, check the locality and where a req is supposed to go
	// note to do this we first need to check out the topology of the src to make sure it goes where it's supposed to go (?)
}
