package main

import "testing"

func Test_setUnhealthy(t *testing.T) {
	tests := []struct {
		name          string
		wantMessageID string
		wantErr       bool
	}{
		{
			name:          "happy test",
			wantMessageID: "",
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMessageID, err := setUnhealthy()
			if (err != nil) != tt.wantErr {
				t.Errorf("setUnhealthy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMessageID != tt.wantMessageID {
				t.Errorf("setUnhealthy() = %v, want %v", gotMessageID, tt.wantMessageID)
			}
		})
	}
}
