package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
	"github.com/taehoio/iac/production/cloudrun"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		project, err := iac.NewProject(ctx, iac.ProductionProjectId, iac.ProductionProjectId)
		if err != nil {
			return err
		}

		if err := cloudrun.RunCloudRunServices(ctx, project); err != nil {
			return err
		}

		return nil
	})
}
