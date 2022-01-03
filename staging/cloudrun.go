package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

const (
	registryBasePath = "asia-northeast1-docker.pkg.dev/taehoio-global/docker-registry/"
)

func runCloudRunServices(ctx *pulumi.Context, project *organizations.Project) error {
	notionproxy, err := newNotionproxyCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	apigateway, err := newApigatewayCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	baemincrypto, err := newBaemincryptoCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	youtube2notion, err := newYoutube2notionCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	notionproxySA := stringOutPtrToStringOutput(notionproxy.Template.Spec().ServiceAccountName())
	apigatewaySA := stringOutPtrToStringOutput(apigateway.Template.Spec().ServiceAccountName())
	baemincryptoSA := stringOutPtrToStringOutput(baemincrypto.Template.Spec().ServiceAccountName())
	youtube2notionSA := stringOutPtrToStringOutput(youtube2notion.Template.Spec().ServiceAccountName())

	_, err = projects.NewIAMBinding(ctx, "service-profiler-agent", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.Sprintf("serviceAccount:%s", notionproxySA),
			pulumi.Sprintf("serviceAccount:%s", apigatewaySA),
			pulumi.Sprintf("serviceAccount:%s", baemincryptoSA),
			pulumi.Sprintf("serviceAccount:%s", youtube2notionSA),
		},
		Role: pulumi.String("roles/cloudprofiler.agent"),
	}, pulumi.Protect(false))
	if err != nil {
		return err
	}

	_, err = projects.NewIAMBinding(ctx, "service-trace-agent", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.Sprintf("serviceAccount:%s", notionproxySA),
			pulumi.Sprintf("serviceAccount:%s", apigatewaySA),
			pulumi.Sprintf("serviceAccount:%s", baemincryptoSA),
			pulumi.Sprintf("serviceAccount:%s", youtube2notionSA),
		},
		Role: pulumi.String("roles/cloudtrace.agent"),
	}, pulumi.Protect(false))
	if err != nil {
		return err
	}

	return nil
}

func stringOutPtrToStringOutput(spo pulumi.StringPtrOutput) pulumi.StringOutput {
	return spo.ApplyT(func(sp *string) string {
		return *sp
	}).(pulumi.StringOutput)
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

	imageTag := "99155bcd92ea40e40657ea290a7b174154dd72fa"

	notionproxyCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Project:                  project.ProjectId,
		Location:                 pulumi.String(iac.TokyoLocation),
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Template: cloudrun.ServiceTemplateArgs{
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", registryBasePath, serviceName, imageTag),
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
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
		Service:  notionproxyCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	_, err = cloudrun.NewDomainMapping(ctx, "staging-taehoio", &cloudrun.DomainMappingArgs{
		Project:  project.ProjectId,
		Location: pulumi.String(iac.TokyoLocation),
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

	return notionproxyCloudRunService, nil
}

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

	imageTag := "7973144f977996369f6ba2afd6227327fad227a6"

	apigatewayCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
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
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", registryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
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
		Service:  apigatewayCloudRunService.Name,
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
			RouteName:       apigatewayCloudRunService.Name,
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
			RouteName:       apigatewayCloudRunService.Name,
			CertificateMode: pulumi.String("AUTOMATIC"),
		},
	})
	if err != nil {
		return nil, err
	}

	return apigatewayCloudRunService, nil
}

func newBaemincryptoCloudRunService(ctx *pulumi.Context, project *organizations.Project) (*cloudrun.Service, error) {
	serviceName := "baemincrypto"

	sa, err := serviceaccount.NewAccount(ctx, serviceName, &serviceaccount.AccountArgs{
		Project:     project.ProjectId,
		AccountId:   pulumi.String(serviceName),
		DisplayName: pulumi.String(serviceName),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	imageTag := "2e2dcbb5904afcf069a25b7f0b27d7799d0cd6a5"

	baemincryptoCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
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
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.Sprintf("%s%s:%s", registryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
								ContainerPort: pulumi.Int(50051),
								Name:          pulumi.String("h2c"),
							},
						},
						Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{},
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
		Service:  baemincryptoCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Members: pulumi.StringArray{
			pulumi.Sprintf("serviceAccount:%s", "apigateway@taehoio-staging.iam.gserviceaccount.com"),
			pulumi.Sprintf("user:%s", "taeho@taeho.io"),
		},
	}); err != nil {
		return nil, err
	}

	return baemincryptoCloudRunService, nil
}

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

	youtube2notionCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
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
						Image: pulumi.Sprintf("%s%s:%s", registryBasePath, serviceName, imageTag),
						Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
							cloudrun.ServiceTemplateSpecContainerPortArgs{
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
		Service:  youtube2notionCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	return youtube2notionCloudRunService, nil
}
