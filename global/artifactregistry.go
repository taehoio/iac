package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

func newDockerRegistry(ctx *pulumi.Context, project *organizations.Project) (*artifactregistry.Repository, error) {
	return artifactregistry.NewRepository(ctx, "docker-registry", &artifactregistry.RepositoryArgs{
		Project:      project.ProjectId,
		Location:     pulumi.String(iac.TokyoLocation),
		Format:       pulumi.String("DOCKER"),
		RepositoryId: pulumi.String("docker-registry"),
	}, pulumi.Protect(false))
}
