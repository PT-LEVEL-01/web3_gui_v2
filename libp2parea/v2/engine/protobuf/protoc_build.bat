rem protoc --go_out=. *.proto
protoc --gofast_out=. *.proto
protoc --js_out=import_style=commonjs,binary:js packet_v2.proto
pause