package gerrors_test

import (
	"testing"

	"github.com/seinshah/gerrors"
	"google.golang.org/grpc/codes"
)

func TestDefaultCoreCallback(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		code               gerrors.Code
		expectedCode       gerrors.Code
		expectedIdentifier string
		expectedGrpcCode   codes.Code
	}{
		{
			name:               "existing code",
			code:               gerrors.NotFound,
			expectedCode:       gerrors.NotFound,
			expectedIdentifier: "not-found",
			expectedGrpcCode:   codes.NotFound,
		},
		{
			name:               "non existing code",
			code:               gerrors.Code(1000),
			expectedCode:       gerrors.Unknown,
			expectedIdentifier: "unknown",
			expectedGrpcCode:   codes.Unknown,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			core := gerrors.NewMapper(gerrors.Unknown, gerrors.GetDefaultMapping()).Lookup(tc.code)

			if core.GetInternalCode() != tc.expectedCode {
				t.Errorf("expected code %d, got %d", tc.expectedCode, core.GetInternalCode())
			}

			if core.GetIdentifier() != tc.expectedIdentifier {
				t.Errorf("expected identifier %s, got %s", tc.expectedIdentifier, core.GetIdentifier())
			}

			if len(core.GetDefaultMessage()) < 1 {
				t.Errorf("expected core to have default message: %v", core)
			}

			if coreg, ok := core.(gerrors.CoreGRPCError); ok {
				if coreg.GetGRPCCode() != tc.expectedGrpcCode {
					t.Errorf("expected grpc code %d, got %d", tc.expectedGrpcCode, coreg.GetGRPCCode())
				}
			} else {
				t.Errorf("expected error to implement CoreGrpcError interface: %v", core)
			}
		})
	}
}
