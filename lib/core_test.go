package autosetip

import (
	"github.com/alibabacloud-go/tea/tea"
	"reflect"
	"testing"
)

func TestEcsTarget_portRange(t1 *testing.T) {
	type fields struct {
		Region          string
		AccessKey       string
		SecretKey       string
		Endpoint        string
		SecurityGroupId string
		Port            []string
	}
	tests := []struct {
		name   string
		fields fields
		want   *string
	}{
		{
			name: "portRange should join Port array",
			fields: fields{
				Region:          "",
				AccessKey:       "",
				SecretKey:       "",
				Endpoint:        "",
				SecurityGroupId: "",
				Port:            []string{"80", "443"},
			},
			want: tea.String("80/80,443/443"),
		},
		{
			name: "portRange should use default port",
			fields: fields{
				Region:          "",
				AccessKey:       "",
				SecretKey:       "",
				Endpoint:        "",
				SecurityGroupId: "",
				Port:            []string{},
			},
			want: tea.String("22/22"),
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := EcsTarget{
				Region:          tt.fields.Region,
				AccessKey:       tt.fields.AccessKey,
				SecretKey:       tt.fields.SecretKey,
				Endpoint:        tt.fields.Endpoint,
				SecurityGroupId: tt.fields.SecurityGroupId,
				Port:            tt.fields.Port,
			}
			if got := t.portRange(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("portRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
