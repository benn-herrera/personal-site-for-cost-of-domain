package main

import (
	"strings"
	"testing"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func TestFmtCoord(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{5, "5"},
		{5.5, "5.5"},
		{-3.25, "-3.25"},
		{0.0001, "0.0001"},     // %g would render this as 1e-04
		{10000000, "10000000"}, // %g would render this as 1e+07
	}
	for _, c := range cases {
		got := fmtCoord(c.in)
		if got != c.want {
			t.Errorf("fmtCoord(%v) = %q, want %q", c.in, got, c.want)
		}
		if strings.ContainsAny(got, "eE") {
			t.Errorf("fmtCoord(%v) = %q contains scientific notation", c.in, got)
		}
	}
}

func TestTransformPt(t *testing.T) {
	// rawX/rawY are 26.6 fixed point: value 128 == 2.0, 64 == 1.0
	x, y := transformPt(fixed.Int26_6(128), fixed.Int26_6(64), 2.0, 10.0, 20.0)
	// x = xOffset + (128/64)*scale = 10 + 2*2 = 14
	// y = baseline - (64/64)*scale = 20 - 1*2 = 18
	if x != 14 || y != 18 {
		t.Errorf("transformPt = (%v, %v), want (14, 18)", x, y)
	}
}

func TestSegmentsToSVGPath(t *testing.T) {
	t.Run("move + line closes with Z", func(t *testing.T) {
		segs := []sfnt.Segment{
			{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{{X: 0, Y: 0}}},
			{Op: sfnt.SegmentOpLineTo, Args: [3]fixed.Point26_6{{X: 64, Y: 0}}},
		}
		got := segmentsToSVGPath(segs, 1.0, 0.0, 10.0)
		if got != "M0 10L1 10Z" {
			t.Errorf("got %q, want %q", got, "M0 10L1 10Z")
		}
	})

	t.Run("quadratic segment", func(t *testing.T) {
		segs := []sfnt.Segment{
			{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{{X: 0, Y: 0}}},
			{Op: sfnt.SegmentOpQuadTo, Args: [3]fixed.Point26_6{{X: 64, Y: 64}, {X: 128, Y: 0}}},
		}
		got := segmentsToSVGPath(segs, 1.0, 0.0, 10.0)
		if got != "M0 10Q1 9 2 10Z" {
			t.Errorf("got %q, want %q", got, "M0 10Q1 9 2 10Z")
		}
	})

	t.Run("multiple subpaths each get a Z", func(t *testing.T) {
		segs := []sfnt.Segment{
			{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{{X: 0, Y: 0}}},
			{Op: sfnt.SegmentOpLineTo, Args: [3]fixed.Point26_6{{X: 64, Y: 0}}},
			{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{{X: 0, Y: 0}}},
			{Op: sfnt.SegmentOpLineTo, Args: [3]fixed.Point26_6{{X: 64, Y: 0}}},
		}
		got := segmentsToSVGPath(segs, 1.0, 0.0, 10.0)
		if strings.Count(got, "Z") != 2 {
			t.Errorf("expected 2 subpaths (2 Z), got %q", got)
		}
		if !strings.HasPrefix(got, "M") || !strings.HasSuffix(got, "Z") {
			t.Errorf("path should start with M and end with Z: %q", got)
		}
	})
}
