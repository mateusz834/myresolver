# myresolver
Detect who resolved your DNS query.

## Try it out
Just query the `get.my-resolver.834834.xyz` domain.

### CLI interface

```bash
$ dig get.my-resolver.834834.xyz +short TXT
"Query resolved by: '162.158.101.80'"
"ASN 13335: 'CLOUDFLARENET'"
```
```bash
$ dig get.my-resolver.834834.xyz +short TXT
"Query resolved by: '2a00:1450:4025:802::105'"
"ASN 15169: 'GOOGLE'"
```

### Web interface
https://my-resolver.834834.xyz/
