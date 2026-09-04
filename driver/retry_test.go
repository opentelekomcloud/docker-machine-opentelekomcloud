package opentelekomcloud

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", want: false},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, want: true},
		{name: "wrapped unexpected EOF", err: fmt.Errorf("GET response: %w", io.ErrUnexpectedEOF), want: true},
		{name: "string only", err: errors.New("response body: unexpected EOF"), want: true},
		{name: "plain EOF", err: io.EOF, want: false},
		{name: "other error", err: errors.New("409 Conflict"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, isUnexpectedEOF(test.err))
		})
	}
}

func TestRetryOnUnexpectedEOF(t *testing.T) {
	attempts := 0
	resets := 0
	err := retryOnUnexpectedEOF("GET EIP", func() { resets++ }, func() error {
		attempts++
		if attempts == 1 {
			return io.ErrUnexpectedEOF
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, 1, resets)
}

func TestRetryOnUnexpectedEOFDoesNotRetryOtherErrors(t *testing.T) {
	attempts := 0
	resets := 0
	wantErr := errors.New("409 Conflict")
	err := retryOnUnexpectedEOF("GET EIP", func() { resets++ }, func() error {
		attempts++
		return wantErr
	})

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, attempts)
	assert.Zero(t, resets)
}

func TestRetryOnUnexpectedEOFRetriesOnlyOnce(t *testing.T) {
	attempts := 0
	resets := 0
	err := retryOnUnexpectedEOF("GET EIP", func() { resets++ }, func() error {
		attempts++
		return io.ErrUnexpectedEOF
	})

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, 1, resets)
}
