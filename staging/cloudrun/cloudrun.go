package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func RunCloudRunServices(ctx *pulumi.Context, project *organizations.Project) error {
	notionproxy, err := newNotionproxyCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	apigateway, err := newApigatewayCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	user, err := newUserCloudRunService(ctx, project)
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
	taehoioStrapi, err := newTaehoioStrapiCloudRunService(ctx, project)
	if err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-profiler-agent",
		"roles/cloudprofiler.agent",
		notionproxy,
		apigateway,
		user,
		baemincrypto,
		youtube2notion,
		taehoioStrapi,
	); err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-trace-agent",
		"roles/cloudtrace.agent",
		notionproxy,
		apigateway,
		user,
		baemincrypto,
		youtube2notion,
		taehoioStrapi,
	); err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-cloud-sql",
		"roles/cloudsql.client",
		user,
		taehoioStrapi,
	); err != nil {
		return err
	}

	return nil
}

func newIAMBinding(
	ctx *pulumi.Context,
	project *organizations.Project,
	name string,
	role string,
	svcs ...*cloudrun.Service,
) error {
	var members pulumi.StringArray
	for _, svc := range svcs {
		members = append(
			members,
			pulumi.Sprintf(
				"serviceAccount:%s",
				stringOutPtrToStringOutput(svc.Template.Spec().ServiceAccountName()),
			),
		)
	}

	_, err := projects.NewIAMBinding(
		ctx,
		name,
		&projects.IAMBindingArgs{
			Project: project.ProjectId,
			Members: members,
			Role:    pulumi.String(role),
		},
	)
	return err
}

func stringOutPtrToStringOutput(spo pulumi.StringPtrOutput) pulumi.StringOutput {
	return spo.ApplyT(func(sp *string) string {
		return *sp
	}).(pulumi.StringOutput)
}