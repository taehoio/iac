package iac

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	OrgId            = "1052569693362"
	BillingAccount   = "01FD1D-5CA95C-CA4195"
	TokyoLocation    = "asia-northeast1"
	GlobalProjectId  = "taehoio-global"
	StagingProjectId = "taehoio-staging"
)

func NewProject(ctx *pulumi.Context, projectId, projectName string) (*organizations.Project, error) {
	return organizations.NewProject(ctx, projectName, &organizations.ProjectArgs{
		AutoCreateNetwork: pulumi.Bool(true),
		BillingAccount:    pulumi.String(BillingAccount),
		OrgId:             pulumi.String(OrgId),
		ProjectId:         pulumi.String(projectId),
	})
}
