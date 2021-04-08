## TWINS Demo Reset

This Go package allows resetting the state of the TWINS demo over HTTP. The
following actions are taken during the reset:

1. All the TWINS related services are stopped. See `pkg/usr/lib/systemd/user/twins.target`.
2. The researcher, biobank and mediator databases are reset to a snapshot copy.
3. All the TWINS related services are started. See `pkg/usr/lib/systemd/user/twins.target`.
4. The DARCs are reset to allow permissions only to the initial owner.

### Compilation

You may compile the binary by executing `go build` in the root of the repository.

### Usage

You may start the server by executing `./twins-demo-reset serve`. The server
runs at port `9999` by default. Please refer to `./twins-demo-reset serve -h`
for more options.

### Deployment

The HTTP server is run as a non-root user and listens for connections locally.
The endpoint is exposed to the outside world via an nginx reverse proxy which
uses HTTP Basic Auth for authentication.

Please look at the handler for `/reset-demo` in `/etc/nginx/sites-enabled/wp.conf`
for more information.

The server is managed by systemd user instance. Please have a look at
`pkg/usr/lib/systemd/user/twins_demo_reset.service` in the repository root for more information.

All systemd units are installed at `/usr/lib/systemd/user` while installing the
debian archive.
