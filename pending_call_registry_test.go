package enigma

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPendingCallRegistry(t *testing.T) {

	ctx := context.Background()

	// Setup call registry
	pcr := newPendingCallRegistry()

	// Register calls
	pc1 := pcr.registerPendingCall(ctx)
	ctxWithReservedRequestID, reservedRequestID := pcr.WithReservedRequestID(ctx)
	pc2 := pcr.registerPendingCall(ctx)
	pc3 := pcr.registerPendingCall(ctxWithReservedRequestID)

	// Check that the third registered call actually uses the reserved request id
	assert.Equal(t, reservedRequestID, pc3.ID)

	// Extract pending calls
	extractedpc2 := pcr.removePendingCall(pc2.ID)
	assert.Equal(t, pc2, extractedpc2)

	extractedpc1 := pcr.removePendingCall(pc1.ID)
	assert.Equal(t, pc1, extractedpc1)

	extractedpc3 := pcr.removePendingCall(pc3.ID)
	assert.Equal(t, pc3, extractedpc3)

	// Extract twice should return nil
	doubleExtracted1 := pcr.removePendingCall(pc1.ID)
	assert.Nil(t, doubleExtracted1)

}
