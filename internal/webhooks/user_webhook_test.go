package webhooks

import "testing"

func Test_validName(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "only small alphanumeric",
			args: args{
				v: "hello",
			},
			want: true,
		},
		{
			name: "only small alphanumeric and -",
			args: args{
				v: "hello-world",
			},
			want: true,
		},
		{
			name: "endwith -",
			args: args{
				v: "hello-world-",
			},
			want: false,
		},
		{
			name: "startwith -",
			args: args{
				v: "-hello-world",
			},
			want: false,
		},
		{
			name: "contain .",
			args: args{
				v: "hello.world",
			},
			want: false,
		},
		{
			name: "capital",
			args: args{
				v: "helloWorld",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validName(tt.args.v); got != tt.want {
				t.Errorf("validName() = %v, want %v", got, tt.want)
			}
		})
	}
}
