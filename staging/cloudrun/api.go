package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/secretmanager"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
)

func newApiCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "api"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName + "-sa"), // gcp service account length should be between 6 and 30
		DisplayName: pulumi.String(serviceName + "-sa"),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	oneononeMysqlPasswordSecret, err := secretmanager.NewSecret(ctx, serviceName+"-secret-oneonone-mysql-password", &secretmanager.SecretArgs{
		Project:  project.ProjectId,
		SecretId: pulumi.String(serviceName + "-secret-oneonone-mysql-password"),
		Replication: &secretmanager.SecretReplicationArgs{
			Automatic: pulumi.Bool(true),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = secretmanager.NewSecretIamMember(ctx, serviceName+"-secret-access-mysql-password", &secretmanager.SecretIamMemberArgs{
		Project:  project.ProjectId,
		SecretId: oneononeMysqlPasswordSecret.ID(),
		Role:     pulumi.String("roles/secretmanager.secretAccessor"),
		Member:   pulumi.Sprintf("serviceAccount:%s", sa.Email),
	}, pulumi.DependsOn([]pulumi.Resource{
		oneononeMysqlPasswordSecret,
	}))
	if err != nil {
		return nil, err
	}

	imageTag := "8c5fac8be4baf0315312d274253a11c945f86a9c"

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
								Name:          pulumi.String("http1"),
								ContainerPort: pulumi.Int(18082),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ENV"),
								Value: pulumi.String("staging"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_MYSQL_NETWORK_TYPE"),
								Value: pulumi.String("unix"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_MYSQL_ADDRESS"),
								Value: pulumi.String("/cloudsql/taehoio-staging:asia-northeast1:taehoio-shared-mysql"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_MYSQL_USER"),
								Value: pulumi.String("oneonone_sa"),
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name: pulumi.String("ONEONONE_MYSQL_PASSWORD"),
								ValueFrom: &cloudrun.ServiceTemplateSpecContainerEnvValueFromArgs{
									SecretKeyRef: &cloudrun.ServiceTemplateSpecContainerEnvValueFromSecretKeyRefArgs{
										Name: oneononeMysqlPasswordSecret.SecretId,
										Key:  pulumi.String("1"),
									},
								},
							},
							cloudrun.ServiceTemplateSpecContainerEnvArgs{
								Name:  pulumi.String("ONEONONE_MYSQL_DATABASE_NAME"),
								Value: pulumi.String("oneonone"),
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

	_, err = cloudrun.NewDomainMapping(ctx, "api-staging-taehoio", &cloudrun.DomainMappingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Name:     pulumi.String("api-staging.taeho.io"),
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
