package fibonacci

import "testing"

func TestCalculate(t *testing.T) {
	type args struct {
		n uint
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "N=0",
			args: args{n: 0},
			want: 0,
		},
		{
			name: "N=1",
			args: args{n: 1},
			want: 1,
		},
		{
			name: "N=3",
			args: args{n: 3},
			want: 2,
		},
		{
			name: "N=10",
			args: args{n: 10},
			want: 55,
		},
		{
			name: "N=20",
			args: args{n: 20},
			want: 6_765,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Calculate(tt.args.n); got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
