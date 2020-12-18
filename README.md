# Diagnostic logging

[Diagnostic logging](https://tersesystems.com/blog/2019/10/05/diagnostic-logging-citations-and-sources/)
is a developer-oriented style of logging to troubleshoot programs running in production.
Verbose logging is usually enabled on demand with feature flags.
In this example `logging` HTTP endpoint is used instead.

```sh
$ go run ./cmd/server/
$ curl localhost:8000\?user_id=111
{"component":"api","level":"error","path":"/","ts":"2020-12-18T20:34:27.634519Z","user_id":"111"}

$ curl localhost:9000/logging -d 'logger=debug'
debug logger enabled

$ curl localhost:8000\?user_id=111
{"component":"api","level":"debug","path":"/","ts":"2020-12-18T20:34:55.456714Z","user_id":"111"}
{"component":"api","level":"info","path":"/","ts":"2020-12-18T20:34:55.456798Z","user_id":"111"}
{"component":"api","level":"warn","path":"/","ts":"2020-12-18T20:34:55.456803Z","user_id":"111"}
{"component":"api","level":"error","path":"/","ts":"2020-12-18T20:34:55.456807Z","user_id":"111"}
```

It's also possible to filter debug level logs by a key-value pair.
Let's say we want to see verbose logs for `user_id=555`.

```sh
$ curl localhost:9000/logging -d 'logger=debug&key=user_id&value=555'
filtered debug logger enabled: user_id=555

$ curl localhost:8000\?user_id=111
{"component":"api","level":"error","path":"/","ts":"2020-12-18T20:35:43.186666Z","user_id":"111"}

$ curl localhost:8000\?user_id=555
{"component":"api","level":"debug","path":"/","ts":"2020-12-18T20:35:53.187252Z","user_id":"555"}
{"component":"api","level":"info","path":"/","ts":"2020-12-18T20:35:53.187296Z","user_id":"555"}
{"component":"api","level":"warn","path":"/","ts":"2020-12-18T20:35:53.187302Z","user_id":"555"}
{"component":"api","level":"error","path":"/","ts":"2020-12-18T20:35:53.187307Z","user_id":"555"}
```

Once logs are collected, the standard logger should be enabled back.

```sh
$ curl localhost:9000/logging -d 'logger=standard'
standard logger enabled

$ curl localhost:8000\?user_id=555
{"component":"api","level":"error","path":"/","ts":"2020-12-18T20:42:38.968994Z","user_id":"555"}
```
