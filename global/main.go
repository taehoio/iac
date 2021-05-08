package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/taehoio/iac"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		project, err := iac.NewProject(ctx, iac.GlobalProjectId, iac.GlobalProjectId)
		if err != nil {
			return err
		}

		if err := runTaehoioDNSRecordSets(ctx, project); err != nil {
			return err
		}

		if _, err := newPulumiCICDServiceAccount(ctx); err != nil {
			return err
		}

		if _, err := newDockerRegistry(ctx); err != nil {
			return err
		}

		return nil
	})
}
