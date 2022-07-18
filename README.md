# GoMercury
A simple Golang API that provides information related to IP addresses.

GoMercury lets you pull up whois information from a domain and searches MaxMinds GeoIP database from an IP address.

## Usage
Domain as input
```sh
curl --data "reddit.com" us-central1-gomercury-356415.cloudfunctions.net/GoMercury
```
IP address as input
```sh
curl --data "212.2.69.135" us-central1-gomercury-356415.cloudfunctions.net/GoMercury
```
## Limitations
No subdomains in query
