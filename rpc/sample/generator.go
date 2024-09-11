package sample

//go:generate ./tools/protoc.exe --proto_path=. --go_opt=paths=source_relative --go_out=./ echo.proto
//--go-grpc_out=./grpc --go-grpc_opt=paths=source_relative
////go:generate ./tools/pbgen.exe --tmpl_file=./rpc.tmpl --output=*.rpc.go echo.proto
