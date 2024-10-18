package nodeinfo

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_getLastSystemBootTime(t *testing.T) {
	tests := []struct {
		name    string
		want    time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	tm, err := LastSystemBootTime()
	assert.NoError(t, err)
	t.Logf(tm.Format(time.RFC3339))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LastSystemBootTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("LastSystemBootTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastSystemBootTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}
