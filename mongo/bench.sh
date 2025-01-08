#! /bin/bash

go run bench_bulk_upsert.go -size 4 -batchCount 1 -total 100000 -head true
go run bench_bulk_upsert.go -size 4 -batchCount 10 -total 100000
go run bench_bulk_upsert.go -size 4 -batchCount 100 -total 100000
go run bench_bulk_upsert.go -size 4 -batchCount 1000 -total 100000
go run bench_bulk_upsert.go -size 4 -batchCount 10000 -total 100000
go run bench_bulk_upsert.go -size 4 -batchCount 20000 -total 100000

go run bench_bulk_upsert.go -size 32 -batchCount 1 -total 100000
go run bench_bulk_upsert.go -size 32 -batchCount 10 -total 100000
go run bench_bulk_upsert.go -size 32 -batchCount 100 -total 100000
go run bench_bulk_upsert.go -size 32 -batchCount 1000 -total 100000
go run bench_bulk_upsert.go -size 32 -batchCount 10000 -total 100000
go run bench_bulk_upsert.go -size 32 -batchCount 20000 -total 100000

go run bench_bulk_upsert.go -size 128 -batchCount 1 -total 100000
go run bench_bulk_upsert.go -size 128 -batchCount 10 -total 100000
go run bench_bulk_upsert.go -size 128 -batchCount 100 -total 100000
go run bench_bulk_upsert.go -size 128 -batchCount 1000 -total 100000
go run bench_bulk_upsert.go -size 128 -batchCount 10000 -total 100000
go run bench_bulk_upsert.go -size 128 -batchCount 20000 -total 100000

go run bench_bulk_upsert.go -size 256 -batchCount 1 -total 100000
go run bench_bulk_upsert.go -size 256 -batchCount 10 -total 100000
go run bench_bulk_upsert.go -size 256 -batchCount 100 -total 100000
go run bench_bulk_upsert.go -size 256 -batchCount 1000 -total 100000
go run bench_bulk_upsert.go -size 256 -batchCount 10000 -total 100000
go run bench_bulk_upsert.go -size 256 -batchCount 20000 -total 100000
