package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newDockerRegistry(ctx *pulumi.Context) (*artifactregistry.Repository, error) {
	return artifactregistry.NewRepository(ctx, "docker-registry", &artifactregistry.RepositoryArgs{
		Format:       pulumi.String("DOCKER"),
		RepositoryId: pulumi.String("docker-registry"),
	}, pulumi.Protect(false))
}
