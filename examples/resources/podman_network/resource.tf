locals {
  container_host = "ssh://user@localhost/run/podman/podman.sock#pubkey=ssh-ed25519+AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh"
}

# Create two Podman networks to host various containerized services. The
# network called "external" can initiate connections to the outside world, and
# the network called "internal" cannot. Note that containers can be associated
# with multiple networks. For example, you might have some backend services in
# the "internal" network, and you might also have an HTTPS proxy that needs to
# forward connections to those services (and is therefore associated with the
# "internal" network), but it also needs to periodically talk to the public
# LetsEncrypt ACME server as well in order to obtain HTTPS certificates (and
# therefore it also needs to be a member of the "external" network).

resource "podman_network" "external" {
  container_host = local.container_host
  dns_enabled    = true
  name           = "external"
}

resource "podman_network" "internal" {
  container_host = local.container_host
  dns_enabled    = true
  internal       = true
  name           = "internal"
}
