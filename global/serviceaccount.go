package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newPulumiCICDServiceAccount(ctx *pulumi.Context, project *organizations.Project) (*serviceaccount.Account, error) {
	return serviceaccount.NewAccount(ctx, "pulumi-cicd", &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String("pulumi-cicd"),
		DisplayName: pulumi.String("pulumi-cicd"),
	}, pulumi.Protect(false))
}

func newDockerRegistryServiceAccount(ctx *pulumi.Context, project *organizations.Project) (*serviceaccount.Account, error) {
	return serviceaccount.NewAccount(ctx, "docker-registry", &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String("docker-registry"),
		DisplayName: pulumi.String("docker-registry"),
	}, pulumi.Protect(false))
}
