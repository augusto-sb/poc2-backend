module github.com/augusto-sb/poc1-backend

go 1.22.12

replace (
	example.com/auth => ./auth
	example.com/entity => ./entity
	example.com/router => ./router
)

require example.com/router v0.0.0-00010101000000-000000000000

require (
	example.com/auth v0.0.0-00010101000000-000000000000 // indirect
	example.com/entity v0.0.0-00010101000000-000000000000 // indirect
)
