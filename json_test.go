package hapijson

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

var (
	jsonGetSetData []byte
)

type TestSet struct {
	path   []interface{}
	expect interface{}

	// the value of before updating
	before interface{}
	// the value being updated
	updatingVal interface{}
	updatedPath []interface{}

	handleErr func(e error) (fail bool)

	setID int
}

func init() {
	var e error
	if jsonGetSetData, e = ioutil.ReadFile("./testdata/test_get_set.json"); e != nil {
		panic(e)
	}
}

func TestCompletely(t *testing.T) {
	t.Run("Getters", TestGetters)
	t.Run("Setters", TestSetters)
	t.Run("Validator", TestValidate)
	t.Run("Prettifer/Minify", TestPrettifyAndMinify)
}

func TestValidate(t *testing.T) {

	var count int
	for i, j := range jsonValidTestSet {
		if e := Validate([]byte(j)); e != nil {
			count++
			fmt.Printf("Error occured at NO.%d json, %q\n", i, e.Error())
		}
	}
	if count == 0 {
		t.Log("Validation Test All success!")
	} else {
		t.Fatal(fmt.Errorf("Valid Test Failed, errors occurred times: %d ", count))
	}

	count = 0
	for i, j := range jsonInvalidTestSet {
		if e := Validate([]byte(j)); e == nil {
			count++
			t.Logf("Expected failed bu success at NO.%d json, %q\n", i, j)
		}
	}
	if count == 0 {
		t.Log("Invalidation Test success!")
	} else {
		t.Fatal(fmt.Errorf("Invalidation Test Failed, success occurred times: %d", count))
	}

}

func TestPrettifyAndMinify(t *testing.T) {
	j, e := ioutil.ReadFile("./testdata/test_20K.json")
	if e != nil {
		t.Fatal(e)
	}
	j = Minify(j)
	Prettify(j, 2)
}

func TestGetters(t *testing.T) {
	t.Run("get Root", TestRoot)
	t.Run("get String", TestString)
	t.Run("get Int", TestInt)
	t.Run("get Int64", TestInt64)
	t.Run("get Float", TestFloat)
	t.Run("get Bool", TestBool)
	t.Run("get Map", TestMap)
	t.Run("get StringArray", TestStringArray)
	t.Run("get IntArray", TestIntArray)
	t.Run("get Int64Array", TestInt64Array)
	t.Run("get FloatArray", TestFloatArray)
	t.Run("get BoolArray", TestBoolArray)
	t.Run("get MapArray", TestMapArray)
	t.Run("get InterfaceArray", TestInterfaceArray)

	t.Run("get Size", TestSize)

	t.Run("SliceOf", TestSliceOf)

}
func TestSetters(t *testing.T) {
	t.Run("set String", TestSetString)
	t.Run("set Int", TestSetInt)
	t.Run("set Int64", TestSetInt64)
	t.Run("set Float", TestSetFloat)
	t.Run("set Bool", TestSetBool)
	t.Run("set Map", TestSetMap)
	t.Run("set StringArray", TestSetStringArray)
	t.Run("set IntArray", TestSetIntArray)
	t.Run("set Int64Array", TestSetInt64Array)
	t.Run("set FloatArray", TestSetFloatArray)
	t.Run("set BoolArray", TestSetBoolArray)
	t.Run("set MapArray", TestSetMapArray)
	t.Run("set InterfaceArray", TestSetInterfaceArray)

	t.Run("Merge", TestMerge)
	t.Run("Append", TestAppend)
	t.Run("Remove", TestRemove)
	t.Run("Clear", TestClear)
	t.Run("Increase", TestIncrAndDecr)

}

func TestRoot(t *testing.T) {
	testSet := []TestSet{
		{
			path:        []interface{}{},
			expect:      []interface{}{},
			updatingVal: []byte(`[]`),
		},
		{
			path:        []interface{}{},
			expect:      map[string]interface{}{},
			updatingVal: []byte(`{}`),
		},
		{
			path:        []interface{}{},
			expect:      "string",
			updatingVal: []byte(`"string"`),
		},
		{
			path:        []interface{}{},
			expect:      123456789,
			updatingVal: []byte("123456789"),
		},
		{
			path:        []interface{}{},
			expect:      1.1,
			updatingVal: []byte(`1.1`),
		},
		{
			path:        []interface{}{},
			expect:      true,
			updatingVal: []byte(`true`),
		},
		{
			path:        []interface{}{},
			expect:      false,
			updatingVal: []byte(`false`),
		},
		{
			path:        []interface{}{},
			expect:      nil,
			updatingVal: []byte(`null`),
		},
	}
	for _, set := range testSet {
		if val, e := Get(set.updatingVal.([]byte), set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %s but got %s", set.expect, val)
		}
	}
}

func TestString(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"title"}, expect: "Game of Thrones"},
		{path: []interface{}{"ratings", 1, "TV.com"}, expect: "9/10"},
		{path: []interface{}{"ratings", 2, "ROTTEN TOMATO"}, expect: "89%"},
		{path: []interface{}{"cast", 0, "Kit Harington"}, expect: "Jon Snow"},
		{path: []interface{}{"reviews", 0, "review", 0, "vote"}, expect: "756/757"},
		{path: []interface{}{"reviews", 0, "review", 1}, expect: "season 8 😂👍"},
		{path: []interface{}{"reviews", 0, "review", 2}, expect: "season 8 😂👍"},
		{path: []interface{}{"reviews", 1, "review", 1}, expect: "the GREATEST ever! ❤"},
		{path: []interface{}{"reviews", 1, "review", 2}, expect: "the GREATEST ever! ❤"},
		{path: []interface{}{"reviews", 2, "user"}, expect: "Mary come here 👄"},
		{path: []interface{}{"reviews", 2, "review", 1}, expect: "dragon die for sb's terrible writing."},
	}
	for _, set := range testSet {
		str, e := String(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if str != set.expect.(string) {
			t.Logf("Expected %s but got %s", set.expect, str)
			t.Fail()
		}
	}

}

func TestInt(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"number of seasons"}, expect: 8},
		{path: []interface{}{"reviews", 0, "review", 0, "time"}, expect: 1591241446849},
		{path: []interface{}{"reviews", 0, "review", 0, "stars"}, expect: 8},
		{path: []interface{}{"reviews", 1, "review", 0, "time"}, expect: 1591231446849},
		{path: []interface{}{"reviews", 1, "review", 0, "stars"}, expect: 10},
		{path: []interface{}{"reviews", 2, "review", 0, "time"}, expect: 1591231846849},
		{path: []interface{}{"reviews", 2, "review", 0, "stars"}, expect: 2},
	}
	for _, set := range testSet {
		val, e := Int(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if val != set.expect.(int) {
			t.Logf("Expected %d but got %d", set.expect, val)
			t.Fail()
		}
	}

}
func TestInt64(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"id"}, expect: 1591231846849159000},
		{path: []interface{}{"relevant", "Episodes", 1, 1}, expect: 1591231846849159596},
		{path: []interface{}{"reviews", 2, "review", 0, "time"}, expect: 1591231846849},
	}
	for _, set := range testSet {
		val, e := Int64(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if val != int64(set.expect.(int)) {
			t.Logf("Expected %d but got %d", set.expect, val)
			t.Fail()
		}
	}
}
func TestFloat(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"ratings", 0, "IMDB"}, expect: 9.3},
		{path: []interface{}{"reviews", 2, "review", 2}, expect: 19.0e-21},
		{path: []interface{}{"reviews", 1, "review", 3}, expect: 3.141592653},
	}
	for _, set := range testSet {
		val, e := Float(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if val != set.expect.(float64) {
			t.Logf("Expected %100.10f but got %100.10f", set.expect, val)
			t.Fail()
		}
	}
}
func TestBool(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"liked"}, expect: true},
		{path: []interface{}{"reviews", 0, "review", 0, "spoiler"}, expect: false},
		{path: []interface{}{"reviews", 2, "review", 0, "spoiler"}, expect: true},
	}
	for _, set := range testSet {
		val, e := Bool(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if val != set.expect.(bool) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}

func TestMap(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"ratings", 0}, expect: map[string]interface{}{"IMDB": 9.3}},
		{path: []interface{}{"reviews", 1}, expect: map[string]interface{}{
			"user": "I am Tony.",
			"review": []interface{}{
				map[string]interface{}{
					"time":    1591231446849,
					"stars":   10,
					"vote":    "19090/19095",
					"spoiler": false,
				},
				"the GREATEST ever! ❤",
				"the GREATEST ever! ❤",
				3.141592653,
				2,
				-1,
				false,
				true,
				nil,
			},
		}},
	}
	for _, set := range testSet {
		val, e := Map(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestStringArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"genre"}, expect: []string{"Fantasy", " Action", "Adventure", "Drama"}},
		{path: []interface{}{"relevant", "Plot Keywords"}, expect: []string{" based on novel ", " dragon ", " politics ", " queen ", " nudity"}},
	}
	for _, set := range testSet {
		val, e := StringArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}
func TestIntArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []int{2019, 2017, 2016, 2015, 2014, 2013, 2012, 2011}},
		{path: []interface{}{"relevant", "Episodes", 0, "seasons"}, expect: []int{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	for _, set := range testSet {
		val, e := IntArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}
func TestInt64Array(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []int64{2019, 2017, 2016, 2015, 2014, 2013, 2012, 2011}},
		{path: []interface{}{"relevant", "Episodes", 1}, expect: []int64{1591231846849159000, 1591231846849159596, 1591231846849159596}},
	}
	for _, set := range testSet {
		val, e := Int64Array(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}
func TestFloatArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"float64a"}, expect: []float64{3.141592653, 4.141592653, 5.141592653, 7.141592653, 21, 0, -1, -3.141592653, -4.141592653}},
	}
	for _, set := range testSet {
		val, e := FloatArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}
func TestBoolArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"boola"}, expect: []bool{false, true, false, true, true, false}},
	}
	for _, set := range testSet {
		val, e := BoolArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestMapArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"cast"}, expect: []map[string]interface{}{
			{"Kit Harington": "Jon Snow", "another": 32423},
			{"Emilia Clarke": "Daenerys Targaryen"},
			{"Sophie Turner": "Sansa Stark"},
			{"Maisie Williams": "Arya Stark"},
		}},
		{path: []interface{}{"ratings"}, expect: []map[string]interface{}{
			{"IMDB": 9.3},
			{"TV.com": "9/10"},
			{"ROTTEN TOMATO": "89%"},
		}},
	}
	for _, set := range testSet {
		val, e := MapArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestInterfaceArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"cast"}, expect: []interface{}{
			map[string]interface{}{"Kit Harington": "Jon Snow", "another": 32423},
			map[string]interface{}{"Emilia Clarke": "Daenerys Targaryen"},
			map[string]interface{}{"Sophie Turner": "Sansa Stark"},
			map[string]interface{}{"Maisie Williams": "Arya Stark"},
		}},
		{path: []interface{}{"ratings"}, expect: []interface{}{
			map[string]interface{}{"IMDB": 9.3},
			map[string]interface{}{"TV.com": "9/10"},
			map[string]interface{}{"ROTTEN TOMATO": "89%"},
		}},

		{path: []interface{}{"special"}, expect: []interface{}{
			"the GREATEST ever! ❤",
			"the GREATEST ever! ❤",
			float64(3.141592653),
			2,
			-1,
			false,
			true,
			nil,
		}},
	}
	for _, set := range testSet {
		val, e := InterfaceArray(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestSliceOf(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"cast", 0, "Kit Harington"}, expect: `"Jon Snow"`},
	}
	for _, set := range testSet {
		val, e := SliceOf(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(set.expect, string(val)) {
			t.Logf("Expected %v but got %v", set.expect, string(val))
			t.Fail()
		}
	}
}

func TestSize(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "Plot Keywords"}, expect: 5},
		{path: []interface{}{"relevant", "Episodes", 0, "seasons"}, expect: 8},
		{path: []interface{}{"reviews", 2, "review", 0}, expect: 4},
		{path: []interface{}{"cast", 1}, expect: 1},
	}
	for _, set := range testSet {
		val, e := Size(jsonGetSetData, set.path...)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(set.expect, val) {
			t.Logf("Expected %d but got %d", set.expect, val)
			t.Fail()
		}
	}
}

func TestSetString(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"title"}, expect: "Breaking Bad"},
		{path: []interface{}{"ratings", 1, "TV.com"}, expect: "8.9/10"},
		{path: []interface{}{"ratings", 2, "ROTTEN TOMATO"}, expect: "87%"},
		// {path: []interface{}{"cast", 0, "Bryan Cranston"}, expect: "Walter White"},
		{path: []interface{}{"reviews", 0, "review", 0, "vote"}, expect: "881/890"},
		{path: []interface{}{"reviews", 0, "review", 1}, expect: "season 5 the BEST ever! 👍"},
		{path: []interface{}{"reviews", 0, "review", 2}, expect: "STAY OUT OF MY TERRITORY"},
		{path: []interface{}{"reviews", 1, "review", 1}, expect: "the GREATEST show ever! Breaking bad"},
		{path: []interface{}{"reviews", 1, "review", 2}, expect: "Pinky Man! 😆 "},
		{path: []interface{}{"reviews", 2, "user"}, expect: "Mr.White"},
		{path: []interface{}{"reviews", 2, "review", 1}, expect: "Holly christ season FIVE!!!"},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if str, e := String(data, set.path...); e != nil {
			t.Fatal(e)
		} else if str != set.expect.(string) {
			t.Logf("Expected %s but got %s", set.expect, str)
			t.Fail()
		}
	}

}

func TestSetInt(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"number of seasons"}, expect: 5},
		{path: []interface{}{"reviews", 0, "review", 0, "time"}, expect: 1591245446849},
		{path: []interface{}{"reviews", 0, "review", 0, "stars"}, expect: 9},
		{path: []interface{}{"reviews", 1, "review", 0, "time"}, expect: 1591239446849},
		{path: []interface{}{"reviews", 1, "review", 0, "stars"}, expect: 10},
		{path: []interface{}{"reviews", 2, "review", 0, "time"}, expect: 1591235846849},
		{path: []interface{}{"reviews", 2, "review", 0, "stars"}, expect: 10},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Int(data, set.path...); e != nil {
			t.Fatal(e)
		} else if val != set.expect.(int) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}
func TestSetInt64(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"id"}, expect: 1591231846849159001},
		{path: []interface{}{"relevant", "Episodes", 1, 1}, expect: 1591231846849159597},
		{path: []interface{}{"reviews", 2, "review", 0, "time"}, expect: 1591231886849},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Int64(data, set.path...); e != nil {
			t.Fatal(e)
		} else if val != int64(set.expect.(int)) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}
func TestSetFloat(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"ratings", 0, "IMDB"}, expect: 9.5},
		{path: []interface{}{"reviews", 2, "review", 2}, expect: 19.1e-21},
		{path: []interface{}{"reviews", 1, "review", 3}, expect: 3.1415926535},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Float(data, set.path...); e != nil {
			t.Fatal(e)
		} else if val != set.expect.(float64) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}

}
func TestSetBool(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"liked"}, expect: false},
		{path: []interface{}{"reviews", 0, "review", 0, "spoiler"}, expect: true},
		{path: []interface{}{"reviews", 2, "review", 0, "spoiler"}, expect: false},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Bool(data, set.path...); e != nil {
			t.Fatal(e)
		} else if val != set.expect.(bool) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}

func TestSetMap(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"ratings", 0}, expect: map[string]interface{}{"IMDB": 9.9}},
		{path: []interface{}{"reviews", 1}, expect: map[string]interface{}{
			"user": "Mr.Pinkman",
			"review": []interface{}{
				map[string]interface{}{
					"time":    1591234446849,
					"stars":   9,
					"vote":    "190900/190950",
					"spoiler": true,
					"profile": true,
				},
				"Pinkman is cool",
				"pinkman the guy ❤",
				3.1415926535897,
				200,
				-1100,
				true,
				false,
				true,
			},
		}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Map(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}

func TestSetStringArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"genre"}, expect: []string{"Action", "Adventure", "Drama", "Crime"}},
		{path: []interface{}{"relevant", "Plot Keywords"}, expect: []string{"cooking", "selling", "buying", "laudring", "lying", "running"}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := StringArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}
func TestSetIntArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []int{2013, 2012, 2011, 2010, 2009, 2008}},
		{path: []interface{}{"relevant", "Episodes", 0, "seasons"}, expect: []int{1, 2, 3, 4, 5}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := IntArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}

func TestSetInt64Array(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []int64{2013, 2012, 2011, 2010, 2009, 2008}},
		{path: []interface{}{"special", 1}, expect: []int64{1591231846849159000, 1591231846849159596, 1591231846849159596}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := Int64Array(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}

}
func TestSetFloatArray(t *testing.T) {
	var testSet = []TestSet{
		// {path: []interface{}{"relevant", "years"}, expect: []int64{2013, 2012, 2011, 2010, 2009, 2008}},
		// {path: []interface{}{"special", 1}, expect: []int64{1591231846849159000, 1591231846849159596, 1591231846849159596}},
		{path: []interface{}{"relevant", "years"}, expect: []float64{3.141592653, 4.141592653, 5.141592653, 7.141592653, 21, 0, -1, -3.141592653, -4.141592653}},
		{path: []interface{}{"special", 1}, expect: []float64{3.141592653, 4.141592653, 5.141592653, 7.141592653, 21, 0, -1, -3.141592653, -4.141592653}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := FloatArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}

}
func TestSetBoolArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []bool{false, true, false, true, true, false}},
		{path: []interface{}{"special", 1}, expect: []bool{false, true, false, true, true, false}},
		// {path: []interface{}{"relevant", "years"}, expect: []int64{2013, 2012, 2011, 2010, 2009, 2008}},
		// {path: []interface{}{"special", 1}, expect: []int64{1591231846849159000, 1591231846849159596, 1591231846849159596}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := BoolArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}

}
func TestSetMapArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"relevant", "years"}, expect: []map[string]interface{}{
			{"Kit Harington": "Jon Snow"},
			{"Emilia Clarke": "Daenerys Targaryen"},
			{"Sophie Turner": "Sansa Stark"},
			{"Maisie Williams": "Arya Stark"},
		}},
		{path: []interface{}{"special", 1}, expect: []map[string]interface{}{
			{"Kit Harington": "Jon Snow"},
			{"Emilia Clarke": "Daenerys Targaryen"},
			{"Sophie Turner": "Sansa Stark"},
			{"Maisie Williams": "Arya Stark"},
		}},
		// {path: []interface{}{"relevant", "years"}, expect: []int64{2013, 2012, 2011, 2010, 2009, 2008}},
		// {path: []interface{}{"special", 1}, expect: []int64{1591231846849159000, 1591231846849159596, 1591231846849159596}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := MapArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}

}
func TestSetInterfaceArray(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"cast"}, expect: []interface{}{
			map[string]interface{}{"Bryan Cranston": "Walter White"},
			map[string]interface{}{"Ann Gunn": "Skyler White"},
			map[string]interface{}{"Aaron Pual": "Jesse Pinkman"},
			map[string]interface{}{"Betsy Brandt": "Marie Schrader"},
			map[string]interface{}{"RJ Mitte": "Walter Whiter, Jr."},
			map[string]interface{}{"Dean Norris": "Hank   Schrader"},
		}},
		{path: []interface{}{"ratings"}, expect: []interface{}{
			map[string]interface{}{"IMDB": 9.5},
			map[string]interface{}{"ROTTEN TOMATO": "90%"},
		}},

		{path: []interface{}{"special"}, expect: []interface{}{1234}},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Set(data, set.expect, set.path...); e != nil {
			t.Fatal(e)
		} else if val, e := InterfaceArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %v but got %v", set.expect, val)
			t.Fail()
		}
	}
}

func TestAppend(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"reviews", 2, "review"},
			before: []interface{}{
				map[string]interface{}{
					"time":    1591231846849,
					"stars":   2,
					"vote":    "100/100",
					"spoiler": true,
				},
				"dragon die for sb's terrible writing.",
				19.0e-21,
				19.0e2,
			},
			updatingVal: []interface{}{
				"from test append",
				"a random value",
				3.141592653879,
				123456789,
				[]interface{}{1, 23, 345, 567},
			},
			expect: []interface{}{
				map[string]interface{}{
					"time":    1591231846849,
					"stars":   2,
					"vote":    "100/100",
					"spoiler": true,
				},
				"dragon die for sb's terrible writing.",
				19.0e-21,
				19.0e2,
				"from test append",
				"a random value",
				3.141592653879,
				123456789,
				[]interface{}{1, 23, 345, 567},
			},
		},
	}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		if data, e := Append(data, set.path, set.updatingVal.([]interface{})...); e != nil {
			t.Fatal(e)
		} else if val, e := InterfaceArray(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestRemove(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"cast", 1}, updatedPath: []interface{}{"cast"}, expect: 3, setID: 0},
		{path: []interface{}{"cast", 1}, updatedPath: []interface{}{"cast"}, expect: 2, setID: 0},
		{path: []interface{}{"cast", 1}, updatedPath: []interface{}{"cast"}, expect: 1, setID: 0},
		{path: []interface{}{"cast", 0}, updatedPath: []interface{}{"cast"}, expect: 0, setID: 0},
		{path: []interface{}{"reviews", 1, "review", 3}, updatedPath: []interface{}{"reviews", 1, "review", 3},
			expect: 2, setID: 1},
		{path: []interface{}{"reviews", 2, "review", 0, "spoiler"},
			updatedPath: []interface{}{"reviews", 2, "review", 0, "spoiler"}, expect: nil, setID: 2,
			handleErr: func(e error) (fail bool) {
				// spoiler key was removed
				return strings.Index(e.Error(), "not found") == -1
			}},
		{path: []interface{}{}, updatedPath: []interface{}{}, expect: map[string]interface{}{}, setID: 3},
	}

	data := append([]byte{}, jsonGetSetData...)
	var val interface{}
	var e error
	for _, set := range testSet {
		if data, e = Remove(data, set.path...); e != nil {
			if set.handleErr != nil {
				if set.handleErr(e) {
					t.Fatal(e)
				}
			} else {
				t.Fatal(e)
			}
		}
		if set.setID == 0 {
			if val, e = Size(data, set.updatedPath...); e != nil {
				if set.handleErr != nil {
					if set.handleErr(e) {
						t.Fatal(e)
					}
				} else {
					t.Fatal(e)
				}
			}
		} else if set.setID == 1 || set.setID == 2 {
			if val, e = Get(data, set.updatedPath...); e != nil {
				if set.handleErr != nil {
					if set.handleErr(e) {
						t.Fatal(e)
					}
				} else {
					t.Fatal(e)
				}
			}
		} else if set.setID == 3 {
			if data, e = Remove(data, set.updatedPath...); e != nil {
				t.Fatal(e)
			}
			if val, e = Get(data, set.updatedPath...); e != nil {
				t.Fatal(e)
			}
		} else {
			t.Fatal(set.setID)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}

	}

}

func TestClear(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"number of seasons"},
			before: 8,
			expect: 0,
		},
		{path: []interface{}{"liked"},
			before: true,
			expect: false,
		},
		{path: []interface{}{"title"},
			before: "Game of Thrones",
			expect: "",
		},
		{path: []interface{}{"ratings", 0, "IMDB"},
			before: 9.3,
			expect: 0,
		},
		{path: []interface{}{"ratings", 1},
			before: map[string]interface{}{"TV.com": "9/10"},
			expect: map[string]interface{}{},
		},
		{path: []interface{}{"ratings"},
			expect: []interface{}{},
		},
		{path: []interface{}{},
			expect: map[string]interface{}{},
		},
	}
	data := append([]byte{}, jsonGetSetData...)
	var val interface{}
	var e error
	for _, set := range testSet {
		if data, e = Clear(data, set.path...); e != nil {
			t.Fatal(e)
		}
		if val, e = Get(data, set.path...); e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestMerge(t *testing.T) {
	var testSet = []TestSet{
		{path: []interface{}{"ratings", 0},
			before:      map[string]interface{}{"IMDB": 9.3},
			updatingVal: map[string]interface{}{"Toxic": 9.8, "MataCritic": 9.9},
			expect:      map[string]interface{}{"IMDB": 9.3, "Toxic": 9.8, "MataCritic": 9.9},
		},
		{path: []interface{}{"cast", 0},
			before:      map[string]interface{}{"Kit Harington": "Jon Snow"},
			updatingVal: map[string]interface{}{"Kit Harington": "Jon Snow 123"},
			expect:      map[string]interface{}{"Kit Harington": []interface{}{"Jon Snow", "Jon Snow 123"}, "another": 32423},
			setID:       1,
		},
		{path: []interface{}{"the best ever"},
			before: map[string]interface{}{
				"merge": 234,
				"ary":   []interface{}{1, 2, 3, 4},
				"why?": map[string]interface{}{
					"reason1": "brilliant!", "reason2": "wonderful",
				},
			},
			updatingVal: map[string]interface{}{
				"merge": "Game of Zones",
				"ary":   5,
				"why?": map[string]interface{}{
					"reason1": "greatest!", "reason3": "good good good",
				},
			},
			expect: map[string]interface{}{
				"merge": []interface{}{234, "Game of Zones"},
				"ary":   []interface{}{1, 2, 3, 4, 5},
				"why?": []interface{}{
					map[string]interface{}{"reason1": "brilliant!", "reason2": "wonderful"},
					map[string]interface{}{"reason1": "greatest!", "reason3": "good good good"},
				},
			},
		},
		{
			path: []interface{}{"test merge not map"},
			before: map[string]interface{}{
				"merge": 234,
				"ary":   []interface{}{1, 2, 3},
				"why?": map[string]interface{}{
					"reason1": "brilliant!", "reason2": "wonderful",
				},
			},
			updatingVal: []interface{}{
				"merge", "Game of Zones",
				"ary", []interface{}{4, 5, 6},
				"why?", map[string]interface{}{
					"reason1": "goodness!", "reason3": "GOAT",
				},
			},
			expect: map[string]interface{}{
				"merge": []interface{}{234, "Game of Zones"},
				"ary":   []interface{}{1, 2, 3, 4, 5, 6},
				"why?": []interface{}{
					map[string]interface{}{"reason1": "brilliant!", "reason2": "wonderful"},
					map[string]interface{}{"reason1": "goodness!", "reason3": "GOAT"},
				},
			},
			setID: 3,
		},
		{
			path: []interface{}{"test merge no preserve"},
			before: map[string]interface{}{
				"merge": 234,
				"so\"escaped \\<quote>" + string([]byte{0xe2, 0x80, 0xa8}) + string([]byte{0xe2, 0x80, 0xa9}): "escaped",
				"why?": map[string]interface{}{
					"reason1": "brilliant!", "reason2": "wonderful",
				},
			},
			updatingVal: []interface{}{
				"merge", "Game of Zones",
				"so\"esca\tp\ned \\<quote>" + string([]byte{0xe2, 0x80, 0xa8}) + string([]byte{0xe2, 0x80, 0xa9}), `y\"<eah>\n\t\r"\\\` + string([]byte{0xe2, 0x80, 0xa8}) + string([]byte{0xe2, 0x80, 0xa9}),
				"why?", map[string]interface{}{
					"reason1": "goodness!", "reason3": "GOAT",
				},
			},
			expect: map[string]interface{}{
				"merge": "Game of Zones",
				"so\"esca\tp\ned \\<quote>" + string([]byte{0xe2, 0x80, 0xa8}) + string([]byte{0xe2, 0x80, 0xa9}): "y\"<eah>\n\t\r\"\\\\" + string([]byte{0xe2, 0x80, 0xa8}) + string([]byte{0xe2, 0x80, 0xa9}),
				"why?": map[string]interface{}{
					"reason1": "goodness!", "reason3": "GOAT",
				},
			},
			setID: 4, // no preserve.
		},
	}
	data := append([]byte{}, jsonGetSetData...)
	var e error
	var val interface{}
	for _, set := range testSet {
		data := data

		if set.setID == 4 {
			if data, e = Merge(data, false, set.path, set.updatingVal.([]interface{})...); e != nil {
				t.Fatal(e)
			}
		} else if set.setID == 3 {
			if data, e = Merge(data, true, set.path, set.updatingVal.([]interface{})...); e != nil {
				t.Fatal(e)
			}
		} else {
			if data, e = Merge(data, true, set.path, set.updatingVal.(map[string]interface{})); e != nil {
				t.Fatal(e)
			}
		}
		if val, e = Map(data, set.path...); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(val, set.expect) {
			t.Logf("Expected %#v but got %#v", set.expect, val)
			t.Fail()
		}
	}
}

func TestIncrAndDecr(t *testing.T) {
	testSet := []TestSet{
		// int 1
		{
			path:        []interface{}{"incr", 0},
			before:      1,
			updatingVal: 100,
			expect:      101,
		},
		{
			path:        []interface{}{"incr", 0},
			before:      101,
			updatingVal: int64(100),
			expect:      201,
		},
		{
			path:        []interface{}{"incr", 0},
			before:      201,
			updatingVal: float32(107.1),
			expect:      308.1,
		},
		/* 		{
			path:        []interface{}{"incr", 0},
			before:      308.1,
			updatingVal: 100.1,
			expect:      408.2,
		}, */
		// int
		{
			path:        []interface{}{"incr", 1},
			before:      2147483647,
			updatingVal: 100,
			expect:      2147483747,
		},
		{
			path:        []interface{}{"incr", 1},
			before:      2147483747,
			updatingVal: int64(100),
			expect:      2147483847,
		},
		{
			path:        []interface{}{"incr", 1},
			before:      2147483847,
			updatingVal: 100.1,
			expect:      2147483947.1,
		},

		// int64
		{
			path:        []interface{}{"incr", 2},
			before:      9223372036854775807,
			updatingVal: int64(-100),
			expect:      9223372036854775707,
		},
		// float32
		{
			path:        []interface{}{"incr", 3},
			before:      9.3,
			updatingVal: 0.2,
			expect:      9.5,
		},
		{
			path:        []interface{}{"incr", 3},
			before:      9.5,
			updatingVal: 2,
			expect:      11.5,
		},
		{
			path:        []interface{}{"incr", 3},
			before:      11.5,
			updatingVal: int64(3),
			expect:      14.5,
		},
		{
			path:        []interface{}{"incr", 3},
			before:      14.5,
			updatingVal: float32(3.2),
			expect:      17.7,
		},
		// float64
		{
			path:        []interface{}{"incr", 4},
			before:      123456.123456,
			updatingVal: 1005.10001,
			expect:      124461.223466,
		},
	}
	var e error
	var val interface{}
	data := append([]byte{}, jsonGetSetData...)
	for _, set := range testSet {
		data, e = Incr(data, set.updatingVal, set.path...)
		if e != nil {
			t.Fatal(e)
		} else if val, e = Get(data, set.path); e != nil {
			t.Fatal(e)
		} else if !reflect.DeepEqual(set.expect, val) {
			t.Logf("Expected %v but got %v, %T, %T", set.expect, val, set.expect, val)
			t.Fail()
		}
	}

}
