# myresolver
Detect who resolved your DNS query.

## Try it out

### CLI interface
Just query the `get.my-resolver.834834.xyz` domain, `dig get.my-resolver.834834.xyz TXT +short`.

```bash
$ dig @1.1.1.1 get.my-resolver.834834.xyz TXT +short
"Query resolved by: '162.158.101.80'"
"ASN 13335: 'CLOUDFLARENET'
```
```bash
$ dig @8.8.8.8 get.my-resolver.834834.xyz TXT +short
"Query resolved by: '2a00:1450:4025:800::102'"
"ASN 15169: 'GOOGLE'"
```

### Web interface
https://my-resolver.834834.xyz/
