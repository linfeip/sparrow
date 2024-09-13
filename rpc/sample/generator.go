package sample

//go:generate ./tools/protoc.exe --plugin=protoc-gen-sparrow=./tools/protoc-gen-sparrow.exe --proto_path=. --go_opt=paths=source_relative --sparrow_out=. --go_out=./ echo.proto
////go:generate ./tools/pbgen.exe --tmpl_file=./rpc.tmpl --output=*.rpc.go echo.proto
//--go-grpc_out=./grpc --go-grpc_opt=paths=source_relative
