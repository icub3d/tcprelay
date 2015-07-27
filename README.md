# tcprelay

Program tcprelay relays TCP traffic between clients and servers.

# Example

Here is a sample session using the golang docker container:

    sudo docker run -it --rm golang
    root@adb076a42801:/go# apt-get update
    root@adb076a42801:/go# apt-get install telnet
    root@adb076a42801:/go# go get -u github.com/icub3d/tcprelay
    root@adb076a42801:/go# go get -u github.com/icub3d/tcprelay/echoserver
    root@adb076a42801:/go# go get -u github.com/icub3d/tcprelay/httpserver
    root@adb076a42801:/go# tcprelay&
    2015/07/24 21:30:44 addr: :8000, port range: :8001-9000
    [1] 121
    root@adb076a42801:/go# echoserver&
    2015/07/24 21:30:52 client connection: :8001
    [2] 124
    root@adb076a42801:/go# echoserver -reverse &
    2015/07/24 21:30:58 client connection: :8002
    [3] 127
    root@adb076a42801:/go# httpserver &
    2015/07/24 21:31:06 client connection: :8003
    [4] 130
    root@adb076a42801:/go# telnet
    telnet> open localhost 8001
    Trying ::1...
    Connected to localhost.
    Escape character is '^]'.
    hello, world!
    hello, world!
    ^]
    telnet> close
    Connection closed.
    telnet> open localhost 8002
    Trying ::1...
    Connected to localhost.
    Escape character is '^]'.
    hello, world!

    !dlrow ,olleh^]
    telnet> quit
    root@adb076a42801:/go# curl localhost:8003
    <b>hello, world wide web</b>

If you link the open port (in the above case 8003) to an external port, you can
test the httpserver in your browser!

# Developing

You can look at the echoserver and httpserver for examples of how to use the
relay server for your own servers. They both make use of the
[github.com/icub3d/tcprelay/relay](https://godoc.org/github.com/icub3d/tcprelay/relay)
package. At a high level,

* Make a connection to the relay server.
* Get the client address that clients can use to connect.
* Start reading messages from the relay server and handle them.
* Send messages to the relay server to tell it what to do or what to send to clients.

If you are using Go, you can handle the messages yourself but you might find it
easier to use  the net.Conn and net.Listener interfaces that the relay package
provides. The httpserver example does this and makes integration with existing
services extremely simple.

Non-Go servers can still make use of the relay server. Those servers just need
to be able to consume and create JSON messages for and from the relay. The
messages should be patterned after the relay.Message structure in the
documentation linked above. The documentation on for each type describes in
detail what the message is for and how it should be used. You can also review
the source code for the relay.Listener functions to see how messages can be
handled.
