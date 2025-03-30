locals {
  container_host = "ssh://user@localhost/run/podman/podman.sock#pubkey=ssh-ed25519+AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh"
}

resource "podman_image" "postgres_17" {
  container_host = local.container_host
  preserve       = true
  reference      = "docker.io/library/postgres:17.5@sha256:6efd0df010dc3cb40d5e33e3ef84acecc5e73161bd3df06029ee8698e5e12c60"
  policy         = "missing"
}
