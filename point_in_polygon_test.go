package concaveman_test

import (
	"testing"

	"github.com/wsw0108/concaveman"
)

func TestPointInPolygon(t *testing.T) {
	type args struct {
		point concaveman.Point
		poly  []concaveman.Point
	}
	box := []concaveman.Point{
		{1, 1},
		{1, 2},
		{2, 2},
		{2, 1},
	}
	flag := []concaveman.Point{
		{1, 1},
		{10, 1},
		{5, 5},
		{10, 10},
		{1, 10},
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "box/1",
			args: args{
				point: concaveman.Point{1.5, 1.5},
				poly:  box,
			},
			want: true,
		},
		{
			name: "box/2",
			args: args{
				point: concaveman.Point{1.2, 1.9},
				poly:  box,
			},
			want: true,
		},
		{
			name: "box/3",
			args: args{
				point: concaveman.Point{0, 1.9},
				poly:  box,
			},
			want: false,
		},
		{
			name: "box/4",
			args: args{
				point: concaveman.Point{1.5, 2},
				poly:  box,
			},
			want: false,
		},
		{
			name: "box/5",
			args: args{
				point: concaveman.Point{1.5, 2.2},
				poly:  box,
			},
			want: false,
		},
		{
			name: "box/6",
			args: args{
				point: concaveman.Point{3, 5},
				poly:  box,
			},
			want: false,
		},
		{
			name: "flag/1",
			args: args{
				point: concaveman.Point{2, 5},
				poly:  flag,
			},
			want: true,
		},
		{
			name: "flag/2",
			args: args{
				point: concaveman.Point{3, 5},
				poly:  flag,
			},
			want: true,
		},
		{
			name: "flag/3",
			args: args{
				point: concaveman.Point{4, 5},
				poly:  flag,
			},
			want: true,
		},
		{
			name: "flag/4",
			args: args{
				point: concaveman.Point{10, 5},
				poly:  flag,
			},
			want: false,
		},
		{
			name: "flag/5",
			args: args{
				point: concaveman.Point{11, 5},
				poly:  flag,
			},
			want: false,
		},
		{
			name: "flag/6",
			args: args{
				point: concaveman.Point{9, 5},
				poly:  flag,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := concaveman.PointInPolygon(tt.args.point, tt.args.poly); got != tt.want {
				t.Errorf("PointInPolygon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointInPolygonOffset(t *testing.T) {
	type args struct {
		point concaveman.Point
		poly  []concaveman.Point
		start int
		end   int
	}
	box := []concaveman.Point{
		{100, 101},
		{102, 103},
		{1, 1},
		{1, 2},
		{2, 2},
		{2, 1},
		{200, 201},
	}
	flag := []concaveman.Point{
		{100, 101},
		{1, 1},
		{10, 1},
		{5, 5},
		{10, 10},
		{1, 10},
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "box/1",
			args: args{
				point: concaveman.Point{1.5, 1.5},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: true,
		},
		{
			name: "box/2",
			args: args{
				point: concaveman.Point{1.2, 1.9},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: true,
		},
		{
			name: "box/3",
			args: args{
				point: concaveman.Point{0, 1.9},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: false,
		},
		{
			name: "box/4",
			args: args{
				point: concaveman.Point{1.5, 2},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: false,
		},
		{
			name: "box/5",
			args: args{
				point: concaveman.Point{1.5, 2.2},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: false,
		},
		{
			name: "box/6",
			args: args{
				point: concaveman.Point{3, 5},
				poly:  box,
				start: 2,
				end:   6,
			},
			want: false,
		},
		{
			name: "flag/1",
			args: args{
				point: concaveman.Point{2, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: true,
		},
		{
			name: "flag/2",
			args: args{
				point: concaveman.Point{3, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: true,
		},
		{
			name: "flag/3",
			args: args{
				point: concaveman.Point{4, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: true,
		},
		{
			name: "flag/4",
			args: args{
				point: concaveman.Point{10, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: false,
		},
		{
			name: "flag/5",
			args: args{
				point: concaveman.Point{11, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: false,
		},
		{
			name: "flag/6",
			args: args{
				point: concaveman.Point{9, 5},
				poly:  flag,
				start: 1,
				end:   len(flag),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := concaveman.PointInPolygonOffset(tt.args.point, tt.args.poly, tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("PointInPolygonOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}
