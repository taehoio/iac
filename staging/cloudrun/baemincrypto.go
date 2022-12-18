package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newBaemincryptoCloudRunService(
	ctx *pulumi.Context,
	project *organizations.Project,
	invokingServices []*cloudrun.Service,
) (*cloudrun.Service, error) {
	serviceName := "baemincrypto"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName),
		DisplayName: pulumi.String(serviceName),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	imageTag := "4b3049a6e84204d21a407a08481042c894b90134"

	service, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Project:                  project.ProjectId,
		Location:                 pulumi.String(iac.TokyoLocation),
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Metadata: cloudrun.ServiceMetadataArgs{
			Annotations: pulumi.ToStringMap(map[string]string{
				"run.googleapis.com/ingress": "all",
			}),
		},
		Template: cloudrun.ServiceTemplateArgs{
			Metadata: cloudrun.ServiceTemplateMetadataArgs{
				Annotations: pulumi.ToStringMap(map[string]string{
					"autoscaling.knative.dev/maxScale":         "100",
					"run.googleapis.com/execution-environment": "gen1",
				}),
			},
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", iac.DockerRegistryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								Name:          pulumi.String("h2c"),
								ContainerPort: pulumi.Int(50051),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ENV"),
								Value: pulumi.String("staging"),
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

	if _, err := cloudrun.NewIamBinding(ctx, serviceName+"-invoker", &cloudrun.IamBindingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Service:  service.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Members:  servicesToMembers(invokingServices),
	}); err != nil {
		return nil, err
	}

	return service, nil
}
