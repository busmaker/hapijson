package hapijson

//TODO: bug: handle duplicated keys situation.
// 		e.g. {"key": 1, "key": 2} the later one should overwrite the former one,
//		Solution 1: when retrieve the value of a key, we should start searching from the back.
//		Solution 2: when search a key, we search to the end of the object, and use the last key to return.

//TODO: Clearify error info.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

type valType int8

const (
	valUnknown valType = iota
	valString
	valArray
	valObject
	valNumber
	valTrue
	valFalse
	valNull

	valFloat // for decimal numbers
)

var (
	ErrInvalidJSONPayload = errors.New("Invalid JSON payload")
)

// Path returns pathNodes in []interface{} for calling Merge/Append handly,
// pathNodes left empty means get to the root element of json
func Path(pathNodes ...interface{}) []interface{} {
	return pathNodes
}

// Set sets a value to the last node of the pathNodes which may be a key or an index.
// val can be any go standard type or the slices or arrays of them except chan, e.g. string, int, bool, float,
// interface{}, []string, []int, []bool, []map[string]interface{}... etc, for map only supports map[string] interface{}.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json payload, it doesn't do checking inside,
// if caller doesn't sure the validation of the json data, could call Validate()
// to ensure it.
//
// It would not allocate a new memory if the data has suffient place to contain the new value,
// this means it might do modification right at the data, if caller wants to preserve the original data after
// set, then make a copy beforehand is needed.
//
func Set(data []byte, val interface{}, pathNodes ...interface{}) (newData []byte, e error) {
	var start, end int
	if start, end, _, _, e = path(data, 0, pathNodes...); e != nil {
		return
	}
	var valJSON []byte
	if valJSON, _, e = toJSON(val); e != nil {
		return
	}
	newData, _, _ = updatePayload(data, valJSON, start, end, rootEndOf(data))
	return
}

// Merge merges objects, e.g. json:
//	{"key": "val", "key2":["val2"], "key3": {"k3": 3}, "key4": "Stay the same"}
// being merged vals:
//	{"key": 1, "key2": 2, "key3": 3}
// after merged, if preserve is true
// 	{"key": ["val", 1], "key2": ["val2", 2], "key3":[{"key3": 3}, 3],  "key4": "Stay the same"}
// if preserve is false
// 	{"key": 1, "key2": 2, "key3": 3, "key4": "Stay the same"}
//
// if preserve is true, it merges the values of new keys with the existing keys' into arrays, otherwise it overwirtes
// the existing keys. e.g. "key4" in above has not matched with any of the new keys so it stays what it is.
//
// vals must be either map[string]interface{}, e.g.
//	Merge(data, preserve, pathNodes, map[string]interface{}{"key":"val", ...})
// or pairs of arguments as key sets, e.g.
//	Merge(data, preserve, pathNodes, "key", "val", "key2", "val2", ....)
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json encoding, it doesn't do checking inside ...
// See the Note part of Set().
func Merge(data []byte, preserve bool, pathNodes []interface{}, vals ...interface{}) (newData []byte, e error) {
	var start, end int
	var vtype valType
	if start, end, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	} else if vtype != valObject {
		return nil, genNotTypeError("not json object", pathNodes)
	}
	var ok bool
	var m map[string]interface{}
	if ln := len(vals); ln == 1 {
		if m, ok = vals[0].(map[string]interface{}); !ok {
			e = fmt.Errorf("%v is not a valid argument", vals[0])
			return
		}
	} else if ln%2 != 0 {
		e = fmt.Errorf("%q missing its value", vals[len(vals)-1])
		return
	} else {
		m = map[string]interface{}{}
		for i := 0; i < ln; i += 2 {
			if key, ok := vals[i].(string); !ok {
				e = fmt.Errorf("%v must be string as a key name", vals[i])
				return
			} else {
				m[key] = vals[i+1]
			}
		}
	}
	newData, _, _, e = merge(data, start, end, rootEndOf(data), preserve, m)
	return
}

// Append appends vals to the last node of the pathNodes which must be an array.
// vals must be the go bulit-in types or the slices or arrays of them,
// map only accepts in map[string]interface{} this kind,
// e.g. string, int, bool, ..., map[string]interface{}, []string, []int, []bool, []float, or []interface{}
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside...
// See the Note part of Set().
func Append(data []byte, pathNodes []interface{}, vals ...interface{}) (newData []byte, e error) {
	if len(vals) == 0 {
		return data, nil
	}
	var start, end int
	var vtype valType
	if start, end, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	} else if vtype != valArray {
		return nil, genNotTypeError("not json array", pathNodes)
	}

	newData, _, _, e = appendElements(data, start, end, rootEndOf(data), vtype, vals...)
	return
}

// Remove removes a key set or an elements from the last node of
// the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside...
// See the Note part of Set().
func Remove(data []byte, pathNodes ...interface{}) (newData []byte, e error) {
	if len(pathNodes) == 0 {
		// removeing the root element
		return Clear(data, pathNodes...)
	}
	var start, end, veryStart int
	if start, end, veryStart, _, e = path(data, 0, pathNodes...); e != nil {
		return
	}
	newData, _, _ = remove(data, start, end, veryStart, rootEndOf(data))
	return
}

func remove(payload []byte, start, end, veryStart, rootEnd int) (newPayload []byte, newEnd, newRootEnd int) {
	// if the key set or element is in the middle of values,
	// we need to remove its seperator the comma ',' as well,
	// e.g. [..., "want removed", ...], we need to remove either its left one or the right one comma.
	// luckly, veryStart points the very start of a key or an element, which may be one of the ',' '[' or '{',
	// so we don't need loop forward, when veryStart is pointing to ','.

	if payload[veryStart] == ',' {
		start = veryStart
	} else { //
		// start++ // keep the [ or {
		if payload[end] != ',' {
		forwarding: // forwarding to see if there is a ','
			for ; end < len(payload); end++ {
				switch payload[end] {
				case ',':
					end++
					break forwarding
				case ']', '}':
					break forwarding // no comma needs to be deleted.
				}
			}
		} else {
			end++ // delete the following ,
		}
	}
	return updatePayload(payload, nil, start, end, rootEnd)
}

// Clear removes all keys, indexes or reset values of the last node of the pathNodes.
//	Object -> {}
//	Array 	-> []
//	String 	-> ""
//	Number 	-> 0
//	Boolean -> false
// 	null    -> null
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside...
// See the Note part of Set().
func Clear(data []byte, pathNodes ...interface{}) (newData []byte, e error) {
	var start, end int
	var vtype valType
	if start, end, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	}
	rootEnd := rootEndOf(data)
	switch vtype {
	case valObject, valArray, valString:
		newData, _, _ = updatePayload(data, nil, start+1, end-1, rootEnd)
	case valNumber, valFloat:
		data[start] = '0'
		newData, _, _ = updatePayload(data, nil, start+1, end, rootEnd)
	case valTrue:
		newData, _, _ = updatePayload(data, []byte{'f', 'a', 'l', 's', 'e'}, start, end, rootEnd)
	}
	return
}

// Incr increases the number of the last node of pathNodes by delta,
// delta must be type of int, int64, float32 or float64.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json encoding, it doesn't do checking inside...
// See the Note part of Set().
func Incr(data []byte, delta interface{}, pathNodes ...interface{}) (newData []byte, e error) {
	return inOrDecrease(data, delta, pathNodes...)
}

func inOrDecrease(data []byte, delta interface{}, pathNodes ...interface{}) (newData []byte, e error) {

	var start, end int
	var vtype valType
	if start, end, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	}
	if vtype != valNumber && vtype != valFloat {
		return nil, genNotTypeError("not json number", pathNodes)
	}
	iNum, e := fromJSON(data, start, end, vtype)
	if e != nil {
		return
	}
	var result []byte
	switch n := iNum.(type) {
	case int:
		switch d := delta.(type) {
		case int:
			result = []byte(strconv.Itoa(n + d))
		case int64:
			result = []byte(strconv.FormatInt(int64(n)+d, 10))
		case float32:
			result = []byte(strconv.FormatFloat(float64(n)+float64(d), 'f', -1, 32))
		case float64:
			result = []byte(strconv.FormatFloat(float64(n)+d, 'f', -1, 64))
		default:
			e = fmt.Errorf("increase type of %T is unsupported by now", n)
			return
		}
	case int64:
		switch d := delta.(type) {
		case int:
			result = []byte(strconv.FormatInt(n+int64(d), 10))
		case int64:
			result = []byte(strconv.FormatInt(n+d, 10))
		case float32:
			result = []byte(strconv.FormatFloat(float64(float32(n)+d), 'f', -1, 32))
		case float64:
			result = []byte(strconv.FormatFloat(float64(n)+d, 'f', -1, 64))
		default:
			e = fmt.Errorf("increase type of %T is unsupported by now", n)
			return
		}
	case float32:
		switch d := delta.(type) {
		case int:
			result = []byte(strconv.FormatFloat(float64(n+float32(d)), 'f', -1, 32))
		case int64:
			result = []byte(strconv.FormatFloat(float64(n)+float64(d), 'f', -1, 32))
		case float32:
			result = []byte(strconv.FormatFloat(float64(n+d), 'f', -1, 32))
		case float64:
			result = []byte(strconv.FormatFloat(float64(n)+d, 'f', -1, 64))
		default:
			e = fmt.Errorf("increase type of %T is unsupported by now", n)
			return
		}
	case float64:
		switch d := delta.(type) {
		case int:
			result = []byte(strconv.FormatFloat(n+float64(d), 'f', -1, 64))
		case int64:
			result = []byte(strconv.FormatFloat(n+float64(d), 'f', -1, 64))
		case float32:
			result = []byte(strconv.FormatFloat(n+float64(d), 'f', -1, 32))
		case float64:
			result = []byte(strconv.FormatFloat(n+d, 'f', -1, 64))
		default:
			e = fmt.Errorf("increase type of %T is unsupported by now", n)
			return
		}
	default:
		e = fmt.Errorf("increase type of %T is unsupported by now", n)
		return
	}
	newData, _, _ = updatePayload(data, result, start, end, rootEndOf(data))
	return
}

// Get gets val from the last node of the pathNodes, it may be a key or an index.
// PS: use specific functions like String,Int,Bool or StringArray , etc.., can
// get a specific type of value instead of interface{}.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Get(data []byte, pathNodes ...interface{}) (val interface{}, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e != nil {
		return
	}
	return fromJSON(data, start, end, vtype)
}

func genNotTypeError(tip string, pathNodes []interface{}) (e error) {
	if len(pathNodes) > 0 {
		e = fmt.Errorf("path node: %v is %s", pathNodes[len(pathNodes)-1], tip)
	} else {
		e = fmt.Errorf("root element is %s", tip)
	}
	return
}

// String gets val in string from the last node of the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func String(data []byte, pathNodes ...interface{}) (val string, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valString {
			val, _, e = unescapeString(data, start+1)
			return
		}
		e = genNotTypeError("not json string", pathNodes)
	}
	return
}

// SliceOf gets the slice of the last node of the pathNodes
// ignoring the value of json type. E.g.
//	data: {"number": 32, "obj":{"object":1}} -> SliceOf(data, "number") returns 32
// So it could even does like this:
//	SliceOf(data, "obj") -> returns {"object":1}
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func SliceOf(data []byte, pathNodes ...interface{}) (val []byte, e error) {
	start, end, _, _, e := path(data, 0, pathNodes...)
	if e == nil {
		return data[start:end], nil
	}
	return
}

// StringArray gets val in string array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func StringArray(data []byte, pathNodes ...interface{}) (val []string, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return stringArray(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// Int gets val in int from the last node of the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Int(data []byte, pathNodes ...interface{}) (val int, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valNumber || vtype == valFloat {
			return strconv.Atoi(string(data[start:end]))
		}
		e = genNotTypeError("not json number", pathNodes)
	}
	return
}

// IntArray gets val in int array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func IntArray(data []byte, pathNodes ...interface{}) (val []int, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return intArray(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// Int64 gets val in int64 from the last node of the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Int64(data []byte, pathNodes ...interface{}) (val int64, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valNumber || vtype == valFloat {
			return strconv.ParseInt(string(data[start:end]), 10, 64)
		}
		e = genNotTypeError("not json number", pathNodes)
	}
	return
}

// Int64Array gets val in int64 array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Int64Array(data []byte, pathNodes ...interface{}) (val []int64, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return int64Array(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// Float gets val in float64 from the last node of the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Float(data []byte, pathNodes ...interface{}) (val float64, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valFloat || vtype == valNumber {
			return strconv.ParseFloat(string(data[start:end]), 64)
		}
		e = genNotTypeError("not json number", pathNodes)
	}
	return
}

// FloatArray gets val in float64 array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func FloatArray(data []byte, pathNodes ...interface{}) (val []float64, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return floatArray(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// Bool gets val in bool from the last node of the pathNodes which may be a key or an index.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Bool(data []byte, pathNodes ...interface{}) (val bool, e error) {
	var vtype valType
	if _, _, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	} else if vtype == valTrue {
		val = true
	} else if vtype == valFalse {
		val = false
	} else {
		e = genNotTypeError("not json boolean", pathNodes)
	}
	return
}

// BoolArray gets val in bool array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func BoolArray(data []byte, pathNodes ...interface{}) (val []bool, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return boolArray(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// Map gets val in map[string]interface{} from the last node of the pathNodes which must be an object.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func Map(data []byte, pathNodes ...interface{}) (val map[string]interface{}, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valObject {
			return toMap(data, start, end)
		}
		e = genNotTypeError("not json object", pathNodes)
	}
	return
}

// MapArray gets val in map[string]interface{} array from the last node of the pathNodes which must be an array of object.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func MapArray(data []byte, pathNodes ...interface{}) (val []map[string]interface{}, e error) {
	start, _, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return mapArray(data, start)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

// InterfaceArray gets val in interface array from the last node of the pathNodes which must be an array.
//
// pathNodes left empty means get to the root element of json
//
// Note: this function assuming data is a valid json data, it doesn't do checking inside.
func InterfaceArray(data []byte, pathNodes ...interface{}) (val []interface{}, e error) {
	start, end, _, vtype, e := path(data, 0, pathNodes...)
	if e == nil {
		if vtype == valArray {
			return interfaceArray(data, start, end)
		}
		e = genNotTypeError("not json array", pathNodes)
	}
	return
}

func stringArray(payload []byte, start int) (val []string, e error) {
	var i, end int
	val = []string{}
	var vType valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vType, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vType != valString {
			return nil, fmt.Errorf("No.%d item %s is not json string", i, string(payload[start:end]))
		} else if temp, _, err := unescapeString(payload, start+1); err != nil {
			return nil, err
		} else if val = append(val, temp); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func intArray(payload []byte, start int) (val []int, e error) {
	var i, num, end int
	val = []int{}
	var vType valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vType, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vType != valNumber {
			return nil, fmt.Errorf("No.%d item: %s is not json number", i, string(payload[start:end]))
		} else if num, e = strconv.Atoi(string(payload[start:end])); e != nil {
			return
		} else if val = append(val, num); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func int64Array(payload []byte, start int) (val []int64, e error) {
	var i, end int
	var num int64
	val = []int64{}
	var vType valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vType, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vType != valNumber {
			e = fmt.Errorf("No.%d item: %s is not a number type", i, string(payload[start:end]))
			return
		} else if num, e = strconv.ParseInt(string(payload[start:end]), 10, 64); e != nil {
			return
		} else if val = append(val, num); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func floatArray(payload []byte, start int) (val []float64, e error) {
	var i, end int
	var num float64
	val = []float64{}
	var vType valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vType, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vType != valNumber && vType != valFloat {
			return nil, fmt.Errorf("No.%d item: %s is not a number type", i, string(payload[start:end]))
		} else if num, e = strconv.ParseFloat(string(payload[start:end]), 64); e != nil {
			return
		} else if val = append(val, num); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func boolArray(payload []byte, start int) (val []bool, e error) {
	var i, end int
	var v bool
	val = []bool{}
	var vtype valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vtype, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vtype == valTrue {
			v = true
		} else if vtype == valFalse {
			v = false
		} else {
			return nil, fmt.Errorf("No.%d item: %s is not a boolean type", i, string(payload[start:end]))
		}

		if val = append(val, v); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func mapArray(payload []byte, start int) (val []map[string]interface{}, e error) {
	var i, end int
	var m map[string]interface{}
	val = []map[string]interface{}{}
	var vtype valType
	var next, emtpy bool
	for pos := start + 1; pos < len(payload); i++ {
		if pos, start, end, vtype, next, emtpy, e = nextValue(payload, pos); emtpy || e != nil {
			return
		} else if vtype != valObject {
			return nil, fmt.Errorf("No.%d item: %q is not a Object type", i, string(payload[start:end]))
		} else if m, e = toMap(payload, start, end); e != nil {
			return
		}

		if val = append(val, m); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

func interfaceArray(payload []byte, start, end int) (val []interface{}, e error) {
	val = []interface{}{}
	var ele interface{}
	var vStart, vEnd int
	var vType valType
	var next, empty bool
	// p + 1 skip the [
	for pos := start + 1; pos < end; {
		if pos, vStart, vEnd, vType, next, empty, e = nextValue(payload, pos); empty || e != nil {
			return
		}
		if ele, e = fromJSON(payload, vStart, vEnd, vType); e != nil {
			return
		}

		if val = append(val, ele); next {
			pos++
		} else {
			return
		}
	}
	return nil, ErrInvalidJSONPayload
}

// Size returns the length of an array or how many keys does an object has,
// if payload is neither array nor object then fail.
//
// pathNodes left empty means get to the root element of json
func Size(data []byte, pathNodes ...interface{}) (s int, e error) {
	var start, end int
	var vtype valType
	if start, end, _, vtype, e = path(data, 0, pathNodes...); e != nil {
		return
	} else if vtype != valArray && vtype != valObject {
		return 0, fmt.Errorf("Value is neither Array nor Object, but %d", vtype) // ErrNotArray
	}
	return size(data, vtype, start, end)
}

func size(payload []byte, vtype valType, start, end int) (size int, e error) {
	pos := start + 1 // skip the { or [
	if vtype == valObject {
		for ; pos < end; pos++ {
			var hasKey bool
			if pos, _, hasKey, e = nextKey(payload, pos, false); !hasKey || e != nil {
				return
			}
			size++
			if newPos, _, _, _, next, _, e := nextValue(payload, pos); e != nil {
				return 0, e
			} else if next {
				pos = newPos
			} else {
				return size, nil
			}
		}
		return 0, ErrInvalidJSONPayload
	}

	// array
	for ; pos < end; pos++ {
		newPos, _, _, _, next, end, _ := nextValue(payload, pos)
		if end {
			return
		}
		size++
		if next {
			pos = newPos
		} else {
			return
		}
	}
	return 0, ErrInvalidJSONPayload
}

// veryStart is the start of a key or an element, it points to the '{', '[' or ',' prior to the key or element.
func path(payload []byte, startPos int, pathNodes ...interface{}) (start, end, veryStart int, vtype valType, e error) {
	if ln := len(pathNodes); ln == 0 {
		var ok bool
		if start, end, vtype, ok = root(payload); !ok { // to the root of payload
			e = ErrInvalidJSONPayload
		}
		return
	} else if ln == 1 {
		if pns, ok := pathNodes[0].([]interface{}); ok {
			pathNodes = pns
		}
	}

readArgs:
	for argI, what := range pathNodes {
		startPos, _ = skipWhites(payload, startPos)
		veryStart = startPos
		if key, ok := what.(string); ok { // key
			if payload[startPos] != '{' {
				e = fmt.Errorf("the value of %q is not a json object", key)
				return
			}
			var tempKey string
			var next, hasKey bool
			for startPos++; startPos < len(payload); startPos++ { // iterate keys for the key.
				if startPos, tempKey, hasKey, e = nextKey(payload, startPos, true); e != nil {
					return
				} else if !hasKey {
					e = fmt.Errorf(`Error at No.%d in arguments: key %q is not found`, argI+1, key)
					return
				} else if startPos, start, end, vtype, next, _, e = nextValue(payload, startPos); e != nil {
					return
				} else if key == tempKey { // ok, done.
					startPos = start // the pos now is that where the value of this key start at.
					continue readArgs
				} else if !next {
					e = fmt.Errorf(`Error at No.%d in arguments: key %q is not found`, argI+1, key)
					return
				} else {
					veryStart = startPos
				}
			}
			e = ErrInvalidJSONPayload
			return

		} else if index, ok := what.(int); ok { // index
			if payload[startPos] != '[' {
				e = fmt.Errorf("the value of %v is not a json array", what)
				return
			}
			var aryLength int
			var empty, next bool
			for startPos++; startPos < len(payload); startPos++ {
				if startPos, start, end, vtype, next, empty, e = nextValue(payload, startPos); e != nil {
					return
				} else if empty {
					e = fmt.Errorf(`Error at No.%d in arguments: index %d out of range, the len is %d`,
						argI+1, index, aryLength)
					return
				}

				if aryLength == index { // ok, get it
					startPos = start
					continue readArgs
				}
				aryLength++
				if !next {
					e = fmt.Errorf(`Error at No.%d in arguments: index %d out of range, the len is %d`,
						argI+1, index, aryLength)
					return
				} else {
					veryStart = startPos
				}

			}
			e = ErrInvalidJSONPayload
			return
		} else {
			e = fmt.Errorf("Unsupported type %T, %v", what, what)
			return
		}
	}
	return
}

// to the root element of json payload.
func root(payload []byte) (rootStart, rootEnd int, vtype valType, ok bool) {
	if rootStart, ok = skipWhites(payload, 0); !ok {
		return
	}
	if b := payload[rootStart]; b == '{' {
		vtype = valObject
	} else if b == '[' {
		vtype = valArray
	} else if b == '"' {
		vtype = valString
	} else if b == 'f' {
		vtype = valFalse
	} else if b == 't' {
		vtype = valTrue
	} else if b == 'n' {
		vtype = valNull
	} else if b >= '0' && b <= '9' || b == '-' || b == '.' {
		if bytes.Index(payload, []byte(".")) == -1 {
			vtype = valNumber
		} else {
			vtype = valFloat
		}
	} else {
		return
	}
	if rootEnd, ok = skipWhitesLast(payload); ok {
		rootEnd++ // plus 1 so the whole payload can be covered by start and end.
	}
	return
}

func rootEndOf(payload []byte) (rootEnd int) {
	rootEnd, ok := skipWhitesLast(payload)
	if !ok {
		rootEnd = len(payload)
	} else {
		rootEnd++
	}
	return
}

// encodeToUTF8 is true encodes unicode code points \uxxxx to utf8.
func nextKey(payload []byte, curIndex int, encodeToUTF8 bool) (newPos int, key string, hasKey bool, e error) {
	for pos := curIndex; pos < len(payload); pos++ {
		switch payload[pos] {
		case '"':
			if key, pos, e = unescapeString(payload, pos+1); e != nil {
				return
			}
		case ':':
			newPos, hasKey = pos+1, true
			return
		case '}': // end of an object, it's an empty {}
			return
		}
	}
	e = ErrInvalidJSONPayload
	return
}

// valStart and valEnd is meant to be used to get a slice out of payload, newPos always
// points to a position that the next process should begin at.
//
// valStart usually is equal to curIndex, except curIndex is a white space,
// e.g. '"key":  234  ', the valStart is at '2', not the ' ' right after ':',
// as for valEnd, end at the ' ' right after '4' not the vary last ' '.
//
// hasNext is true means has next element or key set.
//
// PS: If the payload is a valid json payload, this function will definitely not fail, so
// we don't need do error-check in it.
//
// If payload is an array, then the caller must skip the array opening '[', means curIndex must + 1 or like so
// before passing it.
//
func nextValue(payload []byte, curIndex int) (newPos, valStart, valEnd int, vtype valType, hasNext, emptyArray bool, e error) {

	// a value must ends with one of this delimiters, ',' or '}' or ']'.
	for newPos = curIndex; newPos < len(payload); newPos++ {
		switch payload[newPos] {
		/* Delimiters */
		case ',': // done
			hasNext = true
			goto done
		case '}', ']': // done
			goto done

		/* End of Delimiters */

		case '"':
			valStart = newPos
			for newPos++; newPos < len(payload); newPos++ {
				if b := payload[newPos]; b == '\\' {
					newPos++
				} else if b == '"' {
					break
				}
			}
			valEnd, vtype = newPos+1, valString
		case '{':
			valStart = newPos
			if newPos, e = readToClose(payload, '{', '}', newPos); e != nil {
				return
			}
			valEnd, vtype = newPos+1, valObject
		case '[':
			valStart = newPos
			if newPos, e = readToClose(payload, '[', ']', newPos); e != nil {
				return
			}
			valEnd, vtype = newPos+1, valArray
		case ' ', '\n', '\t', '\r' /* , '\f', '\b' */ : // white characters
		default:
			// numbers
			if b := payload[newPos]; b >= '0' && b <= '9' || b == '-' || b == '.' {
				valStart, vtype = newPos, valNumber
				if b == '.' {
					vtype = valFloat
				}

			} else if b == 'f' { // false
				valStart, vtype = newPos, valFalse
				newPos += 4
			} else if b == 't' { // true
				valStart, vtype = newPos, valTrue
				newPos += 3
			} else if b == 'n' { // null
				valStart, vtype = newPos, valNull
				newPos += 3
			} else {
				e = ErrInvalidJSONPayload
				return
			}
			for newPos++; newPos < len(payload); newPos++ {
				if b := payload[newPos]; b == ',' {
					hasNext = true
					goto done
				} else if b == ']' || b == '}' {
					goto done
				} else if (b == ' ' || b == '\n' || b == '\t' || b == '\r' /*  || b == '\f' || b == '\b' */) && valEnd == 0 {
					valEnd = newPos
				} else if b == '.' {
					vtype = valFloat
				}
			}

		}
	}
	e = ErrInvalidJSONPayload
	return
done:
	emptyArray = valStart == 0
	if valEnd == 0 { // false, true,  null, or number ends with ',' '}' ']'
		valEnd = newPos
	}
	return
}

// unescapeString reads, unescapes string, or encodes it to utf8
// NOTE: caller must skip the first '"', there is say pos = the first '"' + 1
func unescapeString(payload []byte, pos int) (str string, newPos int, e error) {
	var strStart = pos
	for ; pos < len(payload); pos++ {
		if b := payload[pos]; b == '\\' {
			str += string(payload[strStart:pos]) // get rid of the '\'
			pos++
			// encode to utf8
			if b = payload[pos]; b == 'u' {
				encoded, tempPos, e := toUTF8(payload, pos+1)
				if e != nil {
					return "", 0, e
				}
				str = fmt.Sprintf("%s%s", str, encoded)
				pos, strStart = tempPos-1, tempPos
			} else {
				// unescape
				if b == 'x' {
					// hex unit \xXX
					rst, e := toDecimal(payload, pos+1, pos+3)
					if e != nil {
						return "", 0, e
					}
					str = fmt.Sprintf("%s%s", str, []byte{byte(rst)})
					pos += 2
					strStart = pos + 1
					continue
				}
				strStart = pos + 1
				switch b {
				case 'n':
					str += "\n"
				case 't':
					str += "\t"
				case 'r':
					str += "\r"
				case 'f':
					str += "\f"
				case 'b':
					str += "\b"
				default:
					str += string(b)
				}
			}
		} else if b == '"' {
			str += string(payload[strStart:pos])
			newPos = pos
			return
		}
	}
	e = ErrInvalidJSONPayload
	return
}

var (
	whiteSpaceSize = 128
	whiteSpaces    = bytes.Repeat([]byte{' '}, whiteSpaceSize)
	emptyBytes     = []byte{}
)

// updatePayload removes or writes part of the p.payload,
// if payload has enought spaces to write it won't allocate new memory for the payload.
// e.g. removing is definitely not allocating new memory, it removes bytes in payload by whiting them.
//
// When calling for removal, newVal must be nil.
func updatePayload(payload []byte, newVal []byte, start, end, rootEnd int) (updated []byte, newEnd, newRootEnd int) {

	var newLen int
	if newVal == nil { // removing
		newVal = emptyBytes
	}
	newLen = len(newVal)
	// count the length needed.
	offset := end - start - newLen
	space := len(payload) - (rootEnd - offset)
	if space >= 0 {
		updated = payload
	} else {
		updated = make([]byte, len(payload)+-space)
		copy(updated, payload)
	}
	// this line must run first, otherwise when offset is larger than 0,
	// the payload would be corrupted.
	copy(updated[end-offset:], updated[end:rootEnd])

	copy(updated[start:], newVal)
	start += newLen
	newEnd, newRootEnd = start, rootEnd-offset

	// fill the remaining space with white characters
	for offset = newRootEnd; offset < len(updated); offset += whiteSpaceSize {
		copy(updated[offset:], whiteSpaces)
	}
	return
}

func toMap(payload []byte, start, end int) (m map[string]interface{}, e error) {

	m = map[string]interface{}{}
	var val interface{}
	var key string
	var vStart, vEnd int
	var vType valType
	var next, hasKey bool
	// pos + 1 skip the {
	for pos := start + 1; pos < end; {
		if pos, key, hasKey, e = nextKey(payload, pos, true); e != nil {
			return
		} else if !hasKey {
			return // done.
		}
		if pos, vStart, vEnd, vType, next, _, e = nextValue(payload, pos); e != nil {
			return
		}
		if val, e = fromJSON(payload, vStart, vEnd, vType); e != nil {
			return
		}
		m[key] = val
		if next {
			pos++
		} else {
			return
		}
	}
	e = ErrInvalidJSONPayload
	return
}

func fromJSON(payload []byte, start, end int, vtype valType) (val interface{}, e error) {
	switch vtype {
	case valString:
		val, _, e = unescapeString(payload, start+1)
	case valNumber:
		return parseInt(payload, start, end)
		// val, e = strconv.ParseInt(string(payload[start:end]), 10, 64)
	case valFloat:
		val, e = strconv.ParseFloat(string(payload[start:end]), 64)
	case valArray:
		val, e = interfaceArray(payload, start, end)
	case valObject:
		val, e = toMap(payload, start, end)
	case valFalse:
		val = false
	case valTrue:
		val = true
	case valNull:
		// don't do nothing.
	default:
		panic(fmt.Errorf("The valType:%d should be added in to case", vtype))
	}
	return
}

// only supports string, numerical values(e.g. int, float...), boolean, nil, map or array of them
func toJSON(val interface{}) (j []byte, jtype valType, e error) {
	if val == nil {
		j = []byte("null")
		return
	}
	switch v := val.(type) {
	case string:
		j, jtype = []byte(fmt.Sprintf(`"%s"`, escape(v))), valString
	case int, int64, uint, uint64, uint8 /* byte, */, int8, uint16, int16, uint32, int32 /* rune */ :
		j, jtype = []byte(fmt.Sprintf("%d", v)), valNumber
	case bool:
		j = []byte(fmt.Sprintf("%t", v))
		if v {
			jtype = valTrue
		} else {
			jtype = valFalse
		}
	case float64:
		j, jtype = []byte(strconv.FormatFloat(v, 'f', -1, 64)), valFloat
	case float32:
		j, jtype = []byte(strconv.FormatFloat(float64(v), 'f', -1, 32)), valFloat
	case []byte:
		j, jtype = []byte(fmt.Sprintf("%q", string(v))), valString
	case []string:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, str := range v {
			j = append(j, []byte(fmt.Sprintf(`"%s",`, escape(str)))...)
		}
		j[len(j)-1] = ']'

	case []int:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, i := range v {
			j = append(j, []byte(fmt.Sprintf("%d,", i))...)
		}
		j[len(j)-1] = ']'

	case []interface{}:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		var jval []byte
		// validate vals
		for _, val := range v {
			if jval, _, e = toJSON(val); e != nil {
				return
			}
			j = append(j, jval...)
			j = append(j, ',')
		}
		j[len(j)-1] = ']'

	case []int64:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, i := range v {
			j = append(j, []byte(fmt.Sprintf("%d,", i))...)
		}
		j[len(j)-1] = ']'

	case map[string]interface{}:
		var objJSON string
		for key, val := range v {
			if j, _, e = toJSON(val); e != nil {
				return
			}
			objJSON += fmt.Sprintf(`"%s":%s,`, escape(key), string(j))
		}
		objJSON = objJSON[:len(objJSON)-1] // get rid of the last comma
		j, jtype = []byte(fmt.Sprintf("{%s}", objJSON)), valObject
	case []map[string]interface{}:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		var mapJSON []byte
		for _, m := range v {
			if mapJSON, _, e = toJSON(m); e != nil {
				return
			}
			j = append(j, mapJSON...)
			j = append(j, ',')
		}
		j[len(j)-1] = ']'

	case []float64:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, f := range v {
			j = append(j, []byte(fmt.Sprintf("%s,", strconv.FormatFloat(f, 'f', -1, 64)))...)
		}
		j[len(j)-1] = ']'
	case []float32:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, f := range v {
			j = append(j, []byte(fmt.Sprintf("%s,", strconv.FormatFloat(float64(f), 'f', -1, 32)))...)
		}
		j[len(j)-1] = ']'
	case []bool:
		if len(v) == 0 {
			j, jtype = []byte{'[', ']'}, valArray
			return
		}
		j, jtype = []byte{'['}, valArray
		for _, b := range v {
			if b {
				j = append(j, 't', 'r', 'u', 'e', ',')
			} else {
				j = append(j, 'f', 'a', 'l', 's', 'e', ',')
			}
		}
		j[len(j)-1] = ']'
	default:
		e = fmt.Errorf("type %T is unsupported", v)
	}
	return
}

const (
	int32Max = "2147483647"
	int64Max = "9223372036854775807"
	uintMax  = "18446744073709551615"
)

// type of n would be one of the int, int64 or uint64
func parseInt(payload []byte, start, end int) (n interface{}, e error) {
	var numLen = end - start
	if payload[0] == '-' {
		numLen--
	}

	if ln := numLen - len(int32Max); ln < 0 { // trying to parse into int
		return strconv.Atoi(string(payload[start:end]))
	} else if ln == 0 {
		if strconv.IntSize == 64 {
			return strconv.Atoi(string(payload[start:end]))
		}
		// in 32bits system.
		for i := start; i < end; i++ {
			j := i - start
			if payload[i] > int32Max[j] {
				return strconv.ParseInt(string(payload[start:end]), 10, 64) // to int64
			}
		}
		return strconv.Atoi(string(payload[start:end]))
	} else if ln := numLen - len(int64Max); ln < 0 { // trying to parse into int64
		if strconv.IntSize == 64 {
			return strconv.Atoi(string(payload[start:end]))
		}
		// in 32bits system.
		return strconv.ParseInt(string(payload[start:end]), 10, 64)
	} else if ln == 0 {
		for i := start; i < end; i++ {
			j := i - start
			if payload[i] > int64Max[j] {
				return strconv.ParseUint(string(payload[start:end]), 10, 64) // paser into uint64
			}
		}
		if strconv.IntSize == 64 {
			return strconv.Atoi(string(payload[start:end])) // int is 64 bits, so we parse it into int
		}
		return strconv.ParseInt(string(payload[start:end]), 10, 64)
	}
	return strconv.ParseUint(string(payload[start:end]), 10, 64) // paser into uint64
}

func appendElements(payload []byte, start, end, rootEnd int, vtype valType, vals ...interface{}) (newPayload []byte,
	newEnd, newRootEnd int, e error) {

	var j []byte
	var valJSON = []byte{}
	for _, val := range vals {
		if j, _, e = toJSON(val); e != nil {
			return
		}
		valJSON = append(valJSON, j...)
		valJSON = append(valJSON, ',')
	}
	valJSON[len(valJSON)-1] = ']'
	newPayload, newEnd, newRootEnd = appendJSON(payload, valJSON, vtype, start, end, rootEnd)
	return
}

// NOTE: json must ends with a '}' or ']' if payloadType is Object or Array
func appendJSON(payload, json []byte, payloadType valType, start, end, rootEnd int) (newPayload []byte, newEnd, newRootEnd int) {
	var appendAt int
	if payloadType == valObject {
		if json[len(json)-1] != '}' {
			panic("appendJSON: json must ends with a '}'")
		}
		appendAt = end - 1
	} else if payloadType == valArray {
		if json[len(json)-1] != ']' {
			panic("appendJSON: json must ends with a ']'")
		}
		appendAt = end - 1
	} else {
		appendAt = end
	}
	// check if there is necessary add a comma,
	// e.g. the current array [1, 2], we need a comma for adding elements,
	// but if it's [] then we don't need comma.
	var addComma bool
	for i := end - 2; i > start; i-- {
		b := payload[i]
		if addComma = b != ' ' && b != '\t' && b != '\n' && b != '\r'; /* && b != '\f' && b != '\b' */ addComma {
			break
		}
	}
	if addComma {
		j := make([]byte, 1+len(json))
		j[0] = ','
		copy(j[1:], json)
		json = j
	}
	newPayload, newEnd, newRootEnd = updatePayload(payload, json, appendAt, end, rootEnd)
	return
}

func merge(payload []byte, start, end, rootEnd int, preserve bool, val map[string]interface{}) (newPayload []byte,
	newEnd, newRootEnd int, e error) {

	var oldKey string
	var next, empty bool
	var vStart, vEnd int
	var newValType, oldValType valType
	var newValJSON []byte
readNewKey:
	for newKey, newVal := range val {
		if newValJSON, newValType, e = toJSON(newVal); e != nil {
			return
		}
	readOldKey:
		for i := start + 1; i < end; i++ {
			if i, oldKey, next, e = nextKey(payload, i, true); e != nil {
				return
			} else if next {
				if i, vStart, vEnd, oldValType, next, empty, e = nextValue(payload, i); e != nil {
					return
				} else if empty {
					// no value is found, set vStar and vEnd with key end
					vStart, vEnd, oldValType = i, i, valNull
				}
			}
			if oldKey != newKey {
				if !next {
					// this new key doesn't match with any old keys, so add it to the object end directly
					newValJSON = []byte(fmt.Sprintf(`"%s":%s}`, escape(newKey), string(newValJSON)))
					payload, end, rootEnd = appendJSON(payload, newValJSON, valObject, start, end, rootEnd)
					continue readNewKey
				}
				continue readOldKey
			}
			if !preserve { // don't preseve the old value, overwrite it directly.
				payload, newEnd, rootEnd =
					updatePayload(payload, newValJSON, vStart, vEnd, rootEnd)
				end = end + (newEnd - vEnd)
				continue readNewKey
			} else {
				// merging
				// merge the oVal and newVal .
				switch oldValType {
				case valArray:
					if newValType == valArray { // append newval into array
						newValJSON = newValJSON[1:] // skip the [
					} else {
						newValJSON = append(newValJSON, ']')
					}
					payload, newEnd, rootEnd =
						appendJSON(payload, newValJSON, oldValType, vStart, vEnd, rootEnd)

					end = end + (newEnd - vEnd)
					continue readNewKey
				default:
					goto merge_to_array
					/* 				case valString, valTrue, valFalse, valNumber, valFloat, valNull: // merge to array
					   					goto merge_to_array
					   				case valObject:
					   					if newValType == valObject {
					   						if payload, end, rootEnd, e =
					   							merge(payload, vStart, vEnd, rootEnd, preserve, newVal.(map[string]interface{})); e != nil {
					   							return
					   						}
					   						continue readNewKey
					   					} else {
					   						goto merge_to_array
					   					} */
				}
				e = ErrInvalidJSONPayload
				return
				// must jump over
			merge_to_array:
				oldValJSON := payload[vStart:vEnd]
				if newValType == valArray {
					if len(newValJSON) == 2 { // empty array
						newValJSON = append(newValJSON[1:], oldValJSON...)
					} else {
						newValJSON[1] = ','
						newValJSON = append(newValJSON, oldValJSON...)
					}
					newValJSON = append(newValJSON, ']')
				} else {
					j := make([]byte, 3+len(oldValJSON)+len(newValJSON))
					j[0] = '['
					copy(j[1:], oldValJSON)
					i := 1 + len(oldValJSON)
					j[i], i = ',', i+1
					copy(j[i:], newValJSON)
					i += len(newValJSON)
					j[i] = ']'
					newValJSON = j
				}
				payload, newEnd, rootEnd =
					updatePayload(payload, newValJSON, vStart, vEnd, rootEnd)
				end = end + (newEnd - vEnd)
				continue readNewKey

			}
		}
	}
	newPayload, newEnd, newRootEnd = payload, end, rootEnd
	return
}

// for object { and array [
func readToClose(payload []byte, opener, closer byte, currentIndex int) (newPos int, e error) {
	var level int
	for newPos = currentIndex + 1; newPos < len(payload); newPos++ {
		switch payload[newPos] {
		case opener:
			level++
		case closer:
			if level == 0 {
				return
			}
			level--
		case '"':
			for newPos++; newPos < len(payload); newPos++ {
				if b := payload[newPos]; b == '\\' {
					newPos++
				} else if b == '"' {
					break
				}
			}
		}
	}
	e = ErrInvalidJSONPayload
	return
}

func skipWhites(payload []byte, pos int) (newPos int, ok bool) {
	for newPos = pos; newPos < len(payload); newPos++ {
		switch payload[newPos] {
		case ' ', '\t', '\n', '\r' /* , '\f', '\b' */ : // white characters.
			continue
		}
		ok = true
		return
	}
	return
}

func skipWhitesLast(payload []byte) (newPos int, ok bool) {
	for newPos = len(payload) - 1; newPos >= 0; newPos-- {
		switch payload[newPos] {
		case ' ', '\t', '\n', '\r' /* , '\f', '\b' */ : // white characters.
			continue
		}
		ok = true
		return
	}
	return
}

/********* Encoding  */

const (
	hex = "0123456789abcdef"
)

// copy from encoding/json HTMLEscape because it's 3x faster than the my wrote one.
func escape(src string) string {
	var dst bytes.Buffer
	start := -1
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c == '<' || c == '>' || c == '&' {
			if start > -1 {
				dst.WriteString(src[start:i])
				start = -1
			}
			dst.WriteString(`\u00`)
			dst.WriteByte(hex[c>>4])
			dst.WriteByte(hex[c&0xF])
			continue
		}
		// Convert U+2028 and U+2029 (E2 80 A8 and E2 80 A9).
		if c == 0xE2 && i+2 < len(src) && src[i+1] == 0x80 && src[i+2]&^1 == 0xA8 {
			if start > -1 {
				dst.WriteString(src[start:i])
				start = -1
			}
			dst.WriteString(`\u202`)
			dst.WriteByte(hex[src[i+2]&0xF])
			i += 2
			continue
		}

		if c == '"' {
			if start > -1 {
				dst.WriteString(src[start:i])
				start = -1
			}
			dst.WriteString(`\"`)
			continue
		}

		if c == '\\' {
			if start > -1 {
				dst.WriteString(src[start:i])
				start = -1
			}
			if i++; i < len(src) {
				c = src[i]
				if c == '\\' {
					dst.WriteString(`\\`) // just write doulbe slashes
					continue
				}
				if c == '"' || c == 'n' || c == 'r' || c == 't' ||
					c == 'u' || c == 'x' || c == '/' {

					dst.WriteByte('\\')
					dst.WriteByte(c)
					continue
				}
				i--
			}
			dst.WriteString(`\\`) // escape the \
			continue
		}

		// other unwanted characters
		if start == -1 {
			start = i
		}
	}
	if start > -1 {
		if start == 0 { // didn't escape any characters, just return src
			return src
		}
		dst.WriteString(src[start:])
	}
	return dst.String()
}

// hex to decimal
func toDecimal(payload []byte, start, end int) (decimal int, e error) {
	var pow = float64(end - start)
	for ; start < end; start++ {
		pow--
		if b := payload[start]; b >= '0' && b <= '9' {
			decimal += int(b-'0') * int(math.Pow(16, pow))
		} else if b >= 'a' && b <= 'f' {
			decimal += int(b-'a'+10) * int(math.Pow(16, pow))
		} else if b >= 'A' && b <= 'F' {
			decimal += int(b-'A'+10) * int(math.Pow(16, pow))
		} else {
			e = ErrInvalidJSONPayload
			return
		}
	}
	return
}

// transform unicode string \uxxxx to utf8, start must skip the \u prefix.
func toUTF8(payload []byte, start int) (result string, end int, e error) {
encoding:
	var w1, decimal int
	var surrogate bool
to_decimal:
	end = start + 4
	// convert hex to decimal
	pow := float64(4)
	for ; start < end; start++ {
		pow--
		if b := payload[start]; b >= '0' && b <= '9' {
			decimal += int(b-'0') * int(math.Pow(16, pow))
		} else if b >= 'a' && b <= 'f' {
			decimal += int(b-'a'+10) * int(math.Pow(16, pow))
		} else if b >= 'A' && b <= 'F' {
			decimal += int(b-'A'+10) * int(math.Pow(16, pow))
		} else {
			e = ErrInvalidJSONPayload
			return
		}
	}

	// check if it's the first code unit of utf16 surrogate
	if decimal >= 0xD800 && decimal <= 0xDBFF {
		if len(payload)-start > 6 { // check there \u following
			if payload[start] == '\\' && payload[start+1] == 'u' {
				start += 2 // skip the \u
				surrogate = true
				w1, decimal = decimal, 0
				goto to_decimal
			}
		}
	}
	if surrogate {
		decimal = 0x10000 | (w1 & 0x3ff << 10) | (decimal & 0x3ff)
	}

	if decimal > 0x10FFFF {
		// invalid unicode
		return
	}
	// encode to utf8
	if decimal >= 0x10000 {
		result += string([]byte{
			0xf0 | (0x7 & byte(decimal>>18)),  // 1111 0xxx
			0x80 | (0x3f & byte(decimal>>12)), // 10 xxxxxx
			0x80 | (0x3f & byte(decimal>>6)),  // 10 xxxxxx
			0x80 | (0x3f & byte(decimal)),     // 10 xxxxxx
		})

	} else if decimal >= 0x800 {
		result += string([]byte{
			0xe0 | (0xf & byte(decimal>>12)), // 1110 xxxx
			0x80 | (0x3f & byte(decimal>>6)), // 10 xxxxxx
			0x80 | (0x3f & byte(decimal)),    // 10 xxxxxx
		})
	} else if decimal >= 0x80 {
		// 110 & 10
		result += string([]byte{0xc0 | (0x1f | byte(decimal>>6)), 0x80 | (0x3f & byte(decimal))})

	} else {
		// numbers or letters
		result += string([]byte{byte(decimal)})
	}
	if len(payload)-start >= 6 && payload[start] == '\\' && payload[start+1] == 'u' {
		start += 2 // still got unicode to encode
		goto encoding
	}
	return
}

/***********End of encoding */

/*********** Validation functions   */
const (
	statusAry = '['
	statusObj = '{'
	statusKey = 'K'
	statusVal = 'V'

	statusValEnd = 'v'
)

// Validate checks payload is a valid json encoding.
func Validate(payload []byte) (e error) {
	_, _, e = validate(payload, statusVal, 0)
	return
}

func validate(payload []byte, status byte, pos int) (newPos int, closed bool, e error) {
	oStatus := status

validating:
	for newPos = pos; newPos < len(payload); newPos++ {
		switch b := payload[newPos]; b {
		case '{':
			if status != statusAry && status != statusVal {
				goto fail
			}
			if newPos, closed, e = validate(payload, statusObj, newPos+1); e != nil {
				return
			} else if !closed {
				goto fail
			}
			status, closed = statusValEnd, false
		case '[':
			if status != statusAry && status != statusVal {
				goto fail
			}
			if newPos, closed, e = validate(payload, statusAry, newPos+1); e != nil {
				return
			} else if !closed {
				goto fail
			}
			// closed = false
			status, closed = statusValEnd, false
		case '}':
			if status != statusObj && status != statusValEnd || oStatus != statusObj {
				goto fail
			}
			closed = true
			return

		case ']':
			if status != statusAry && status != statusValEnd || oStatus != statusAry {
				goto fail
			}
			closed = true
			return

		case '"':
			var ok bool
		readQuotes:
			for newPos++; newPos < len(payload); newPos++ {
				switch payload[newPos] {
				case '\\': // escape
					newPos++
					var remain int8
					if b := payload[newPos]; b == 'u' { //
						// validate unicode, \u must followed by a four-hex-digit string.
						remain = 4
					} else if b == 'x' { // escaped hex unit \x+two-hex-digit
						remain = 2
					} else {
						continue readQuotes
					}
					// validate hex digits
					for newPos++; newPos < len(payload); newPos++ {
						// if b not in the range of 0-9, a-f or A-F then fail.
						if b := payload[newPos]; (b < '0' || b > '9') && (b < 'a' || b > 'f') && (b < 'A' || b > 'F') {
							goto fail // invalid unicode code point or escaped hex unit.
						}
						if remain--; remain == 0 {
							continue readQuotes
						}
					}
				case '"':
					ok = true
					break readQuotes
				}
			}
			if !ok {
				goto fail
			}
			if status == statusKey || status == statusObj {
				status = statusKey
			} else if status == statusVal || status == statusAry {
				status = statusValEnd
			} else {
				goto fail
			}

		case ':':
			if status != statusKey { // follows a key
				goto fail
			}
			status = statusVal
		case ',':
			// comma only follows a value
			if status != statusValEnd {
				goto fail
			}
			if oStatus == statusObj {
				status = statusKey
			} else if oStatus == statusAry {
				status = statusVal
			} else {
				goto fail
			}

		case ' ', '\t', '\n', '\r' /* , '\f', '\b' */ : // white characters.
			continue
		default: // numbers, false, true, null
			if status != statusVal && status != statusAry {
				goto fail
			} else {
				if b := payload[newPos]; (b >= '0' && b <= '9') || b == '-' || b == '.' {
					// number
					pos = newPos
				numbering:
					for newPos++; newPos < len(payload); newPos++ {
						switch payload[newPos] {
						case ',', '}', ']':
							if !validateNumber(payload, pos, newPos) {
								goto fail
							}
							status, newPos = statusValEnd, newPos-1
							continue validating
						case ' ', '\t', '\n', '\r' /* , '\f', '\b' */ : // white characters.
							break numbering
						}
					}
					if !validateNumber(payload, pos, newPos) {
						goto fail
					}
				} else if b == 'f' { // false
					if newPos += 4; newPos >= len(payload) {
						goto fail
					} else if payload[newPos-3] != 'a' || payload[newPos-2] != 'l' ||
						payload[newPos-1] != 's' || payload[newPos] != 'e' {
						goto fail
					}

				} else if b == 't' { // true
					if newPos += 3; newPos >= len(payload) {
						goto fail
					} else if payload[newPos-2] != 'r' || payload[newPos-1] != 'u' ||
						payload[newPos] != 'e' {
						goto fail
					}

				} else if b == 'n' { // null
					if newPos += 3; newPos >= len(payload) {
						goto fail
					} else if payload[newPos-2] != 'u' || payload[newPos-1] != 'l' ||
						payload[newPos] != 'l' {
						goto fail
					}
				} else {
					goto fail
				}
				// loop to the value delimiters, the ',' , '}' and ']'
				for newPos++; newPos < len(payload); newPos++ {
					switch payload[newPos] {
					case ',', '}', ']':
						// leave the ',' and ] and } to the main loop handle.
						status, newPos = statusValEnd, newPos-1
						continue validating
					case ' ', '\t', '\n', '\r' /* , '\f', '\b' */ : // white characters.
					default:
						goto fail
					}
				}
				status = statusValEnd
			}
		}
	}
	if status != statusValEnd {
		goto fail
	}
	return // success
fail:
	e = getErrorInfo(payload, newPos)
	return
}

func getErrorInfo(payload []byte, i int) (e error) {
	if i < len(payload) {
		i++
	}
	// gen error info, line and position.
	line := bytes.Count(payload[:i], []byte{'\n'})
	pos := bytes.LastIndex(payload[:i], []byte{'\n'})
	if pos == -1 {
		pos = 0
	}
	pos = i - pos
	if i > 10 {
		e = fmt.Errorf("Error occured at line: %d, pos:%d, around: %s",
			line+1, pos+1, string(payload[i-10:i]))

	} else {
		e = fmt.Errorf("Error occured at line: %d, pos:%d, around: %s",
			line+1, pos+1, string(payload[:i]))
	}
	return
}
func validateNumber(payload []byte, pos, end int) bool {
	var notFractional, fr, point, exponent, digit, hasDigit bool
	b := payload[pos]
	if end-pos >= 2 {
		// check zero prefix. e.g. 01, this kind of number not allowed has fractional part.
		notFractional = b == '0' && payload[pos+1] != '.'
	}
	// check starts with
	if digit = b >= '0' && b <= '9'; !digit {
		if point = b == '.'; !point {
			if b != '-' {
				return false // fail, number must starts with any of 0-9, '-' or '.'
			}
		}
	} else {
		hasDigit = true
	}
	for pos++; pos < end; pos++ {
		b = payload[pos]
		if digit = b >= '0' && b <= '9'; digit {
			hasDigit = true
			continue
		}
		if b == 'e' || b == 'E' { // validate exponent part
			if exponent {
				return false // fail
			}
			exponent = true
			if pos++; pos < len(payload) {
				if b = payload[pos]; b == '-' || b == '+' {
					continue
				} else if b >= '0' && b <= '9' {
					continue
				}
			} else {
				return false // ends with exponent, fail
			}
		}
		if point = b == '.'; point { // validate fractional part
			if exponent || fr || notFractional {
				// exponent     : e.g. 1e.123, exponent doesn't allow fractional part
				//  decimal     : e.g. 1.0.1, more than one point
				// notFractional: e.g. 01.1, number starts with 0 but more than two digits.
				return false
			}
			fr = true
			continue
		}
		return false // fail, illegal characters.
	}
	return hasDigit && (digit || point) // number must ends with a digit or a point.
}

/********* End of  validation functions*/

// Minify minifys the json document.
func Minify(json []byte) (minified []byte) {
	if len(json) == 0 {
		return json
	}
	end := minify(json, false)
	return json[:end]
}

func minify(payload []byte, keepSpace bool) (end int) {
	if len(payload) == 0 {
		return
	}
	var i, j, start int
	start = -1
	ln := len(payload)
	// minified = make([]byte, len(json))
	for i = 0; i < ln; i++ {
		b := payload[i]
		if keepSpace && b == ',' || b == ':' { // keep the space follows ',' or ':'
			i++
			if i < ln && payload[i] == ' ' {
				if start == -1 {
					start = i
				}
				continue
			}
			i--
		}
		switch b {
		case '"':
			if start == -1 {
				start = i
			}
			for i++; i < ln; i++ {
				if b := payload[i]; b == '\\' {
					i++
				} else if b == '"' {
					break
				}
			}
		case ' ', '\t', '\n', '\r', '\f', '\b': // white characters.
			if start > -1 {
				copy(payload[j:], payload[start:i])
				j += i - start
				start = -1
			}
			continue
		default:
			if start == -1 {
				start = i
			}
		}
	}
	if start > -1 {
		copy(payload[j:], payload[start:i])
		j += i - start
	}
	// playload = playload[:j]
	end = j
	return
}

// Prettify prettifys the json document.
func Prettify(json []byte, indet int) (prettified []byte) {
	minIdent := indet
	indet = 0
	var wrSpace = []byte{',', ' '}
	var buf []byte
	var curPos, cap, allocSize int
	allocSize = bytes.Count(json, []byte{'{'})*minIdent*2 + 1 +
		bytes.Count(json, []byte{'['})*minIdent*2 + 1
	cap = allocSize + len(json)
	buf = make([]byte, cap)

	writeByte := func(b byte) {
		if curPos+1 >= cap {
			cap = cap + 1 + allocSize
			temp := make([]byte, cap)
			copy(temp, buf)
			buf = temp
		}
		buf[curPos] = b
		curPos++
	}
	write := func(b []byte) {
		if curPos+len(b) >= cap {
			cap = cap + len(b) + allocSize
			temp := make([]byte, cap)
			copy(temp, buf)
			buf = temp
		}
		copy(buf[curPos:], b)
		curPos += len(b)
	}

	newline := func() {
		writeByte('\n')
		if indet <= whiteSpaceSize {
			write(whiteSpaces[:indet])
		} else {
			for i := indet; i > 0; i -= whiteSpaceSize {
				write(whiteSpaces[:whiteSpaceSize])
			}
		}
	}

	indent := func(opener byte) {
		indet += minIdent
		writeByte(opener)
		newline()
	}
	cancelIndent := func(startPos int, endPos int) (newPos int) {
		newPos = minify(buf[startPos:endPos], true)
		curPos = startPos + newPos
		return
	}
	close := func(closer byte) {
		indet -= minIdent
		newline()
		writeByte(closer)
	}
	var prettify func(pos int, opener byte) (newPos, length, items, arrayOrObjectNum int, done bool)
	prettify = func(pos int, opener byte) (newPos, length, items, arrayOrObjectNum int, done bool) {
	readJSON:
		for newPos = pos; newPos < len(json); newPos++ {
			if b := json[newPos]; b == '}' || b == ']' {
				if opener+2 == b { // '[' and '{' + 2 == ']' and '}'
					close(b)
					return
				} else {
					writeByte(b)
				}
				continue
			} else if b == '[' || b == '{' {
				tempCurPos := curPos + 1 // this code must at this line.
				arrayOrObjectNum, items = arrayOrObjectNum+1, items+1

				indent(b)
				pos = newPos + 1
				var curLen, curItems, curAONum int
				if newPos, curLen, curItems, curAONum, done = prettify(newPos+1, b); done {
					return
				} else if b == '[' && curLen < 60 && curAONum == 0 {
					cancelIndent(tempCurPos, curPos)
				} else if b == '{' && curItems == 0 {
					cancelIndent(tempCurPos, curPos)
				}
				length = length + (newPos - pos)
				continue
			}
			switch json[newPos] {
			case '"':
				for pos := newPos + 1; pos < len(json); pos++ {
					if b := json[pos]; b == '\\' {
						pos++
					} else if b == '"' {
						write(json[newPos : pos+1])
						newPos, length = pos, length+(pos-newPos)
						continue readJSON
					}
				}
				write(json[newPos:]) // unexpected end
				break readJSON
			case ':':
				wrSpace[0] = ':'
				write(wrSpace)
				length++
			case ',':
				wrSpace[0] = ','
				write(wrSpace)
				newline()
				items, length = items+1, length+1
			case ' ', '\n', '\t', '\r', '\f', '\b':
				continue
			default:
				pos, endPos := newPos+1, 0
				for ; pos < len(json); pos++ {
					switch json[pos] {
					case ']', '}', ',':
						if endPos == 0 {
							endPos = pos
						}
						write(json[newPos:endPos])
						newPos, length = pos-1, length+(endPos-newPos)

						continue readJSON
					case ' ', '\n', '\t', '\r', '\f', '\b':
						if endPos == 0 {
							endPos = pos
						}
					}
				}
				write(json[newPos:]) // unexpected end
				done = true
				break readJSON
			}
		}
		return
	}
	prettify(0, ' ')
	return buf[:curPos]
}

// Marshal calls encoding/json.Marshal directly, implement this is function here
// so we don't have to import encoding/json additionally in most cases.
func Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }

// Unmarshal calls encoding/json.Marshal directly, implement this is function here
// so we don't have to import encoding/json additionally in most cases.
func Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
