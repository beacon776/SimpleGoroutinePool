module demo2

go 1.24.2
require workerpool v1.0.0
replace (
	workerpool v1.0.0 => ./../workerpool2
)