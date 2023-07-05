package processor

import (
	"github.com/pluralsh/helmify/pkg/config"
	"testing"

	"github.com/pluralsh/helmify/internal"
	"github.com/pluralsh/helmify/pkg/metadata"
	"github.com/stretchr/testify/assert"
)

func TestProcessObjMeta(t *testing.T) {
	testMeta := metadata.New(config.Config{ChartName: "chart-name"})
	testMeta.Load(internal.TestNs)
	res, err := ProcessObjMeta(testMeta, internal.TestNs)
	assert.NoError(t, err)
	assert.Contains(t, res, "chart-name.labels")
	assert.Contains(t, res, "chart-name.fullname")
}
