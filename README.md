# snowBypass

> **Fast Golang Tool try differnt methods to bypass 403**


![Image of Cngo](https://i.imgur.com/qH1uDbi.png)

# Install
```
$ go install github.com/yghonem14/snowBypass@latest
```

## Basic Usage
snowBypass accepts only stdin (just right now) :

```
$ cat urls | snowBypass
[X-Original-URL=/admin] https://site.com/admin -> 200
[%2f/] https://site.com/admin%2f/ -> 200
```


## Concurrency

You can set the concurrency value with the `-c` flag:

```
$ cat urls | snowBypass -c 35
```

## Timeout

You can set the timeout by using the `-t`:

```
$ cat urls | snowBypass -t 3
```
