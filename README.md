# Example Caddy Build Server

Currently the [caddy-ansible role](https://github.com/caddy-ansible/caddy-ansible) defaults to downloading builds from github. If the user requests plugins however these are not available on github & the role defaults to using the caddy download page to perform the download. This works fine but this service is provided for free by the caddy project maintainers & should not be relied on.

This repo aims to show how a simple server can be setup which supports the same download endpoint as the caddy download page. This can then be used with the ansible role to provide a location controlled by the user from which to download.
