locals {
  container_host = "ssh://user@localhost/run/podman/podman.sock#pubkey=ssh-ed25519+AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh"
  uid_postgres   = 2001
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

resource "podman_network" "internal" {
  container_host = local.container_host
  dns_enabled    = true
  internal       = true
  name           = "internal"
}

resource "podman_image" "postgres_17" {
  container_host = local.container_host
  preserve       = true
  reference      = "docker.io/library/postgres:17.5@sha256:6efd0df010dc3cb40d5e33e3ef84acecc5e73161bd3df06029ee8698e5e12c60"
  policy         = "missing"
}

resource "podman_container" "postgres_17" {
  container_host = local.container_host
  image          = podman_image.postgres_17.id
  name           = "postgres-17"
  restart_policy = "always"

  env = {
    "POSTGRES_PASSWORD_FILE" = "/run/secrets/pgpassword"
  }

  mounts = [
    {
      options = ["U", "Z"]
      source  = "/srv/postgres-17"
      target  = "/var/lib/postgresql/data"
    }
  ]

  networks = [
    {
      id = podman_network.internal.id
    }
  ]

  secrets = [
    {
      gid    = local.uid_postgres
      path   = "/run/secrets/pgpassword"
      secret = podman_secret.pgpassword.name
      uid    = local.uid_postgres
    }
  ]

  user = {
    group = local.uid_postgres
    user  = local.uid_postgres
  }
}
