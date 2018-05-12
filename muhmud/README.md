# muhmud

A simple MUD.

## How to build
Just use the standard

```
$ go build .
```

To build a docker image, run:
```
$ docker build -t muhmud
```

which can then be launched via:
```
$ docker run --rm -it -p 8080:8080 muhmud 
```
