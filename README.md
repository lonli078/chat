Simple chat

go build -o simple_server server/server.go

./simple_server

```
Usage of ./simple_server:
  -host string
    	Server host (default "0.0.0.0")
  -port string
    	Server port (default "7777")

```

go build -o simple_client client/client.go

./simple_client
```
Usage of ./simple_client:
  -host string
    	Server host (default "127.0.0.1")
  -name string
    	Nickname for chat (default "Foo")
  -port string
    	Server port (default "7777")

```
example for send file:
```
 /send <filename>
```
