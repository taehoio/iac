package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

const (
	registryBasePath = "asia-northeast1-docker.pkg.dev/taehoio-global/docker-registry/"
)

func runCloudRunServices(ctx *pulumi.Context, project *organizations.Project) error {
	if _, err := newNotionproxyCloudRunService(ctx, project); err != nil {
		return err
	}

	return nil
}

func newNotionproxyCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "notionproxy"

	notionproxyCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Location:                 pulumi.String(iac.TokyoLocation),
		Project:                  project.ProjectId,
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(false),
		Template: cloudrun.ServiceTemplateArgs{
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.String(registryBasePath + serviceName),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								ContainerPort: pulumi.Int(3000),
							},
						},
						Resources: cloudrun.ServiceTemplateSpecContainerResourcesArgs{
							Limits: pulumi.StringMap{
								"cpu":    pulumi.String("2000m"),
								"memory": pulumi.String("4096Mi"),
							},
						},
					},
				},
				ServiceAccountName: pulumi.String("879516865171-compute@developer.gserviceaccount.com"),
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

	_, err = cloudrun.NewDomainMapping(ctx, "staging-taehoio", &cloudrun.DomainMappingArgs{
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
		Name:     pulumi.String("staging.taeho.io"),
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
