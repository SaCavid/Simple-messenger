# Simple-messenger 

Simple messenger golang app.

Small Html, Bootstrap, Jquery chat app created for testing.

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