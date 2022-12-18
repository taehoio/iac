package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newPulumiCICDServiceAccount(ctx *pulumi.Context, project *organizations.Project) (*serviceaccount.Account, error) {
	sa, err := serviceaccount.NewAccount(ctx, "pulumi-cicd", &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String("pulumi-cicd"),
		DisplayName: pulumi.String("pulumi-cicd"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = projects.NewIAMBinding(ctx, "pulumi-cicd-owner", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.String("serviceAccount:pulumi-cicd@taehoio-global.iam.gserviceaccount.com"),
		},
		Role: pulumi.String("roles/owner"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return sa, nil
}

func newDockerRegistryServiceAccount(ctx *pulumi.Context, project *organizations.Project) (*serviceaccount.Account, error) {
	sa, err := serviceaccount.NewAccount(ctx, "docker-registry", &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String("docker-registry"),
		DisplayName: pulumi.String("docker-registry"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = projects.NewIAMBinding(ctx, "docker-registry-repo-admin", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.String("serviceAccount:docker-registry@taehoio-global.iam.gserviceaccount.com"),
		},
		Role: pulumi.String("roles/artifactregistry.repoAdmin"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return sa, nil
}
