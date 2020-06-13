package hapijson

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func Benchmark(b *testing.B) {

	var j []byte
	var e error
	// Setter
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf(" - %s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Set"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e = Set(j, "123456789", set.path...); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Getter
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Get"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e = Get(j, set.path...); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Remove
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		temp := make([]byte, len(j))
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Remove"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(temp, j)
				if _, e = Remove(temp, set.path...); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Clear
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Clear"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e = Clear(j, set.path...); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Merge
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Merge"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e = Merge(j, false, Path(set.path[:len(set.path)-1]...), set.path[len(set.path)-1], "ok"); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// SliceOf
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("SliceOf"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e = SliceOf(j, set.path...); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Valid
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Validate"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if e = Validate(j); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	// Minify
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Minify"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Minify(j)
			}
		})
	}
	// Prettify
	for _, set := range benchmarkSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Prettify"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Prettify(j, 2)
			}
		})
	}
	// Append
	for _, set := range benchmarkAppendAndIncrSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Append"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e := Append(j, Path(set.path...), set.before); e != nil {
					b.Fatal(e)
				}
			}
		})
	}

	// Incr
	for _, set := range benchmarkAppendAndIncrSet {
		fPath := set.updatingVal.(string)
		if j, e = ioutil.ReadFile(fPath); e != nil {
			b.Fatal(e)
		}
		name := fmt.Sprintf("-%s", fPath[strings.LastIndex(fPath, "/")+1:])
		b.Run("Incr"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, e := Incr(j, set.expect, set.updatedPath...); e != nil {
					b.Log("***", set.updatedPath, "****")
					b.Fatal(e)
				}
			}
		})
	}

}

var benchmarkSet = []TestSet{
	{
		updatingVal: "./testdata/test_500B.json",
		path:        []interface{}{"reviews", 0, "review", 2, "spoiler"},
	},
	{
		updatingVal: "./testdata/test_1K.json",
		path:        []interface{}{"reviews", 1, "review", 2, "vote"},
	},
	{
		updatingVal: "./testdata/test_2K.json",
		path:        []interface{}{"test merge no preserve", "why?", "reason2"},
	},
	{
		updatingVal: "./testdata/test_20K.json",
		path:        []interface{}{3, "true"},
	},
	{
		updatingVal: "./testdata/test_280K.json",
		path:        []interface{}{99, "number3"},
	},
	{
		updatingVal: "./testdata/test_2.8M.json",
		path:        []interface{}{"first", "html", "body", "iframe", "@data-bm"},
	},
	{
		updatingVal: "./testdata/test_8.6M.json",
		path:        []interface{}{"first", "html", "body", "iframe", "@data-bm"},
	},
}

var benchmarkAppendAndIncrSet = []TestSet{
	{
		updatingVal: "./testdata/test_500B.json",
		path:        []interface{}{"reviews", 0, "review"},
		before:      "benchmark Append",                               // for Append
		updatedPath: []interface{}{"reviews", 0, "review", 2, "time"}, // for incr
		expect:      1,                                                // for incr
	},
	{
		updatingVal: "./testdata/test_1K.json",
		path:        []interface{}{"reviews", 1, "review"},
		before:      "benchmark Append",                               // for Append
		updatedPath: []interface{}{"reviews", 1, "review", 2, "time"}, // for incr
		expect:      1,                                                // for incr
	},
	{
		updatingVal: "./testdata/test_2K.json",
		path:        []interface{}{"incr"},
		before:      "benchmark Append",       // for Append
		updatedPath: []interface{}{"incr", 4}, // for incr
		expect:      1,                        // for incr
	},
	{
		updatingVal: "./testdata/test_20K.json",
		path:        []interface{}{3, "strArray"},
		before:      "benchmark Append",              // for Append
		updatedPath: []interface{}{3, "intArray", 0}, // for incr
		expect:      1,                               // for incr
	},
	{
		updatingVal: "./testdata/test_280K.json",
		path:        []interface{}{99, "number3"},
		before:      "benchmark Append",              // for Append
		updatedPath: []interface{}{99, "number3", 9}, // for incr
		expect:      1,                               // for incr
	},
	{
		updatingVal: "./testdata/test_2.8M.json",
		path:        []interface{}{"first", "html", "body", "iframe", "array"},
		before:      "benchmark Append",                                       // for Append
		updatedPath: []interface{}{"first", "html", "body", "iframe", "incr"}, // for incr
		expect:      1,                                                        // for incr
	},
	{
		updatingVal: "./testdata/test_8.6M.json",
		path:        []interface{}{"first", "html", "body", "iframe", "array"},
		before:      "benchmark Append",                                       // for Append
		updatedPath: []interface{}{"first", "html", "body", "iframe", "incr"}, // for incr
		expect:      1,                                                        // for incr
	},
}
