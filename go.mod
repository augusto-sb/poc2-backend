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
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
