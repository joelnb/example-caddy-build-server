# Example Caddy Build Server

## Motivation

Currently the [caddy-ansible role](https://github.com/caddy-ansible/caddy-ansible) defaults to downloading builds from github. If the user requests plugins however these are not available on github & the role defaults to using the caddy download page to perform the download. This works fine but the service is provided for free by the caddy project maintainers & should not be relied on, especially by people managing many machines.

This repo aims to show how a simple server can be setup which supports the same download endpoint as the caddy download page. This can then be used with the ansible role to provide a location controlled by the user from which to download.

## Usage

I wouldn't really recommend using this code unless you have read and understand it fully. I definitely wouldn't call it 'production-ready'. I created it for my own simple usage & to show how simply a compatible server can be created.

If you have read that and still want to give it a try you can run a local instance with:

```bash
docker run -p 127.0.0.1:8080:8081 --rm -it joelnb/example-caddy-build-server
```

And then download the caddy binary with this example `curl` command:

```bash
curl -v "localhost:8080/api/download?os=linux&arch=amd64&p=github.com/caddy-dns/lego-deprecated" --output caddy
```

If you wanted to share this between multiple machines then adding a reverse proxy with TLS termination would be recommended.

