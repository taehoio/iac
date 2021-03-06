package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

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

	imageTag := "d712864dfe6ebe66838d9f93d9102fa6af323c2c"

	service, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Project:                  project.ProjectId,
		Location:                 pulumi.String(iac.TokyoLocation),
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Template: cloudrun.ServiceTemplateArgs{
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", iac.DockerRegistryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								Name:          pulumi.String("http1"),
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
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Service:  service.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	_, err = cloudrun.NewDomainMapping(ctx, "taehoio", &cloudrun.DomainMappingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Name:     pulumi.String("taeho.io"),
		Metadata: cloudrun.DomainMappingMetadataArgs{
			Namespace: project.ProjectId,
		},
		Spec: cloudrun.DomainMappingSpecArgs{
			RouteName:       service.Name,
			CertificateMode: pulumi.String("AUTOMATIC"),
		},
	})
	if err != nil {
		return nil, err
	}

	return service, nil
}
