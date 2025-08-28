rem protoc --go_out=. *.proto
go install github.com/gogo/protobuf/protoc-gen-gofast
protoc --gofast_out=. *.proto
rem pause