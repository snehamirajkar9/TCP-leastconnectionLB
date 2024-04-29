Least Connection Load Balancer
Features
Ths is a simple TCP load balancer that distributes incoming requests across multiple servers.

it uses least connection algorithm to route requests evenly accross upstream servers
it limits the rate of connections each client can do
it provides mTLS communication
it has an authorization scheme to authenticate clients

The design is present https://github.com/snehamirajkar9/TCP-leastconnectionLB/blob/Design_Document/DesignDoc-Least-ConnectionsTCPLB
