# goroutines

In mūsēum server, there are a couple different goroutines that can sometimes be running at the same time.

## Exhibit cleanup

The exhibit cleanup cleans up expired exhibits and exhibits that have been starting for too long (as specified in the `STARTING_TIMEOUT` env variable).

## HTTP server

The HTTP server is the main goroutine and accepts incoming requests.

## SSE connection

Whenever an application is being started, a loading page is returned. This loading page connects to an SSE endpoint in mūsēum. mūsēum then spins up a goroutine for this connection, connects to NATS and forwards the starting events to the client.
