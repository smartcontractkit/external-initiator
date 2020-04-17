package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetXtzMonitorResponse(t *testing.T) {
	t.Run("creates mock XtzMonitorResponse",
		func(t *testing.T) {
			resp, err := getXtzMonitorResponse("chain_idisnotmeaningfulyet")
			assert.Nil(t, err)
			assert.Equal(t, resp.Hash, "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D")
		})
}

func TestGetXtzOperationsResponse(t *testing.T) {
	t.Run("creates an appropriately structured mock Tezos block",
		func(t *testing.T) {
			resp, err := getXtzOperationsResponse("block_idisnotmeaningfulyet")

			assert.Nil(t, err)

			// should be a 4 element array
			assert.Equal(t, len(resp), 4)

			// first three elements are empty (in the mock, not in
			// real blocks)
			assert.Equal(t, len(resp[0]), 0)
			assert.Equal(t, len(resp[1]), 0)
			assert.Equal(t, len(resp[2]), 0)

			// fourth element has transactions
			assert.Greater(t, len(resp[3]), 0)
		})
}
