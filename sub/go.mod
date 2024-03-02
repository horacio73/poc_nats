module gaudiumsoftware/nats/sub

go 1.22.0

require (
	gaudiumsoftware/nats/util v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.7.1
	github.com/nats-io/nats.go v1.32.0
)

require (
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace gaudiumsoftware/nats/util => ../util
