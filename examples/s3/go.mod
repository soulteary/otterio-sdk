module github.com/soulteary/otterio-sdk/examples/s3

go 1.27.1

require (
	github.com/cheggaaa/pb v1.0.29
	github.com/soulteary/otterio-kits/sio v0.5.2
	// Overridden by `replace` below, to point all versions at the local otterio-sdk source, so version shouldn't matter here.
	github.com/soulteary/otterio-sdk/v7 v7.0.73
	golang.org/x/crypto v0.57.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/soulteary/otterio-kits/crc64nvme v1.1.2 // indirect
	github.com/soulteary/otterio-kits/md5-simd v1.1.3 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
)

replace github.com/soulteary/otterio-sdk/v7 => ../..
