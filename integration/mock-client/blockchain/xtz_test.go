package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetXtzMonitorResponse(t *testing.T) {
	t.Run("creates mock XtzMonitorResponse",
		func(t *testing.T) {
			resp, err := getXtzResponse("monitor")
			require.NoError(t, err)
			monitor, ok := resp.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, monitor["hash"], "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D")
		})
}

func TestGetXtzOperationsResponse(t *testing.T) {
	t.Run("creates an appropriately structured mock Tezos block",
		func(t *testing.T) {
			resp, err := getXtzResponse("operations")
			require.NoError(t, err)
			ops, ok := resp.([]interface{})
			require.True(t, ok)

			// should be a 4 element array
			assert.Equal(t, len(ops), 4)

			fourth, ok := ops[3].([]interface{})
			require.True(t, ok)

			// fourth element has transactions
			assert.Greater(t, len(fourth), 0)
		})
}
