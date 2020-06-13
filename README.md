In most cases, encoding/json is great but if you only want to read or modify a little piece of json it would be
an expensive way to do so, this is why we need hapijson comes into play.

### Install

`go get github.com/LBJ-the-GOAT/hapijson`

### Usage

#### some Getters

```javascript
jsonData := `{"name": "LBJ", "height": 2.04, "title": 3, "teams":["LAL", "CAVS"]}`

hapijson.Get(jsonData, "teams", 0)
// outputs LAL, 0 means the index of an array

hapijson.String(jsonData, "name")
// outputs LBJ

hapijson.Int(jsonData, "title")
// outputs 3

hapijson.StringArray(jsonData, "teams")
// outputs []string{"LAL", "CAVS"}

```

#### Set

```javascript
jsonData := {"teams": ["LAL"] }
jsonData, _ = hapiJSOn.Set(jsonData, "LA Lakers", "teams", 0)
// "teams": ["LAL"] becomes "teams": ["LA Lakers"]

```

#### Merge

```javascript
jsonData := {"name": "LBJ", "height": 2.04}
jsonData, _ = hapijson.Merge(jsonData, true, hapijson.Path(), "mvp", 4, "height": "6 ft. 8 in.")
// jsonData now is {"name": "LBJ", "height": [2.04, "6 ft. 8 in."], "mvp": 4},
// the value of key "height" now became an array, and key "mvp" is appended.

```

#### Append

```javascript

jsonData := {"name": "LBJ", "teams": ["LAL Lakers"]}
jsonData, _ = hapijson.Append(jsonData, hapijson.Path("teams"), "CAVS", "HEAT")
//  "teams":["LA Lakers"] becomes "teams":["LA Lakers", "CAVS", "HEAT"]
```

#### Remove

```javascript
jsonData := {"name": "LBJ", "height": 2.04}
jsonData, _ = hapijson.Remove(jsonData, "height");
// jsonData = {"name": "LBJ"}, key "height" is removed.

```

#### Clear

```javascript

jsonData := {"teams":["LA Lakers", "CAVS", "HEAT"]}
jsonData, _ = hapijson.Clear(jsonData, "teams");
//  jsonData = {"teams":[]}, "teams" array is empty now.

```

#### Increase

```javascript

jsonData := {"name": "LBJ", "title": 3}
jsonData, _ = hapijson.Incr(jsonData, 2, "title");
// jsonData ={"name": "LBJ", "title": 5},  "title": 3 is now increased by 2 to "title": 5

```

main features have been displayed above, look in [/examples](./examples) for more.

### Benchmark

Because the time complexity of getter & setter functions are kind of O(n+m),
n is the position of the last path node in the json data, m is the length of value of the path node,
so I placed the keys at the very end of the data to get **the outcomes of worst situation**.

**The results are pretty close. allocs/op mostly for keys escaping, means the more keys in path nodes the more allocs/op**

| json size | 500B                     | 2K                       | 20K                       | 280K                     | 2.8M                       | 8.6M                       |
| :-------: | ------------------------ | ------------------------ | ------------------------- | ------------------------ | -------------------------- | -------------------------- |
|    Get    | 4253 ns/op, 12 allocs/op | 7762 ns/op, 39 allocs/op | 43011 ns/op, 23 allocs/op | ≈0.4 ms/op, 29 allocs/op | ≈5.1 ms/op, 27 allocs/op   | ≈16.2 ms/op, 27 allocs/op  |
|    Set    | 5382 ns/op, 16 allocs/op | 8027 ns/op, 40 allocs/op | 62824 ns/op, 27 allocs/op | ≈0.4 ms/op, 6 allocs/op  | ≈7 ms/op, 32 allocs/op     | ≈23 ms/op, 32 allocs/op    |
|  Append   | 4677 ns/op, 14 allocs/op | 6919 ns/op, 18 allocs/op | 54962 ns/op, 24 allocs/op | ≈0.56 ms/op, 9 allocs/op | ≈6.86 ms/op, 32 allocs/op  | ≈22.96 ms/op, 34 allocs/op |
|   Merge   | 4559 ns/op, 15 allocs/op | 8876 ns/op, 40 allocs/op | 43238 ns/op, 26 allocs/op | ≈0.41 ms/op, 6 allocs/op | ≈5.05 ms/op, 28 allocs/op  | ≈20.74 ms/op, 28 allocs/op |
|  Remove   | 4286 ns/op, 12 allocs/op | 8249 ns/op, 37 allocs/op | 45354 ns/op, 23 allocs/op | ≈0.4 ms/op, 3 allocs/op  | ≈5.4 ms/op, 25 allocs/op   | ≈17.2 ms/op, 25 allocs/op  |
|   Clear   | 4194 ns/op, 12 allocs/op | 8207 ns/op, 37 allocs/op | 44915 ns/op, 23 allocs/op | ≈0.4 ms/op, 3 allocs/op  | ≈5.3 ms/op, 25 allocs/op   | ≈16.65 ms/op, 25 allocs/op |
|   Incr    | 4078 ns/op, 12 allocs/op | 5640 ns/op, 16 allocs/op | 39334 ns/op, 26 allocs/op | ≈0.56 ms/op, 8 allocs/op | ≈ 6.97 ms/op, 29 allocs/op | ≈15.94 ms/op, 29 allocs/op |
|   Valid   | 1180 ns/op, 0 allocs/op  | 3723 ns/op, 0 allocs/op  | 45410 ns/op, 0 allocs/op  | ≈0.44 ms/op, 0 allocs/op | ≈3.93 ms/op, 0 allocs/op   | ≈11.71 ms/op, 0 allocs/op  |
|  Minify   | 1364 ns/op, 0 allocs/op  | 3799 ns/op, 0 allocs/op  | 38470 ns/op, 0 allocs/op  | ≈0.39 ms/op, 0 allocs/op | ≈ 4.76 ms/op, 0 allocs/op  | ≈13.26 ms/op, 0 allocs/op  |
| Prettify  | 5848 ns/op, 2 allocs/op  | 16282 ns/op, 2 allocs/op | 131155 ns/op, 2 allocs/op | ≈1.66 ms/op, 3 allocs/op | ≈63.41 ms/op, 35 allocs/op | ≈39.44 ms/op, 2 allocs/op  |
