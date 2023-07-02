package zanzibar

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func Test_getStringSliceFromCSV(t *testing.T) {
	table := []struct {
		csv  string
		want []string
	}{
		{"Content-Type,host,X-Uber-Uuid",
			[]string{"Content-Type", "host", "X-Uber-Uuid"},
		},
		{"Content-Type, host , X-Uber-Uuid, ",
			[]string{"Content-Type", "host", "X-Uber-Uuid"},
		},
	}

	for i, tt := range table {
		t.Run("test"+strconv.Itoa(i), func(t *testing.T) {
			got := getStringSliceFromCSV(tt.csv)
			assert.Equal(t, tt.want, got)
		})
	}
}
