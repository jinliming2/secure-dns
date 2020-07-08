# Secure DNS

## Table of Content

## Config

### Example

### Basic config

| Key | Type | Required | Default | Description |
|:---|:---:|:---:|:---:|:---|
| listen | `string[]` | ✔️ | | host and port to listen |
| timeout | `uint` | | `0` | timeout for each DNS request, default 0 to disable |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| user_agent | `string` | | `'secure-dns/VERSION https://github.com/jinliming2/secure-dns'` | User-Agent field for DNS over HTTPS |
| no_user_agent | `boolean` | | `false` | do not send User-Agent header in DNS over HTTPS |

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
| bootstrap | `boolean` | | `false` | this is a bootstrap DNS server, used to resolve names for DNS over HTTPS and DNS over TLS |
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |

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
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |

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
| domain | `string[]` | | | mark this DNS server only used to resolve specified domain names |
| suffix | `string[]` | | | mark this DNS server only used to resolve domain names with specified suffix |
| custom_ecs | `string[]` | | | custom EDNS Subnet to override |
| no_ecs | `boolean` | | `false` | disable EDNS Subnet and remove EDNS Subnet from DNS request |
| user_agent | `string` | | `'secure-dns/VERSION https://github.com/jinliming2/secure-dns'` | User-Agent field for DNS over HTTPS |
| no_user_agent | `boolean` | | `false` | do not send User-Agent header in DNS over HTTPS |

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
