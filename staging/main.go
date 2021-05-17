package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		project, err := iac.NewProject(ctx, iac.StagingProjectId, iac.StagingProjectId)
		if err != nil {
			return err
		}

		if err := runCloudRunServices(ctx, project); err != nil {
			return err
		}

		return nil
	})
}
