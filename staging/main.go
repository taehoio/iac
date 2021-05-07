package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		project, err := iac.NewProject(ctx, iac.StagingProjectId, iac.StagingProjectId)
		if err != nil {
			return err
		}

		helloCloudRunService, err := cloudrun.NewService(ctx, "hello", &cloudrun.ServiceArgs{
			Location:                 pulumi.String(iac.TokyoLocation),
			Project:                  project.ProjectId,
			Name:                     pulumi.String("hello"),
			AutogenerateRevisionName: pulumi.Bool(false),
			Template: cloudrun.ServiceTemplateArgs{
				Spec: cloudrun.ServiceTemplateSpecArgs{
					ContainerConcurrency: pulumi.Int(80),
					Containers: cloudrun.ServiceTemplateSpecContainerArray{
						cloudrun.ServiceTemplateSpecContainerArgs{
							Image: pulumi.String("us-docker.pkg.dev/cloudrun/container/hello"),
							Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
								cloudrun.ServiceTemplateSpecContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
								},
							},
							Resources: cloudrun.ServiceTemplateSpecContainerResourcesArgs{
								Limits: pulumi.StringMap{
									"cpu":    pulumi.String("1000m"),
									"memory": pulumi.String("256Mi"),
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
			return err
		}

		_, err = cloudrun.NewDomainMapping(ctx, "hello-staging-taehoio", &cloudrun.DomainMappingArgs{
			Location: pulumi.String(iac.TokyoLocation),
			Project:  project.ProjectId,
			Name:     pulumi.String("hello.staging.taeho.io"),
			Metadata: cloudrun.DomainMappingMetadataArgs{
				Namespace: project.ProjectId,
			},
			Spec: cloudrun.DomainMappingSpecArgs{
				RouteName:       helloCloudRunService.Name,
				CertificateMode: pulumi.String("AUTOMATIC"),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
