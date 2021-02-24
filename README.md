# pelion-direct-connect-client

pelion-direct-connect-client acts as a local TCP server that accepts local TCP connections and establishes corresponding websocket connections to the Pelion Edge tunneling service through a TCP connection on the Pelion cloud.

### Example Use:

1. Deploy a container to a node.
1. [Install Go1.15](https://golang.org/doc/install).
1. Build the client by running:

   ```
   $ go build
   ```

1. Launch pelion-direct-connect-client locally by providing parameters - `listen-uri`, `cloud-uri` and `api-key`:

   ```
   $ ./pelion-direct-connect-client -listen-uri=<LOCAL_ADDRESS> -cloud-uri=<CLOUD_URI> -api-key=<API_KEY>
   ```

1. Open a browser by pointing to the address with the above `listen-uri`.
