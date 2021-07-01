package e2e

import (
	"context"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-gce/pkg/fuzz"
	"k8s.io/klog"
)

// RunWithSandboxFixed runs the testFunc with a names sandbox, and
// if the sandbox exists it will be reused. This is usefull only on development
func (f *Framework) RunWithSandboxFixed(name string, sanboxNamespace string, t *testing.T, testFunc func(*testing.T, *Sandbox)) {
	t.Run(name, func(t *testing.T) {
		f.lock.Lock()
		sandbox := &Sandbox{
			Namespace: sanboxNamespace,
			f:         f,
			RandInt:   0,
		}
		for _, s := range f.sandboxes {
			if s.Namespace == sandbox.Namespace {
				f.lock.Unlock()
				t.Fatalf("Sandbox %s was created previously by the framework.", s.Namespace)
			}
		}
		klog.V(2).Infof("Using namespace %q for test sandbox", sandbox.Namespace)

		if err := sandbox.Ensure(); err != nil {
			f.lock.Unlock()
			t.Fatalf("error creating sandbox: %v", err)
		}

		f.sandboxes = append(f.sandboxes, sandbox)
		f.lock.Unlock()

		if f.destroySandboxes {
			defer sandbox.Destroy()
		}

		defer sandbox.DumpSandboxInfo(t)
		testFunc(t, sandbox)
	})
}

// Ensure the sandbox.
func (s *Sandbox) Ensure() error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Namespace,
		},
	}
	// TODO: try to get and and create or update
	_, err := s.f.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil && !strings.HasSuffix(err.Error(), "already exists") {
		klog.Errorf("Error creating namespace %q: %v", s.Namespace, err)
		return err
	}

	s.ValidatorEnv, err = fuzz.NewDefaultValidatorEnv(s.f.RestConfig, s.Namespace, s.f.Cloud)
	if err != nil {
		klog.Errorf("Error creating validator env for namespace %q: %v", s.Namespace, err)
		return err
	}

	return nil
}
