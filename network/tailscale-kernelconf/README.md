# Tailscale-KernelConf

Parses kernel arguments from `/proc/cmdline` to configure the tailscale extension via `/var/etc/tailscale/auth.env`

## Installation

Simplest install
```
machine:
  install:
    extensions:
      - image: ghcr.io/siderolabs/tailscale:1.44.0
      - image: ghcr.io/jtcressy-home/tailscale-kernelconf:latest
    extraKernelArgs:
      - tailscale.authkey=tskey-auth-abcd123CNTRL-abcd12345
```

```
> talosctl apply -n node myconfig.yaml
> talosctl upgrade -n node
```

Injection via custom ISO's built with siderolabs/imager
```
docker run --rm -t -v $PWD/_out:/out ghcr.io/siderolabs/imager:v1.5.3 iso \
  --arch amd64 \
  --system-extension-image ghcr.io/siderolabs/tailscale:1.44.0 \
  --system-extension-image ghcr.io/jtcressy-home/tailscale-kernelconf:latest \
  --extra-kernel-arg tailscale.authkey=tskey-auth-abcd123CNTRL-abcd12345
```
Boot your machine with the iso saved to ./_out and it should join your tailnet before even bootstrapping your node with talosctl

## Configuration

Current supported kernel arguments and how they map to environment variables in `/var/etc/tailscale/auth.env`:

- `tailscale.authkey` TS_AUTHKEY: the authkey to use for login.
- `tailscale.hostname` TS_HOSTNAME: the hostname to request for the node.
- `tailscale.accept-dns` TS_ACCEPT_DNS: whether to use the tailnet's DNS configuration. (default false)
- `tailscale.authonce` TS_AUTH_ONCE: if true, only attempt to log in if not already logged in. If false (the default, for backwards compatibility), forcibly log in every time the container starts.

