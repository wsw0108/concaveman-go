package concaveman

import "math"

const (
	epsilon        = 1.1102230246251565e-16
	splitter       = 134217729
	resulterrbound = (3 + 8*epsilon) * epsilon
)

// fast_expansion_sum_zeroelim routine from oritinal code
func sum(elen int, e []float64, flen int, f []float64, h []float64) int {
	var Q, Qnew, hh, bvirt float64
	enow := e[0]
	fnow := f[0]
	eindex := 0
	findex := 0
	if (fnow > enow) == (fnow > -enow) {
		Q = enow
		eindex++
		enow = e[eindex]
	} else {
		Q = fnow
		findex++
		fnow = f[findex]
	}
	hindex := 0
	if eindex < elen && findex < flen {
		if (fnow > enow) == (fnow > -enow) {
			Qnew = enow + Q
			hh = Q - (Qnew - enow)
			eindex++
			enow = e[eindex]
		} else {
			Qnew = fnow + Q
			hh = Q - (Qnew - fnow)
			findex++
			fnow = f[findex]
		}
		Q = Qnew
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
		for eindex < elen && findex < flen {
			if (fnow > enow) == (fnow > -enow) {
				Qnew = Q + enow
				bvirt = Qnew - Q
				hh = Q - (Qnew - bvirt) + (enow - bvirt)
				eindex++
				enow = e[eindex]
			} else {
				Qnew = Q + fnow
				bvirt = Qnew - Q
				hh = Q - (Qnew - bvirt) + (fnow - bvirt)
				findex++
				fnow = f[findex]
			}
			Q = Qnew
			if hh != 0 {
				h[hindex] = hh
				hindex++
			}
		}
	}
	for eindex < elen {
		Qnew = Q + enow
		bvirt = Qnew - Q
		hh = Q - (Qnew - bvirt) + (enow - bvirt)
		eindex++
		enow = e[eindex]
		Q = Qnew
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
	}
	for findex < flen {
		Qnew = Q + fnow
		bvirt = Qnew - Q
		hh = Q - (Qnew - bvirt) + (fnow - bvirt)
		findex++
		fnow = f[findex]
		Q = Qnew
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
	}
	if Q != 0 || hindex == 0 {
		h[hindex] = Q
		hindex++
	}
	return hindex
}

func estimate(elen int, e []float64) float64 {
	Q := e[0]
	for i := 1; i < elen; i++ {
		Q += e[i]
	}
	return Q
}

const (
	ccwerrboundA = (3 + 16*epsilon) * epsilon
	ccwerrboundB = (2 + 12*epsilon) * epsilon
	ccwerrboundC = (9 + 64*epsilon) * epsilon * epsilon
)

var (
	B  [4]float64
	C1 [8]float64
	C2 [12]float64
	D  [16]float64
	u  [4]float64
)

func orient2Dadapt(ax, ay, bx, by, cx, cy, detsum float64) float64 {
	var acxtail, acytail, bcxtail, bcytail float64
	var bvirt, c, ahi, alo, bhi, blo, _i, _j, _0, s1, s0, t1, t0, u3 float64

	acx := ax - cx
	bcx := bx - cx
	acy := ay - cy
	bcy := by - cy

	s1 = acx * bcy
	c = splitter * acx
	ahi = c - (c - acx)
	alo = acx - ahi
	c = splitter * bcy
	bhi = c - (c - bcy)
	blo = bcy - bhi
	s0 = alo*blo - (s1 - ahi*bhi - alo*bhi - ahi*blo)
	t1 = acy * bcx
	c = splitter * acy
	ahi = c - (c - acy)
	alo = acy - ahi
	c = splitter * bcx
	bhi = c - (c - bcx)
	blo = bcx - bhi
	t0 = alo*blo - (t1 - ahi*bhi - alo*bhi - ahi*blo)
	_i = s0 - t0
	bvirt = s0 - _i
	B[0] = s0 - (_i + bvirt) + (bvirt - t0)
	_j = s1 + _i
	bvirt = _j - s1
	_0 = s1 - (_j - bvirt) + (_i - bvirt)
	_i = _0 - t1
	bvirt = _0 - _i
	B[1] = _0 - (_i + bvirt) + (bvirt - t1)
	u3 = _j + _i
	bvirt = u3 - _j
	B[2] = _j - (u3 - bvirt) + (_i - bvirt)
	B[3] = u3

	det := estimate(4, B[:])
	errbound := ccwerrboundB * detsum
	if det >= errbound || -det >= errbound {
		return det
	}

	bvirt = ax - acx
	acxtail = ax - (acx + bvirt) + (bvirt - cx)
	bvirt = bx - bcx
	bcxtail = bx - (bcx + bvirt) + (bvirt - cx)
	bvirt = ay - acy
	acytail = ay - (acy + bvirt) + (bvirt - cy)
	bvirt = by - bcy
	bcytail = by - (bcy + bvirt) + (bvirt - cy)

	if acxtail == 0 && acytail == 0 && bcxtail == 0 && bcytail == 0 {
		return det
	}

	errbound = ccwerrboundC*detsum + resulterrbound*math.Abs(det)
	det += (acx*bcytail + bcy*acxtail) - (acy*bcxtail + bcx*acytail)
	if det >= errbound || -det >= errbound {
		return det
	}

	s1 = acxtail * bcy
	c = splitter * acxtail
	ahi = c - (c - acxtail)
	alo = acxtail - ahi
	c = splitter * bcy
	bhi = c - (c - bcy)
	blo = bcy - bhi
	s0 = alo*blo - (s1 - ahi*bhi - alo*bhi - ahi*blo)
	t1 = acytail * bcx
	c = splitter * acytail
	ahi = c - (c - acytail)
	alo = acytail - ahi
	c = splitter * bcx
	bhi = c - (c - bcx)
	blo = bcx - bhi
	t0 = alo*blo - (t1 - ahi*bhi - alo*bhi - ahi*blo)
	_i = s0 - t0
	bvirt = s0 - _i
	u[0] = s0 - (_i + bvirt) + (bvirt - t0)
	_j = s1 + _i
	bvirt = _j - s1
	_0 = s1 - (_j - bvirt) + (_i - bvirt)
	_i = _0 - t1
	bvirt = _0 - _i
	u[1] = _0 - (_i + bvirt) + (bvirt - t1)
	u3 = _j + _i
	bvirt = u3 - _j
	u[2] = _j - (u3 - bvirt) + (_i - bvirt)
	u[3] = u3
	C1len := sum(4, B[:], 4, u[:], C1[:])

	s1 = acx * bcytail
	c = splitter * acx
	ahi = c - (c - acx)
	alo = acx - ahi
	c = splitter * bcytail
	bhi = c - (c - bcytail)
	blo = bcytail - bhi
	s0 = alo*blo - (s1 - ahi*bhi - alo*bhi - ahi*blo)
	t1 = acy * bcxtail
	c = splitter * acy
	ahi = c - (c - acy)
	alo = acy - ahi
	c = splitter * bcxtail
	bhi = c - (c - bcxtail)
	blo = bcxtail - bhi
	t0 = alo*blo - (t1 - ahi*bhi - alo*bhi - ahi*blo)
	_i = s0 - t0
	bvirt = s0 - _i
	u[0] = s0 - (_i + bvirt) + (bvirt - t0)
	_j = s1 + _i
	bvirt = _j - s1
	_0 = s1 - (_j - bvirt) + (_i - bvirt)
	_i = _0 - t1
	bvirt = _0 - _i
	u[1] = _0 - (_i + bvirt) + (bvirt - t1)
	u3 = _j + _i
	bvirt = u3 - _j
	u[2] = _j - (u3 - bvirt) + (_i - bvirt)
	u[3] = u3
	C2len := sum(C1len, C1[:], 4, u[:], C2[:])

	s1 = acxtail * bcytail
	c = splitter * acxtail
	ahi = c - (c - acxtail)
	alo = acxtail - ahi
	c = splitter * bcytail
	bhi = c - (c - bcytail)
	blo = bcytail - bhi
	s0 = alo*blo - (s1 - ahi*bhi - alo*bhi - ahi*blo)
	t1 = acytail * bcxtail
	c = splitter * acytail
	ahi = c - (c - acytail)
	alo = acytail - ahi
	c = splitter * bcxtail
	bhi = c - (c - bcxtail)
	blo = bcxtail - bhi
	t0 = alo*blo - (t1 - ahi*bhi - alo*bhi - ahi*blo)
	_i = s0 - t0
	bvirt = s0 - _i
	u[0] = s0 - (_i + bvirt) + (bvirt - t0)
	_j = s1 + _i
	bvirt = _j - s1
	_0 = s1 - (_j - bvirt) + (_i - bvirt)
	_i = _0 - t1
	bvirt = _0 - _i
	u[1] = _0 - (_i + bvirt) + (bvirt - t1)
	u3 = _j + _i
	bvirt = u3 - _j
	u[2] = _j - (u3 - bvirt) + (_i - bvirt)
	u[3] = u3
	Dlen := sum(C2len, C2[:], 4, u[:], D[:])

	return D[Dlen-1]
}

func Orient2D(ax, ay, bx, by, cx, cy float64) float64 {
	detleft := (ay - cy) * (bx - cx)
	detright := (ax - cx) * (by - cy)
	det := detleft - detright

	detsum := math.Abs(detleft + detright)
	if math.Abs(det) >= ccwerrboundA*detsum {
		return det
	}

	return -orient2Dadapt(ax, ay, bx, by, cx, cy, detsum)
}
