locals {
  container_host = "ssh://user@localhost/run/podman/podman.sock#pubkey=ssh-ed25519+AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh"
}

variable "pgpassword" {
  type = string
}

resource "podman_secret" "pgpassword" {
  container_host = local.container_host
  name           = "pgpassword"
  value          = var.pgpassword
  value_version  = 1
}
