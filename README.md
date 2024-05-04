# Least Connection Load Balancer

## Features
Ths is a simple TCP load balancer that distributes incoming requests across multiple servers.
- it uses least connection algorithm to route requests evenly accross upstream servers
- it limits the rate of connections each client can do
- it provides mTLS communication
- it has an authorization scheme to authenticate clients
- it checks if upstreams servers are healthy
- it do not reuse connections

## Design
The design is available under the design folder above
## Requirements / dependencies
For the overall usage
- docker
- docker-compose
- make
- openssl
## Makefile
A Makefile helps to build and run the project, and execute some recurrent tasks.
Type `make help` to list all the targets.


## Balancer library
### API
The balancer library exposes the following API:
```go
// NewLoadBalancer initializes a balancer with the given upstreams
func NewLoadBalancer(upstreams []string) *Balancer
// Balance distributes a connection to an upstream server
func (b *Balancer) Balance(conn *net.TCPConn, clientId string, allowedUpstreams []string, errorsChan chan []error) error
```
The `Balance` function takes these parameters:
- `conn` is the incoming TCP connection from the client
- `clientId` is a unique ID for the client (should be the CN in the certificate)
- `allowedUpstreams` limits which upstreams can be selected (used by auth scheme)
- `errorsChan` is used to report any errors during forwarding

### Implementation
The balancer sequentiallly applies the following:
- calls the rate limiter and if the client has exceeded the rate limit, return an error and close the connection
- pick a least used upstream server from the allowed upstreams
- forward the connection to the selected upstream using a connection forwarder

### Testing
The `test` folder contains a `docker-compose.yml` file to spin up dummy servers for testing.

A bash script is provided to call parallelly a big number of requests.

## Server
The server is exposed on port 9443, which is hardcoded in `main.go`.
### Docker
There is a `Dockerfile`. It uses a multi-stage approach to build the binary and copy it into a scratch image.

Running `make docker-run` launches the docker image after building it.

A `docker-compose-yaml` is present at the root level which demonstrates the load balancing in docker containers.

Running `make docker-compose` builds and run the composition.
### MTLS
The server requires mTLS communication, so it needs certificates and keys.

The implementation takes shortcuts as to certificates locations, they are hardcoded to:
- CA Certificate: `certs/ca/ca.crt`
- Server cert file: `certs/server/server.crt`
- Server key file in PEM format:  `certs/server/server.key.pem`

#### Testing
Some scripts are provided to generate certificates and keys for testing. You will need to have `openssl` installed. They are located in the `certs` folder.
The `generate_certs.sh` script generates the CA certificate and the server certificate and key + 2 clients certificates. 

`generate_all_certs.sh` generates alternatives certificates to be able to test the rejection of the client certificate.

`test-mtls.sh` tests the call to the server with 2 different clients.

`test-mtls-alt.sh` tests the call with a certificate that does not use the trusted CA.

### Authotization scheme
The server supports a simple authorization scheme based on the client certificate CN.
The scheme is hardcoded in the `main.go` file for now.

