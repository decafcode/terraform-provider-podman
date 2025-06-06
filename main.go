// Copyright (c) HashiCorp, Inc.
// Copyright (c) Decaf Code
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/decafcode/terraform-provider-podman/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/decafcode/podman",
		Debug:   debug,
	}

	env := provider.PodmanProviderEnv{
		ContainerHost: os.Getenv("CONTAINER_HOST"),
		SshAuthSock:   os.Getenv("SSH_AUTH_SOCK"),
	}

	err := providerserver.Serve(context.Background(), provider.New(version, &env), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
