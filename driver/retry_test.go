package opentelekomcloud

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsEOFLikeError covers both the wrapped-error path (errors.Is) and the
// string-match fallback for errors that transports returned without wrapping.
func TestIsEOFLikeError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil is not EOF", nil, false},
		{"io.EOF direct", io.EOF, true},
		{"io.ErrUnexpectedEOF direct", io.ErrUnexpectedEOF, true},
		{"wrapped io.EOF", fmt.Errorf("wrapper: %w", io.EOF), true},
		{"wrapped io.ErrUnexpectedEOF", fmt.Errorf("http: %w", io.ErrUnexpectedEOF), true},
		{"string-only 'unexpected EOF'", errors.New("unexpected EOF"), true},
		{"string EOF anywhere", errors.New("failed to delete: EOF on stream"), true},
		{"arbitrary other error", errors.New("conflict: 409"), false},
		{"empty string error", errors.New(""), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isEOFLikeError(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestRetryOnEOF_successFirstTry confirms the common-path: no retry when
// the first attempt succeeds. Sleep must NOT fire.
func TestRetryOnEOF_successFirstTry(t *testing.T) {
	calls := 0
	err := retryOnEOF("noop", nil, func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls, "success on first try must not retry")
}

// TestRetryOnEOF_retriesOnEOFThenSucceeds is the scenario that motivated
// SDE-346: OTC's EIP release returns EOF on the first call, succeeds on
// the retry. The helper must swallow the first EOF and return nil.
func TestRetryOnEOF_retriesOnEOFThenSucceeds(t *testing.T) {
	calls := 0
	err := retryOnEOF("flaky", nil, func() error {
		calls++
		if calls == 1 {
			return io.ErrUnexpectedEOF
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, calls, "must retry exactly once on EOF")
}

// TestRetryOnEOF_doesNotRetryOnNonEOF prevents the helper from masking
// unrelated failures (e.g. 409 Conflict on a double-delete).
func TestRetryOnEOF_doesNotRetryOnNonEOF(t *testing.T) {
	calls := 0
	err := retryOnEOF("conflict", nil, func() error {
		calls++
		return errors.New("409 Conflict")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls, "non-EOF errors must not trigger a retry")
	assert.Contains(t, err.Error(), "Conflict")
}

// TestRetryOnEOF_giveUpAfterTwoConsecutiveEOFs ensures we don't loop
// forever — two strikes and we surface the error.
func TestRetryOnEOF_giveUpAfterTwoConsecutiveEOFs(t *testing.T) {
	calls := 0
	err := retryOnEOF("still-broken", nil, func() error {
		calls++
		return io.ErrUnexpectedEOF
	})
	assert.Error(t, err)
	assert.Equal(t, 2, calls, "must give up after one retry")
	assert.True(t, isEOFLikeError(err))
}

// TestRetryOnEOF_resetCallbackFiresExactlyOnceOnEOF is the key SDE-346
// invariant: the transport-reset hook must fire between attempt 1 and 2
// whenever EOF triggers a retry. It MUST NOT fire on non-EOF errors
// (so successful-first-try and 409-Conflict paths don't wastefully reset
// the HTTP pool).
func TestRetryOnEOF_resetCallbackFiresExactlyOnceOnEOF(t *testing.T) {
	resetCalls := 0
	fnCalls := 0
	err := retryOnEOF("eof-with-reset", func() { resetCalls++ }, func() error {
		fnCalls++
		if fnCalls == 1 {
			return io.ErrUnexpectedEOF
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, fnCalls)
	assert.Equal(t, 1, resetCalls, "reset must run exactly once, between the two attempts")
}

func TestRetryOnEOF_resetCallbackDoesNotFireOnSuccess(t *testing.T) {
	resetCalls := 0
	err := retryOnEOF("happy-path", func() { resetCalls++ }, func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, resetCalls)
}

func TestRetryOnEOF_resetCallbackDoesNotFireOnNonEOFError(t *testing.T) {
	resetCalls := 0
	err := retryOnEOF("non-eof", func() { resetCalls++ }, func() error {
		return errors.New("409 Conflict")
	})
	assert.Error(t, err)
	assert.Equal(t, 0, resetCalls,
		"non-EOF errors must not waste the HTTP pool by triggering a reset")
}
