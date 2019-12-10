package typredis_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/typical-go/typical-go/pkg/typobj"
	"github.com/typical-go/typical-go/pkg/typobj"
	"github.com/typical-go/typical-go/pkg/typobj"
	"github.com/typical-go/typical-rest-server/pkg/typredis"
)

func TestModule(t *testing.T) {
	m := &typredis.Module{}
	require.True(t, typobj.IsProvider(m))
	require.True(t, typobj.IsDestroyer(m))
	require.True(t, typobj.IsPreparer(m))
	require.True(t, typobj.IsBuildCommander(m))
	require.True(t, typobj.IsConfigurer(m))
}
