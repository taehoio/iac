package cloudrun

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func RunCloudRunServices(ctx *pulumi.Context, project *organizations.Project) error {
	oneononeapp, err := newOneononeappCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	apigateway, err := newApigatewayCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	user, err := newUserCloudRunService(ctx, project, []*cloudrun.Service{apigateway})
	if err != nil {
		return err
	}
	auth, err := newAuthCloudRunService(ctx, project, []*cloudrun.Service{apigateway, user})
	if err != nil {
		return err
	}
	baemincrypto, err := newBaemincryptoCloudRunService(ctx, project, []*cloudrun.Service{apigateway})
	if err != nil {
		return err
	}
	oneonone, err := newOneononeCloudRunService(ctx, project, []*cloudrun.Service{apigateway})
	if err != nil {
		return err
	}
	youtube2notion, err := newYoutube2notionCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	karrot, err := newKarrotCloudRunService(ctx, project)
	if err != nil {
		return err
	}
	texttospeech, err := newTexttospeechCloudRunService(ctx, project, []*cloudrun.Service{apigateway})
	if err != nil {
		return err
	}
	car, err := newCarCloudRunService(ctx, project, []*cloudrun.Service{apigateway})
	if err != nil {
		return err
	}
	api, err := newApiCloudRunService(ctx, project)
	if err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-profiler-agent",
		"roles/cloudprofiler.agent",
		[]*cloudrun.Service{
			oneononeapp,
			apigateway,
			user,
			auth,
			baemincrypto,
			oneonone,
			youtube2notion,
			karrot,
			texttospeech,
			car,
			api,
		},
	); err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-trace-agent",
		"roles/cloudtrace.agent",
		[]*cloudrun.Service{
			oneononeapp,
			apigateway,
			user,
			auth,
			baemincrypto,
			oneonone,
			youtube2notion,
			karrot,
			texttospeech,
			car,
			api,
		},
	); err != nil {
		return err
	}

	if err := newIAMBinding(
		ctx,
		project,
		"service-cloud-sql",
		"roles/cloudsql.client",
		[]*cloudrun.Service{
			user,
			oneonone,
			api,
		},
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
	svcs []*cloudrun.Service,
) error {
	_, err := projects.NewIAMBinding(
		ctx,
		name,
		&projects.IAMBindingArgs{
			Project: project.ProjectId,
			Members: servicesToMembers(svcs),
			Role:    pulumi.String(role),
		},
	)
	return err
}

func servicesToMembers(svcs []*cloudrun.Service) pulumi.StringArray {
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
	return members
}

func stringOutPtrToStringOutput(spo pulumi.StringPtrOutput) pulumi.StringOutput {
	return spo.ApplyT(func(sp *string) string {
		return *sp
	}).(pulumi.StringOutput)
}
