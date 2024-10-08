# rpc代码生成
protoc --plugin=. --go_out=. --go_opt=paths=source_relative --sparrow_out=. echo.proto