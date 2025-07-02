testdata_coverage:
	go test -C ./testdata/m ./... -coverprofile=coverage_count -covermode=count
	go test -C ./testdata/m ./... -coverprofile=coverage_set -covermode=set
	go test -C ./testdata/m ./... -coverprofile=coverage_atomic -covermode=atomic

update_golden:
	go test . -update

test_cov:
	go test ./... -coverprofile=coverage.out -covermode=atomic -coverpkg=./...

run:
	go run . -c coverage.out
