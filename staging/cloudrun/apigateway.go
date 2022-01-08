package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
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

	imageTag := "478cae66658f8d85206601bc6fa2d4122c6706fb"

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
								Name:  pulumi.String("SHOULD_USE_GRPC_CLIENT_TLS"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("CA_CERT_FILE"),
								Value: pulumi.String("/etc/ssl/certs/ca-certificates.crt"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_PROFILE"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("SHOULD_TRACE"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("IS_IN_GCP"),
								Value: pulumi.String("true"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ID_TOKEN"),
								Value: pulumi.String("NOT_USED_IN_GCP"),
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

	_, err = cloudrun.NewDomainMapping(ctx, "api-staging-taehoio", &cloudrun.DomainMappingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Name:     pulumi.String("api.staging.taeho.io"),
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

	_, err = cloudrun.NewDomainMapping(ctx, "api-taehoio", &cloudrun.DomainMappingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Name:     pulumi.String("api.taeho.io"),
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
