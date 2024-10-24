package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
)

func Test_buildPutItem(t *testing.T) {
	type args struct {
		eventBody string
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
			args: args{eventBody: `{"name":"doctor", "email": "test@example.com", "phone": "0123456", "password": "` + testPassword + `"}`},
			want: map[string]types.AttributeValue{
				"name":  &types.AttributeValueMemberS{Value: "doctor"},
				"email": &types.AttributeValueMemberS{Value: "test@example.com"},
				"phone": &types.AttributeValueMemberS{Value: "0123456"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPutItem(tt.args.eventBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildPutItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, field := range []string{"name", "email", "phone"} {
				if !reflect.DeepEqual(got[field], tt.want[field]) {
					t.Errorf("buildPutItem() field '%v' got = %v, want %v", field, got[field], tt.want[field])
					return
				}
			}

			for _, field := range []string{"id", "createdAt"} {
				if got[field] == nil {
					t.Errorf("buildPutItem() field %v is nil", field)
					return
				}
			}

		})
	}
}
