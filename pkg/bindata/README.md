The files in this folder are auto generated from the files/ and templates/ folders.

- To generate the binary base64 data you need to install `go-bindata`:

```shell
curl --silent --location --output /usr/local/bin/go-bindata https://github.com/kevinburke/go-bindata/releases/download/v3.22.0/go-bindata-linux-amd64
chmod 755 /usr/local/bin/go-bindata
make generate
```

See: https://github.com/kevinburke/go-bindata

- **Do not put anything except Go HTML template files in the templates/ directory.**