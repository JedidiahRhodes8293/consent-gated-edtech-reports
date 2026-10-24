package main

import (
	"reflect"
	"testing"
)

func TestReportingDecision(t *testing.T) {
	input := []Deadline{{Course: "algebra", Learner: "learner-17", DueDate: "2026-11-03"}}
	tests := []struct {
		name    string
		granted bool
		want    []Deadline
		wantErr bool
	}{
		{name: "consented educator sees deadlines", granted: true, want: input},
		{name: "revoked learner is withheld", granted: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReportingDecision(tt.granted, input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("deadlines = %#v, want %#v", got, tt.want)
			}
		})
	}
}
