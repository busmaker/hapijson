package main

import (
	"fmt"

	"github.com/LBJ-the-GOAT/hapijson"
)

var jsonData = []byte(`{
	"name": "LBJ", "height": 2.04, "title": 3, "mvp": 4, 
	"teams": ["LAL", "CAVS", "HEAT"], "No.": [23, 6], "first pick": true, 
	"career": [
		{
			"year": "2003-2010",
			"team": "CAVS"
		}, 
		{
			"year": "2011-2014",
			"team": "HEAT"
		},
		{
			"year": "2014-2018",
			"team": "CAVS"
		}, 
		{
			"year": "2018-present",
			"team": "Lakers"
		}, 
	]
}`)

func main() {
	getters()
	setters()
}

func setters() {
	jsonData := jsonData

	jsonData, _ = hapijson.Set(jsonData, "King LBJ", "name")
	fmt.Println(hapijson.Get(jsonData, "name")) // outputs King LBJ

	jsonData, _ = hapijson.Set(jsonData, "6 ft. 8 in.", "height")
	fmt.Println(hapijson.Get(jsonData, "name")) // outputs 6 ft. 8 in.

	jsonData, _ = hapijson.Remove(jsonData, "teams", 2) // remove HEAT
	fmt.Println(hapijson.Get(jsonData, "teams"))        // outputs []string{"LAL", "CAVS"}

	jsonData, _ = hapijson.Append(jsonData, hapijson.Path("teams"), "HEAT") // add back HEAT
	fmt.Println(hapijson.Get(jsonData, "teams"))                            // outputs []string{"LAL", "CAVS", "HEAT"}

	jsonData, _ = hapijson.Merge(jsonData, false, hapijson.Path(), "mvp", 5, "fmvp", 3)
	// preserve is false, Path points to root, replace with key "mvp" with 5, and append "fmvp": 3
	fmt.Println(hapijson.Get(jsonData, "mvp"))  // outputs 5
	fmt.Println(hapijson.Get(jsonData, "fmvp")) // outputs 3

	jsonData, _ = hapijson.Clear(jsonData, "career") // key "career" becomes []
	fmt.Println(hapijson.Get(jsonData, "career"))    // outputs []intereface{}

	jsonData, _ = hapijson.Incr(jsonData, 2, "title")
	fmt.Println(hapijson.Get(jsonData, "title")) // title is increased by 2,  outputs 5

}

func getters() {

	jsonData := jsonData
	fmt.Println(hapijson.Get(jsonData, "name"))
	// output LBJ, nil

	fmt.Println(hapijson.String(jsonData, "name"))
	// output LBJ, nil

	fmt.Println(hapijson.Int(jsonData, "title"))
	// output 3, nil

	fmt.Println(hapijson.Int64(jsonData, "mvp"))
	// output 4, nil

	fmt.Println(hapijson.Float(jsonData, "height"))
	// output 2.04, nil

	fmt.Println(hapijson.Bool(jsonData, "first pick"))
	// output true, nil

	fmt.Println(hapijson.Map(jsonData, "career", 0))
	// output map[string]intereface{}{ "year": "2003-2010", "team": "CAVS" }..., nil

	fmt.Println(hapijson.Size(jsonData, "career"))
	// output 4, nil

	sliceOf, e := hapijson.SliceOf(jsonData, "career")
	fmt.Println(string(sliceOf), e)
	// output the whole value of career

	fmt.Println(hapijson.StringArray(jsonData, "teams"))
	// output []string{"LAL", "CAVS", "HEAT"}, nil

	fmt.Println(hapijson.IntArray(jsonData, "No."))
	// output []int{23, 6}, nil

	fmt.Println(hapijson.Int64Array(jsonData, "No."))
	// output []int{23, 6}, nil

	fmt.Println(hapijson.FloatArray(jsonData, "No."))
	// output []int{23, 6}, nil

	fmt.Println(hapijson.InterfaceArray(jsonData, "career"))
	// output []map[string]interface{} { "year": "2003-2010", "team": "CAVS" } ....,nil

	fmt.Println(hapijson.MapArray(jsonData, "career"))
	// output []map[string]interface{} { "year": "2003-2010", "team": "CAVS" } ....,nil

}
