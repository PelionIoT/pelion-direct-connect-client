# Access Control with JWT Token

## Overview

Clients using the api endpoint for tunnel `/v3/devices/{id}/services/{address}/connection` can use a jwt token to restrict access to a cloud application though the edge-proxy.

Creation of the jwt is up to the client but a public cetificate that signed the jwt needs to be uploaded into the Pelion Device Management. More information on Pelion Device Management JWT keys can be found [here](https://developer.pelion.com/docs/device-management/current/user-account/jwt-keys.html).

### Example Flow to Access Edge-Proxy Tunnel with JWT

- Create a private key-public key certificate pair
`$ openssl req -newkey rsa:2048 -nodes -keyout private.key -x509 -days 365 -out public.crt`

- Upload the public certificate to the Pelion Device Management verification keys api with an admin access key created from the Pelion Device Management.

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
- Create a jwt in RSASHA256 format and signed it with the private key and certificate, can be done in jwt.io. The supported signing algorithms are RS256, RS384, RS512, ES256, ES384, ES512. HS256 is not supported. Make sure that the exp clause/claim exists and valid.

- Access /v3/devices/{id}/services/{address}/connection with a header "X-Application-ID" corresponding to the same `{application-id}` (where the verification keys are created) and set `Authorization` header with value `Bearer {JWT}` . Using jwt as bearer token without this header will cause the client not be authenticated . `{id}` is the device id or gateway id and `{address}` is the `tunnel_ip:tunnel_port` format.

### JWT Claims to Access Edge-Proxy Tunnel
#### Required claims
- `exp`: JWT expiration date in number of seconds since Epoch format.
##### Optional claims. 
Jwt token can specify one or all of the claims below to restrict access to the application
- `pelion.edge.tunnel.device_id`:  The only device id that could be used to open tunnel, if specified.
- `pelion.edge.tunnel.ip`: The only IP that could be connected over tunnel, if specified.
- `pelion.edge.tunnel.port`:  The only port that could be connected over tunnel, if specified.
