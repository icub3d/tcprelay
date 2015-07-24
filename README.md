# tcprelay

Program tcprelay relays TCP traffic between clients and servers.

# Example

Make sure your _$GOPATH/bin folder_ is in your _$PATH_, then:

    $ go get -u github.com/icub3d/tcprelay
    $ go get -u github.com/icub3d/tcprelay/echoserver
    $ go get -u github.com/icub3d/tcprelay/httpserver
    $ tcprelay
    2015/07/24 15:08:01 addr: :8000, port range: :8001-9000

    # In another terminal or if you backgrounded the previous:
    $ echoserver
    2015/07/24 15:08:11 client connection: :8001

    # In another terminal or if you backgrounded all the previous:
    $ echoserver -reverse
    2015/07/24 15:08:12 client connection: :8002

    # In another terminal or if you backgrounded all the previous:
    $ httpserver -dir=/path/to/some/static/html
    2015/07/24 15:08:13 client connection: :8003

    # In another terminal or if your backgrounded all the previous:
    $ telnet localhost 8001
    Trying ::1...
    Connected to localhost.
    Escape character is '^]'.
    hello, world!
    hello, world!
    ^]

    telnet> quit
    Connection closed.
    $ telnet localhost 8002
    Trying ::1...
    Connected to localhost.
    Escape character is '^]'.
    hello, world!

    !dlrow ,olleh^]

    telnet> quit
    Connection closed.
    $ curl localhost:8003
    # you should see your index.html. You can also try it in the broswer!

# Using

You can look at the echoserver and httpserver for examples of how to use the
relay server. They both make use of the
[github.com/icub3d/tcprelay/relay](https://godoc.org/github.com/icub3d/tcprelay/relay)
package. At a high level,

* Make a connection to the relay server.
* Get the client address that clients can use to connect.
* Start reading messages from the relay server and handle them.
* Send messages to the relay server to tell it what to do or what to send to clients.
