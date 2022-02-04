package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/secretmanager"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newUserCloudRunService(
	ctx *pulumi.Context,
	project *organizations.Project,
	invokingServices []*cloudrun.Service,
) (*cloudrun.Service, error) {
	serviceName := "user"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName + "-sa"), // gcp service account length should be between 6 and 30
		DisplayName: pulumi.String(serviceName + "-sa"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	secret, err := secretmanager.NewSecret(ctx, serviceName+"-secret-mysql-password", &secretmanager.SecretArgs{
		Project:  project.ProjectId,
		SecretId: pulumi.String(serviceName + "-secret-mysql-password"),
		Replication: &secretmanager.SecretReplicationArgs{
			Automatic: pulumi.Bool(true),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = secretmanager.NewSecretIamMember(ctx, serviceName+"-secret-access-mysql-password", &secretmanager.SecretIamMemberArgs{
		Project:  project.ProjectId,
		SecretId: secret.ID(),
		Role:     pulumi.String("roles/secretmanager.secretAccessor"),
		Member:   pulumi.Sprintf("serviceAccount:%s", sa.Email),
	}, pulumi.DependsOn([]pulumi.Resource{
		secret,
	}))
	if err != nil {
		return nil, err
	}

	imageTag := "99c910d64f02c19b521aa106728a22a8074ca2d2"

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
					"run.googleapis.com/cloudsql-instances":    "taehoio-staging:asia-northeast1:taehoio-shared-mysql",
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
								ContainerPort: pulumi.Int(18081),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ENV"),
								Value: pulumi.String("staging"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("MYSQL_NETWORK_TYPE"),
								Value: pulumi.String("unix"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("MYSQL_ADDRESS"),
								Value: pulumi.String("/cloudsql/taehoio-staging:asia-northeast1:taehoio-shared-mysql"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("MYSQL_USER"),
								Value: pulumi.String("taehoio_sa"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name: pulumi.String("MYSQL_PASSWORD"),
								ValueFrom: &cloudrun.ServiceTemplateSpecContainerEnvValueFromArgs{
									SecretKeyRef: &cloudrun.ServiceTemplateSpecContainerEnvValueFromSecretKeyRefArgs{
										Name: secret.SecretId,
										Key:  pulumi.String("3"),
									},
								},
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("MYSQL_DATABASE_NAME"),
								Value: pulumi.String("taehoio"),
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
