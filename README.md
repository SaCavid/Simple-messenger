# Simple-messenger 

Simple messenger golang app.

Small Html, Bootstrap, Jquery chat app created for testing. ~16 hour lost for all code.

App starting Tcp, Tls, Http server.
Port numbers and other configuration can be changed in env file:
####    Tcp port:  2500
####    Tls port:  2501
####    Http port:   80

As communication protocol simple Json object have been used: 

    let message = {
        "From":"", // Message sender
        "To":"",   // Message addressed to 
        "Data":"", // Message content
        "Status":"", // Internal use for stop go routines
        "Users":"" // Simple Users list
    }

### Scenario
    1. Connected users must send 1 message for login with message.From = "Sender Username"
    2. After succesfull "Simple login" message with users list will be send to logged users.
    3. After login user can send message to all users in users list.
    4. If new user connected to network checkIn message will be send to all connected users with "users
        updated list" (Overloads server if there will be many connections)
    
### Have been used Golang:
    1. Interfaces
    2. Methods
    3. Channels
    4. Goroutines for concurrent connections

At the moment all users in a tcp, tls and http (websocket) can see and message each other.

### Feature
    1. Login system must be improved
    2. Database support must be added. Will be enough Postgres and Redis.
    3. Protocols must be improved for better network traffic and speed.
    4. Pool must be added on connect of users. It will stop overload on mass connections
        (mostly after restart of server)
    5. Channels can cach only 8 messages at moment. Must be improved with database support.
    6. Security must be improved between ports.

### Tested
    # Tested for 15-20 minutes. used RAM 0.3 - 0.4 Gb. All users connected Tls server. 
    Every user sended message every 1 second to randomly selected user. 
    Every second sended 10k * 10k messages

    example message = {
        "From":"tlsUser10000",  
        "To":"tlsUser9999", 
        "Data":"Looking for new solution",
        "Status":false,
        "Users":""
    }
    
    # Machine specifications
        Windows 10 Pro
        Intel(R) Core(TM) i7-7700 CPU @ 3.60GHz   3.60 GHz
        64-bit operating system, x64-based processor
        RAM 16.0 GB

    # Result 
    1. Number of goroutines: 20005
    2. Connected users: 10000
    3. Send messages: 31466356
    4. Received Messages: 31466365
    5. Lost packages:  9 (under 0.1% - Mostly not correct timing calculation difference)
### Test tools
### Used Resources
    https://gobyexample.com/
    https://certificatetools.com/
    https://pkg.go.dev/
    https://gist.github.com/denji/12b3a568f092ab951456
    https://getbootstrap.com/
    https://jquery.com/
    https://bootsnipp.com/snippets/1ea0N 