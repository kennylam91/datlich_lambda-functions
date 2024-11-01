package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
	"time"
)

func Test_buildPutItem(t *testing.T) {
	type args struct {
		provider *ServiceProvider
	}
	testPassword := "pass"
	var tests = []struct {
		name    string
		args    args
		want    map[string]types.AttributeValue
		wantErr bool
	}{
		{
			name: "valid body",
			args: args{
				provider: &ServiceProvider{"username", "my name", "test@example.com", "0123456", testPassword, time.Now()},
			},
			want: map[string]types.AttributeValue{
				"username": &types.AttributeValueMemberS{Value: "username"},
				"name":     &types.AttributeValueMemberS{Value: "my name"},
				"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
				"phone":    &types.AttributeValueMemberS{Value: "0123456"},
			},
			wantErr: false,
		},
		{
			name:    "missing required fields",
			args:    args{provider: &ServiceProvider{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid username",
			args:    args{provider: &ServiceProvider{"user name", "name", "test123@gmail.com", "", "pass", time.Now()}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPutItem(tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildPutItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, field := range []string{"username", "name", "email", "phone"} {
					if !reflect.DeepEqual(got[field], tt.want[field]) {
						t.Errorf("buildPutItem() field '%v' got = %v, want %v", field, got[field], tt.want[field])
						return
					}
				}

				for _, field := range []string{"createdAt"} {
					if got[field] == nil {
						t.Errorf("buildPutItem() field %v is nil", field)
						return
					}
				}
			}

		})
	}
}
