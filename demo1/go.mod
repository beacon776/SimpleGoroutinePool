module demo1

go 1.24.2
require workerpool1 v1.0.0
replace (
	workerpool1 v1.0.0 => ./../workerpool1
)
