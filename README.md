# Secure DNS

[![Build Status](https://travis-ci.org/jinliming2/secure-dns.svg?branch=master)](https://travis-ci.org/jinliming2/secure-dns)
[![Go Report Card](https://goreportcard.com/badge/github.com/jinliming2/secure-dns)](https://goreportcard.com/report/github.com/jinliming2/secure-dns)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjinliming2%2Fsecure-dns.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjinliming2%2Fsecure-dns?ref=badge_shield)

A DNS client which implemented DoT and DoH, with load balancing, DNS cache, custom ECS and HOSTs.

## Table of Content

* [Table of Content](#table-of-content)
* [Config](#config)
   * [Example](#example)
   * [Basic config](#basic-config)
   * [Upstream DNS](#upstream-dns)
      * [Traditional DNS](#traditional-dns)
      * [DNS over TLS (DoT)](#dns-over-tls-dot)
      * [DNS over HTTPS (DoH)](#dns-over-https-doh)
   * [Custom Hosts](#custom-hosts)

## Config

### Example

```toml
[config]
listen = [
  '[::1]:53',
  '127.0.0.1:53',
]
custom_ecs = [
  '10.20.30.40',
  '50.60.70.80',
]
round_robin = 'swrr'

[[traditional]]
host = ['8.8.4.4', '8.8.8.8']
bootstrap = true

[[https]]
host = ['dns.google']
path = '/resolve'
google = true

[[https]]
host = ['dns.google']
weight = 10

[[https]]
host = ['1.1.1.1', '1.0.0.1']

[[tls]]
host = ['dns.google']

[[traditional]]
# Resolve private domain name using local DNS server
host = ['10.0.0.1']
suffix = [ 'private.network.org' ]

[hosts.'example.com']
A = [ '127.0.0.1' ]
AAAA = [ '::1' ]
TXT = [ 'text' ]
```

### Basic config

| Key | Type | Required | Default | Description |
|:---|:---:|:---:|:---:|:---|
| listen | `string[]` | ✔️ | | host and port to listen |
| timeout | `uint` | | `5` | timeout in seconds for each DNS request, 0 to disable |
| round_robin | `string` | | `'clock'` | upstream select round robin, can only be `'clock'`, `'random'`, `'wrandom'` or `'swrr'` |
| no_cache | `boolean` | | `false` | disable global DNS result cache |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| user_agent | `string` | | `'secure-dns/VERSION https://github.com/jinliming2/secure-dns'` | User-Agent field for DNS over HTTPS |
| no_user_agent | `boolean` | | `false` | do not send User-Agent header in DNS over HTTPS |
| no_single_inflight | `boolean` | | `false` | do not suppress multiple same outstanding queries |

Example:
```toml
[config]
listen = ['[::1]:53', '127.0.0.1:53']
custom_ecs = ['1.2.3.4', '1:2::3:4']
no_user_agent = true
```

### Upstream DNS

#### Traditional DNS

| Key | Type | Required | Default | Description |
|:---|:---:|:---:|:---:|:---|
| host | `string[]` | ✔️ | | ip addresses |
| port | `uint16` | | `53` | port to use |
| bootstrap | `boolean` | | `false` | mark this is a bootstrap DNS server, only used to resolve names for DNS over HTTPS and DNS over TLS |
| weight | `uint` | | `1` | weight used for weighted round robin, should > 0 |
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| no_single_inflight | `boolean` | | `false` | do not suppress multiple same outstanding queries |

Example:
```toml
[[traditional]]
host = ['8.8.4.4', '8.8.8.8']

[[traditional]]
host = ['1.1.1.1', '1.0.0.1']
bootstrap = true

[[traditional]]
host = ['9.9.9.9']
suffix = [
  'example.com',
]
```

#### DNS over TLS (DoT)

| Key | Type | Required | Default | Description |
|:---|:---:|:---:|:---:|:---|
| host | `string[]` | ✔️ | | ip addresses or host names |
| port | `uint16` | | `853` | port to use |
| hostname | `string` | | | hostname for ip addresses |
| weight | `uint` | | `1` | weight used for weighted round robin, should > 0 |
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| no_single_inflight | `boolean` | | `false` | do not suppress multiple same outstanding queries |

Example:
```toml
[[tls]]
host = ['dns.google']

[[tls]]
host = ['1.1.1.1']
hostname = 'cloudflare-dns.com'
domain = [
  'example.com',
]
```

> Note: If you want to specify hostname in host field, you must specify a traditional DNS server that marked with `bootstrap = true`.

#### DNS over HTTPS (DoH)

| Key | Type | Required | Default | Description |
|:---|:---:|:---:|:---:|:---|
| host | `string[]` | ✔️ | | ip addresses or host names |
| port | `uint16` | | `443` | port to use |
| hostname | `string` | | | hostname for ip addresses |
| path | `string` | | `'/dns-query'` | HTTP URI path to use |
| google | `boolean` | | `false` | use google's DoH query structure |
| cookie | `boolean` | | `false` | enable cookie support for this server |
| weight | `uint` | | `1` | weight used for weighted round robin, should > 0 |
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| user_agent | `string` | | `'secure-dns/VERSION https://github.com/jinliming2/secure-dns'` | User-Agent field for DNS over HTTPS |
| no_user_agent | `boolean` | | `false` | do not send User-Agent header in DNS over HTTPS |
| no_single_inflight | `boolean` | | `false` | do not suppress multiple same outstanding queries |

Example:
```toml
[[https]]
host = ['dns.google']

[[https]]
host = ['8.8.4.4', '8.8.8.8']
hostname = 'dns.google'
path = '/resolve'
google = true

[[https]]
host = ['1.1.1.1']
hostname = 'cloudflare-dns.com'
domain = [
  'example.com',
]
```

> Note: If you want to specify hostname in host field, you must specify a traditional DNS server that marked with `bootstrap = true`.

### Custom Hosts

Example:
```toml
[hosts.'example.com']
A = [
  '127.0.0.1',
  '192.168.1.1',
]
AAAA = ['::1']
TXT = ['this matches example.com']

[hosts.'*.example.com']
A = ['0.0.0.0']
TXT = ['this matches example.com a.example.com a.b.example.com a.b.c.example.com ...']
```


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjinliming2%2Fsecure-dns.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjinliming2%2Fsecure-dns?ref=badge_large)