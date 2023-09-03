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
        The address used by manager.
  -provider string
        The provider of the dynamic configurations. (default "localfiledir")
  -provider.localfiledir.middlewaregroups string
        The directory of the local files storing the dynamic middleware groups. (default "middlewaregroups")
  -provider.localfiledir.routes string
        The directory of the local files storing the dynamic routes. (default "routes")
  -provider.localfiledir.upstreams string
        The directory of the local files storing the dynamic upstreams. (default "upstreams")
  -tls.certfile string
        The path of the certificate file.
  -tls.jsonfile string
        If set, add all the certificates in the file to the server.
  -tls.keyfile string
        The path of the certificate key file.
```
