package client

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewClientWithOpts(t *testing.T) {
	testCases := []struct {
		expectedError            error
		expectedSuppressWarnings bool
		expectedVerbose          bool
		expectedErrorOnWarnings  bool
		opts                     []Opt
	}{
		{
			expectedError:            nil,
			expectedSuppressWarnings: false,
			expectedVerbose:          false,
			expectedErrorOnWarnings:  false,
			opts:                     []Opt{},
		},
		{
			expectedError:            nil,
			expectedSuppressWarnings: true,
			expectedVerbose:          false,
			expectedErrorOnWarnings:  false,
			opts:                     []Opt{WithSuppressWarnings()},
		},
		{
			expectedError:            nil,
			expectedSuppressWarnings: false,
			expectedVerbose:          true,
			expectedErrorOnWarnings:  false,
			opts:                     []Opt{WithVerboseOutput()},
		},
		{
			expectedError:            nil,
			expectedSuppressWarnings: false,
			expectedVerbose:          false,
			expectedErrorOnWarnings:  true,
			opts:                     []Opt{WithErrorOnWarning()},
		},
		{
			expectedError:            nil,
			expectedSuppressWarnings: true,
			expectedVerbose:          false,
			expectedErrorOnWarnings:  true,
			opts:                     []Opt{WithErrorOnWarning(), WithSuppressWarnings()},
		},
	}
	for _, tc := range testCases {
		client, err := NewClient(tc.opts...)
		assert.Check(t, is.Equal(err, tc.expectedError))
		assert.Check(t, is.Equal(client.errorOnWarning, tc.expectedErrorOnWarnings))
		assert.Check(t, is.Equal(client.verbose, tc.expectedVerbose))
		assert.Check(t, is.Equal(client.suppressWarnings, tc.expectedSuppressWarnings))
	}
}
