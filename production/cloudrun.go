package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

func runCloudRunServices(ctx *pulumi.Context, project *organizations.Project) error {
	if _, err := newNotionproxyCloudRunService(ctx, project); err != nil {
		return err
	}

	return nil
}

func newNotionproxyCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "notionproxy"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName),
		DisplayName: pulumi.String(serviceName),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	notionproxyCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Location:                 pulumi.String(iac.TokyoLocation),
		Project:                  project.ProjectId,
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Template: cloudrun.ServiceTemplateArgs{
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.String(iac.DockerRegistryBasePath + serviceName + "@sha256:0ef41356e21c272066057d335bd48a79131b8a4e6e4eace137f6bcdcdb90b4f5"),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								ContainerPort: pulumi.Int(3000),
							},
						},
						Resources: cloudrun.ServiceTemplateSpecContainerResourcesArgs{
							Limits: pulumi.StringMap{
								"cpu":    pulumi.String("1000m"),
								"memory": pulumi.String("512Mi"),
							},
						},
					},
				},
				ServiceAccountName: sa.Email,
				TimeoutSeconds:     pulumi.Int(300),
			},
		},
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	if _, err := cloudrun.NewIamMember(ctx, serviceName+"-everyone", &cloudrun.IamMemberArgs{
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
		Service:  notionproxyCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	_, err = cloudrun.NewDomainMapping(ctx, "taehoio", &cloudrun.DomainMappingArgs{
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
		Name:     pulumi.String("taeho.io"),
		Metadata: cloudrun.DomainMappingMetadataArgs{
			Namespace: project.ProjectId,
		},
		Spec: cloudrun.DomainMappingSpecArgs{
			RouteName:       notionproxyCloudRunService.Name,
			CertificateMode: pulumi.String("AUTOMATIC"),
		},
	})
	if err != nil {
		return nil, err
	}

	return notionproxyCloudRunService, nil
}
