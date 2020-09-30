# telnet_chat_server
Chat server over telnet. Use telnet cli for client.

### To build telnet chat server
    go build .

### To run telnet chat server
    go mod download
    go build .
    ./telnet_chat_server -ip 127.0.0.1 -port 5000
    
### To connect to server
    telnet <server_ip> <server_port>

### Features:
- Multiple rooms
- Logging
- Save chat logs to files (one per chat room)

Note: Uses standard libraries, nothing in addition.