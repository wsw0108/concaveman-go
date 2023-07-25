package predicates_test

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/wsw0108/concaveman-go/predicates"
)

func TestOrient2D(t *testing.T) {
	{
		v := predicates.Orient2D(0, 0, 1, 1, 0, 1)
		if v >= 0 {
			t.Error("clockwise")
		}
	}
	{
		v := predicates.Orient2D(0, 0, 0, 1, 1, 1)
		if v <= 0 {
			t.Error("counterclockwise")
		}
	}
	{
		v := predicates.Orient2D(0, 0, 0.5, 0.5, 1, 1)
		if v != 0 {
			t.Error("collinear")
		}
	}

	// {
	// 	r := 0.95
	// 	q := 18.0
	// 	p := 16.8
	// 	w := math.Pow(2, -43)

	// 	for i := 0; i < 128; i++ {
	// 		for j := 0; j < 128; j++ {
	// 			x := r + w*float64(i)/128.0
	// 			y := r + w*float64(j)/128.0

	// 			o := concaveman.Orient2D(x, y, q, q, p, p)
	// 			// var o2 = robustOrientation[3]([x, y], [q, q], [p, p])

	// 			// if (Math.sign(o) !== Math.sign(o2)) {
	// 			//     assert.fail(`${x},${y}, ${q},${q}, ${p},${p}: ${o} vs ${o2}`);
	// 			// }
	// 		}
	// 	}
	// 	// 512x512 near-collinear
	// }

	{
		f, _ := os.Open("testdata/orient2d.txt")
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			parts := strings.Split(line, " ")
			if len(parts) != 8 {
				continue
			}
			ax, _ := strconv.ParseFloat(parts[1], 64)
			ay, _ := strconv.ParseFloat(parts[2], 64)
			bx, _ := strconv.ParseFloat(parts[3], 64)
			by, _ := strconv.ParseFloat(parts[4], 64)
			cx, _ := strconv.ParseFloat(parts[5], 64)
			cy, _ := strconv.ParseFloat(parts[6], 64)
			sign, _ := strconv.ParseInt(parts[7], 10, 32)
			result := predicates.Orient2D(ax, ay, bx, by, cx, cy)
			if math.Signbit(result) != (sign > 0) {
				t.Errorf("%s: %f vs %d\n", line, result, -sign)
			}
		}
		// 1000 hard fixtures
	}
}
