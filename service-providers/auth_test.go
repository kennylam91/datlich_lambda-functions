package main

import (
	"reflect"
	"testing"
)

func Test_getBasicCredential(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name    string
		args    args
		want    BasicCredential
		wantErr bool
	}{
		{
			name:    "case ok",
			args:    args{body: `{"username": "test", "password": "pass123@"}`},
			want:    BasicCredential{"test", "pass123@"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBasicCredential(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBasicCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBasicCredential() got = %v, want %v", got, tt.want)
			}
		})
	}
}
