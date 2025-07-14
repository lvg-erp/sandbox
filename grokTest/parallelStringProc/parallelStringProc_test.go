package main

import (
	"context"
	"reflect"
	"testing"
)

func TestParallelStringProc(t *testing.T) {
	test := []struct {
		name    string
		input   []string
		workers int
		want    []string
		wantErr bool
		wantCnt int
	}{
		{
			name:    "ToUpperString",
			input:   []string{"hello", "world"},
			workers: 2,
			want:    []string{"HELLO", "WORLD"},
			wantErr: false,
			wantCnt: 2,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			gotWant, err, cnt := processString(context.Background(), tt.input, tt.workers)
			if (err != nil) != tt.wantErr {
				t.Errorf("processString() got = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotWant, tt.want) {
				t.Errorf("processString() got = %v, want %v", gotWant, tt.want)
			}
			if cnt != tt.wantCnt {
				t.Errorf("processString() got = %v, want %v", cnt, tt.wantCnt)
			}
		})
	}

}
