# Design

Muds (especially old-school muds) are a ton of fun. They're text-based,
colorful, and infinitely flexible. They should be fun to implement too, so why
not implement one?

## Client

It's hard to imagine any better client right now than, well, your browser. Since
all you need to do is send/consume a stream of text, why require users to
install telnet or some other client to connect? It also allows us to use richer
text interactions (like links, etc).

## Network messaging

When a player interacts with a MUD, their client is sending and receiving
messages of two different classes:

1. A command and a response (e.g. `> look` and `You see an empty room`)
2. A notification (e.g. `A wild dog enters the room!`).

In general, notifications can happen at any time (and the user wants to know
immediately) but responses only occur in reference to commands. Also, no more
than one command can occur at a time (this lets you time commands for fairness -
for example, when moving between rooms most MUDs will pause a bit before
allowing you to leave the room).

In a traditional telnet based mud the notifications and command/responses are
all sent in the same TCP connection. We can do that in javascript land with
websockets too - but to support some extra messages between client and server we
lift up messages so they aren't just raw text streams. For example, a command
interchange might be:

```
Client: {type: "request", requestId: "431", msg: "pick up sword"}
Server: {type: "response", requestId: "431", msg: "You pick up the sword."}
Server: {type: "notification", msg: "John says \"Hey, that was my sword!\""}
```

Notice the distinction between requests, responses and notifications. It's not
clear if we need that distinction yet; perhaps it's cleaner to just model it as
requests and notification messages.
