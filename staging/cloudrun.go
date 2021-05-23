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
	notionproxySA := stringOutPtrToStringOutput(notionproxy.Template.Spec().ServiceAccountName())
	apigatewaySA := stringOutPtrToStringOutput(apigateway.Template.Spec().ServiceAccountName())
	baemincryptoSA := stringOutPtrToStringOutput(baemincrypto.Template.Spec().ServiceAccountName())

	_, err = projects.NewIAMBinding(ctx, "service-profiler-agent", &projects.IAMBindingArgs{
		Project: project.ProjectId,
		Members: pulumi.StringArray{
			pulumi.Sprintf("serviceAccount:%s", notionproxySA),
			pulumi.Sprintf("serviceAccount:%s", apigatewaySA),
			pulumi.Sprintf("serviceAccount:%s", baemincryptoSA),
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

	notionproxyCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Location:                 pulumi.String(iac.TokyoLocation),
		Project:                  project.ProjectId,
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Template: cloudrun.ServiceTemplateArgs{
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.String(registryBasePath + serviceName + ":78c93977da22bd4de81c48365ff7b80f06117249"),
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

	apigatewayCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Location:                 pulumi.String(iac.TokyoLocation),
		Project:                  project.ProjectId,
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
					"run.googleapis.com/vpc-access-connector": "projects/taehoio-staging/locations/asia-northeast1/connectors/taehoio-vpc-access",
					"run.googleapis.com/vpc-access-egress":    "all-traffic",
					"autoscaling.knative.dev/maxScale":        "100",
				}),
			},
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.String(registryBasePath + serviceName + ":c6b608f0bfb28e27747e54b25392c67c3fba41f0"),
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
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
		Service:  apigatewayCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	_, err = cloudrun.NewDomainMapping(ctx, "api-staging-taehoio", &cloudrun.DomainMappingArgs{
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
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

	baemincryptoCloudRunService, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
		Location:                 pulumi.String(iac.TokyoLocation),
		Project:                  project.ProjectId,
		Name:                     pulumi.String(serviceName),
		AutogenerateRevisionName: pulumi.Bool(true),
		Metadata: cloudrun.ServiceMetadataArgs{
			Annotations: pulumi.ToStringMap(map[string]string{
				"run.googleapis.com/ingress": "internal",
			}),
		},
		Template: cloudrun.ServiceTemplateArgs{
			Metadata: cloudrun.ServiceTemplateMetadataArgs{
				Annotations: pulumi.ToStringMap(map[string]string{
					"run.googleapis.com/vpc-access-connector": "projects/taehoio-staging/locations/asia-northeast1/connectors/taehoio-vpc-access",
					"run.googleapis.com/vpc-access-egress":    "all-traffic",
					"autoscaling.knative.dev/maxScale":        "100",
				}),
			},
			Spec: cloudrun.ServiceTemplateSpecArgs{
				ContainerConcurrency: pulumi.Int(80),
				Containers: cloudrun.ServiceTemplateSpecContainerArray{
					cloudrun.ServiceTemplateSpecContainerArgs{
						Image: pulumi.String(registryBasePath + serviceName + ":2e2dcbb5904afcf069a25b7f0b27d7799d0cd6a5"),
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

	if _, err := cloudrun.NewIamMember(ctx, serviceName+"-everyone", &cloudrun.IamMemberArgs{
		Location: pulumi.String(iac.TokyoLocation),
		Project:  project.ProjectId,
		Service:  baemincryptoCloudRunService.Name,
		Role:     pulumi.String("roles/run.invoker"),
		Member:   pulumi.String("allUsers"),
	}); err != nil {
		return nil, err
	}

	return baemincryptoCloudRunService, nil
}
