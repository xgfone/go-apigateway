## An Example Of The API Gateway Binary.

### Build
```bash
$ go build
```


### Help
```bash
$ apigateway -h
Usage of apigateway:
  -gatewayaddr string
        The address used by gateway. (default ":80")
  -log.level string
        The log level, such as debug, info, warn, error. (default "info")
  -manageraddr string
        The address used by manager. If set, start it.
  -provider string
        The provider of the dynamic configurations. (default "localfiledir")
  -provider.localfiledir.httpmiddlewaregroups string
        The directory of the local files storing the http middleware groups. (default "httpmiddlewaregroups")
  -provider.localfiledir.httproutes string
        The directory of the local files storing the http routes. (default "httproutes")
  -provider.localfiledir.upstreams string
        The directory of the local files storing the upstreams. (default "upstreams")
  -provider.localfiledir.interval duration
        The interval duration to check and reload the configurations. (default "1m")
  -tls.keyfile string
        The path of the certificate key file. If set, start api gateway with HTTPS.
  -tls.certfile string
        The path of the certificate file. If set, start api gateway with HTTPS.
  -tls.jsonfile string
        If set, add all the certificates in the file to the server.
```
