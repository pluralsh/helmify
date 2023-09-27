package rbac

import (
	"testing"

	"github.com/pluralsh/helmify/pkg/metadata"

	"github.com/pluralsh/helmify/internal"
	"github.com/stretchr/testify/assert"
)

const clusterRoleYaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: my-operator-manager-role
aggregationRule:
  clusterRoleSelectors:
  - matchExpressions:
    - key: my.operator.dev/release
      operator: Exists
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list`

func Test_clusterRole_Process(t *testing.T) {
	var testInstance role

	t.Run("processed", func(t *testing.T) {
		obj := internal.GenerateObj(clusterRoleYaml)
		processed, _, err := testInstance.Process(&metadata.Service{}, obj)
		assert.NoError(t, err)
		assert.Equal(t, true, processed)
	})
	t.Run("skipped", func(t *testing.T) {
		obj := internal.TestNs
		processed, _, err := testInstance.Process(&metadata.Service{}, obj)
		assert.NoError(t, err)
		assert.Equal(t, false, processed)
	})
}
