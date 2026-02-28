module tester

go 1.19

replace github.com/bddjr/hahosp => ../

require (
	github.com/bddjr/hahosp v0.0.0
	golang.org/x/net v0.35.0
)

require golang.org/x/text v0.22.0 // indirect
