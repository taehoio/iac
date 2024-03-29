package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newApigatewayCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "apigateway"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName),
		DisplayName: pulumi.String(serviceName),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	imageTag := "2008b7dd6abebd28d82f2e93a5b6372e7ec1a8f5"

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
								Name:          pulumi.String("http1"),
								ContainerPort: pulumi.Int(8080),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ENV"),
								Value: pulumi.String("staging"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("BAEMINCRYPTO_GRPC_SERVICE_ENDPOINT"),
								Value: pulumi.String("baemincrypto-5hwa5dthla-an.a.run.app:443"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("BAEMINCRYPTO_GRPC_SERVICE_URL"),
								Value: pulumi.String("https://baemincrypto-5hwa5dthla-an.a.run.app"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("USER_GRPC_SERVICE_ENDPOINT"),
								Value: pulumi.String("user-5hwa5dthla-an.a.run.app:443"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("USER_GRPC_SERVICE_URL"),
								Value: pulumi.String("https://user-5hwa5dthla-an.a.run.app"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("AUTH_GRPC_SERVICE_ENDPOINT"),
								Value: pulumi.String("auth-5hwa5dthla-an.a.run.app:443"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("AUTH_GRPC_SERVICE_URL"),
								Value: pulumi.String("https://auth-5hwa5dthla-an.a.run.app"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_GRPC_SERVICE_ENDPOINT"),
								Value: pulumi.String("oneonone-5hwa5dthla-an.a.run.app:443"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_GRPC_SERVICE_URL"),
								Value: pulumi.String("https://oneonone-5hwa5dthla-an.a.run.app"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("CAR_GRPC_SERVICE_URL"),
								Value: pulumi.String("https://car-5hwa5dthla-an.a.run.app"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_USE_GRPC_CLIENT_TLS"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("CA_CERT_FILE"),
								Value: pulumi.String("/etc/ssl/certs/ca-certificates.crt"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("IS_IN_GCP"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_PROFILE"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_TRACE"),
								Value: pulumi.String("true"),
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

	return service, nil
}
