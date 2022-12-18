package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newDockerRegistry(ctx *pulumi.Context, project *organizations.Project) (*artifactregistry.Repository, error) {
	repo, err := artifactregistry.NewRepository(ctx, "docker-registry", &artifactregistry.RepositoryArgs{
		Project:      project.ProjectId,
		Location:     pulumi.String(iac.TokyoLocation),
		Format:       pulumi.String("DOCKER"),
		RepositoryId: pulumi.String("docker-registry"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = projects.NewIAMBinding(ctx, "docker-registry-read", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.String("serviceAccount:service-879516865171@serverless-robot-prod.iam.gserviceaccount.com"),
			pulumi.String("serviceAccount:service-290041120046@serverless-robot-prod.iam.gserviceaccount.com"),
		},
		Role: pulumi.String("roles/artifactregistry.reader"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return repo, nil
}
