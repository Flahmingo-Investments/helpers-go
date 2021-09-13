package flog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSugaredLogger(t *testing.T) {
	testCases := []struct {
		name    string
		want    *zap.SugaredLogger
		wantErr bool
	}{
		{
			name:    "should return error: not initialized",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "should return the initialized sugared logger",
			want:    zap.NewExample().Sugar(),
			wantErr: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// sugar is global logger variable in the flog package
			sugar = tc.want

			got, err := SugaredLogger()

			if tc.wantErr {
				assert.Error(t, err)
				assert.IsType(t, ErrNotInitialized, err)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLogger(t *testing.T) {
	testCases := []struct {
		name    string
		want    *zap.Logger
		wantErr bool
	}{
		{
			name:    "should return error: not initialized",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "should return the initialized logger",
			want:    zap.NewExample(),
			wantErr: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// logger is global logger variable in the flog package
			logger = tc.want

			got, err := Logger()

			if tc.wantErr {
				assert.Error(t, err)
				assert.IsType(t, ErrNotInitialized, err)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tc.want, got)
		})
	}
}
