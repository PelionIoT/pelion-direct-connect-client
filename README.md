# edge-proxy-local-client
edge-proxy-local-client serves as a local tcp server that accepts local tcp connection and establish corresponding websocket connection to edge tunneling service per tcp connection on the Pelion cloud

### Example Use:

* Deploy container to a node

* Install Go1.12. See the [instructions](https://golang.org/doc/install) here and build the client by running:
> `$ go build`

* Launch edge-proxy-local-client locally by providing parameters - `listen-uri`, `cloud-uri` and `api-key`
> `$ ./edge-proxy-local-client -listen-uri=<LOCAL_ADDRESS> -cloud-uri=<CLOUD_URI> -api-key=<API_KEY>`

* Open a brower by pointing the address with the above `listen-uri`
