# GoMercury
A simple Golang API that provides information related to IP addresses.

GoMercury lets you pull up whois information from a domain and searches MaxMinds GeoIP database from an IP address.

## Usage
Domain as input
```sh
curl us-central1-gomercury-356415.cloudfunctions.net/GoMercury/?domain=reddit.com
```
IP address as input
```sh
curl us-central1-gomercury-356415.cloudfunctions.net/GoMercury/?ipaddress=8.8.8.8
```
## Limitations
No subdomains in query
