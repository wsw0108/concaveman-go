package concaveman_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/wsw0108/concaveman"
)

var (
	g_points []concaveman.Point
	g_hull   []concaveman.Point
	g_hull2  []concaveman.Point
)

func fillData(filename string, data *[]concaveman.Point) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	var err error
	err = fillData("testdata/points-1k.json", &g_points)
	if err != nil {
		panic(err)
	}
	err = fillData("testdata/points-1k-hull.json", &g_hull)
	if err != nil {
		panic(err)
	}
	err = fillData("testdata/points-1k-hull2.json", &g_hull2)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestConcaveHull(t *testing.T) {
}

func TestDefaultConcaveHull(t *testing.T) {
	result := concaveman.Concaveman(g_points)
	if !reflect.DeepEqual(result, g_hull) {
		t.Fail()
	}
}

func TestTunedConcaveHull(t *testing.T) {
	opt := concaveman.Options{
		Concavity:       3,
		LengthThreshold: 0.01,
	}
	result := concaveman.Concaveman(g_points, opt)
	if !reflect.DeepEqual(result, g_hull2) {
		t.Fail()
	}
}
