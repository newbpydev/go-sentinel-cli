package models

import "testing"

func TestAPITestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   APITestEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: APITestEvent{
				Action:  "pass",
				Package: "pkg/foo",
				Test:    "TestA",
				Time:    "2023-01-01T00:00:00Z",
				Elapsed: 0.1,
				Output:  "test output",
			},
			wantErr: false,
		},
		{
			name: "missing action",
			event: APITestEvent{
				Package: "pkg/foo",
				Test:    "TestA",
			},
			wantErr: true,
			errMsg:  "action is required",
		},
		{
			name: "missing package",
			event: APITestEvent{
				Action: "pass",
				Test:   "TestA",
			},
			wantErr: true,
			errMsg:  "package is required",
		},
		{
			name: "optional fields missing",
			event: APITestEvent{
				Action:  "pass",
				Package: "pkg/foo",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != tc.errMsg {
					t.Errorf("expected error message %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
