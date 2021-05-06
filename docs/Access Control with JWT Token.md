# Access Control with JWT Token

## Overview

Clients using the api endpoint for tunnel `/v3/devices/{id}/services/{address}/connection` needs a jwt token to be both authenticated and authorized.

Creation of the jwt is up to the client but a public cetificate that signed the jwt needs to be uploaded into the Pelion Device Management. An example flow is provided below.

### Example Flow to Access Edge-Proxy Tunnel with JWT

- Create a private key-public key certificate pair
`$ openssl req -newkey rsa:2048 -nodes -keyout private.key -x509 -days 365 -out public.crt`

- Upload the public certificate to the Pelion Device Management verification keys api with an admin access key created from the Pelion Device Management  (TODO: Paste a link on how to create admin access key).

```
curl -0 -v -X POST https://api.us-east-1.mbedcloud.com/v3/applications/{application-id}/verification-keys \
2-H 'Authorization: Bearer ak_2MD...' \
3-H 'content-type: application/json' \
4--data-binary @- << EOF
5{
6    "name": "JWT-test",
7    "certificate": "-----BEGIN CERTIFICATE-----
8    ...
9    -----END CERTIFICATE-----"
10}
11EOF
```
- Create a jwt in RSASHA256 format and signed it with the private key and certificate, can be done in jwt.io. The supported signing algorithms are RS256, RS384, RS512, ES256, ES384, ES512. HS256 is not supported. Make sure that the exp claus/claim exists and valid.

- Access /v3/devices/{id}/services/{address}/connection with a header "X-Application-ID" corresponding to the same `{application-id}` (where the verification keys are created) and set `Authorization` header with value `Bearer {JWT}` . Without this header and bearer token the client will not be authenticated and autorized. `{id}` is the device id or gateway id and `{address}` is the `tunnel_ip:tunnel_port` format.

### Required JWT Claims to Access Edge-Proxy Tunnel
- `exp`: JWT expiration date in number of seconds since Epoch format.
- `pelion.edge.tunnel.device_id`: The only device id that could be used to open tunnel.
- `pelion.edge.tunnel.ip`: The only IP that could be connected over tunnel.
- `pelion.edge.tunnel.port`: The only port that could be connected over tunnel.
