package main

import (
	"reflect"
	"testing"
)

func TestSymbolFrequency(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		workers int
		want    []pair
	}{
		{
			name:    "workers_2",
			input:   []string{"hello", "world", "change"},
			workers: 2,
			want: []pair{
				{Key: "l", Value: 3},
				{Key: "o", Value: 2},
				{Key: "h", Value: 2},
				{Key: "e", Value: 2},
				{Key: "w", Value: 1},
				{Key: "r", Value: 1},
				{Key: "n", Value: 1},
				{Key: "g", Value: 1},
				{Key: "d", Value: 1},
				{Key: "c", Value: 1},
				{Key: "a", Value: 1},
			},
		},
		// TODO: добавить другие тестовые случаи
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapRune := countFrequency(tt.input, tt.workers)
			mapString := mapRuneToMapString(mapRune)
			gotPairs := sortingMapByValue(mapString)
			if !reflect.DeepEqual(gotPairs, tt.want) {
				t.Errorf("count frequency(%v, %d) = %v, want %v", tt.input, tt.workers, gotPairs, tt.want)
			}
		})
	}
}
