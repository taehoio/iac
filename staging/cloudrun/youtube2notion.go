package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newYoutube2notionCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "youtube2notion"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName),
		DisplayName: pulumi.String(serviceName),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	imageTag := "33a2b8326b95b42b4bd09258792f3c753dbd9a8d"

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
					"autoscaling.knative.dev/maxScale": "100",
				}),
			},
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(1),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", iac.DockerRegistryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								Name:          pulumi.String("http1"),
								ContainerPort: pulumi.Int(5000),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_PROFILE"),
								Value: pulumi.String("true"),
							},
						},
						Resources: cloudrun.ServiceTemplateSpecContainerResourcesArgs{
							Limits: pulumi.StringMap{
								"cpu":    pulumi.String("4000m"),
								"memory": pulumi.String("8192Mi"),
							},
						},
						Args: pulumi.StringArray{
							pulumi.String("app.py"),
						},
					},
				},
				ServiceAccountName: sa.Email,
				TimeoutSeconds:     pulumi.Int(60 * 60),
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

	return service, nil
}
