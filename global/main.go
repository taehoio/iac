package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/taehoio/iac"
	"github.com/taehoio/iac/global/dns"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Project
		project, err := iac.NewProject(ctx, iac.GlobalProjectId, iac.GlobalProjectId)
		if err != nil {
			return err
		}

		// DNS
		if err := dns.RunTaehoioDNSRecordSets(ctx, project); err != nil {
			return err
		}
		if err := dns.RunGabojagocomDNSRecordSets(ctx, project); err != nil {
			return err
		}

		// Service Accounts
		if _, err := newPulumiCICDServiceAccount(ctx, project); err != nil {
			return err
		}
		if _, err := newDockerRegistryServiceAccount(ctx, project); err != nil {
			return err
		}

		// Artifact Registry
		if _, err := newDockerRegistry(ctx, project); err != nil {
			return err
		}

		return nil
	})
}
