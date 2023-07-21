module github.com/wsw0108/concaveman

go 1.19

require (
	github.com/tidwall/rbush v0.0.0-20170518135853-2d1bd98b0ed1
	github.com/tidwall/rtree v1.10.0
	github.com/tidwall/tinyqueue v0.1.1
)

require github.com/tidwall/geoindex v1.7.0 // indirect

replace github.com/tidwall/rbush v0.0.0-20170518135853-2d1bd98b0ed1 => github.com/wsw0108/rbush v0.1.1
