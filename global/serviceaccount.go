package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newPulumiCICDServiceAccount(ctx *pulumi.Context) (*serviceaccount.Account, error) {
	return serviceaccount.NewAccount(ctx, "pulumi-cicd", &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("pulumi-cicd"),
		DisplayName: pulumi.String("pulumi-cicd"),
	}, pulumi.Protect(false))
}
