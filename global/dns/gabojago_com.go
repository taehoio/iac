package dns

import (
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/dns"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	gabojagocomDNSManagedZones = "gabojagocom"
)

func RunGabojagocomDNSRecordSets(ctx *pulumi.Context, project *organizations.Project) error {
	mz, err := dns.NewManagedZone(
		ctx,
		"gabojagocom",
		&dns.ManagedZoneArgs{
			Project:      project.ProjectId,
			Name:         pulumi.String(gabojagocomDNSManagedZones),
			DnsName:      pulumi.String("gabojago.com."),
			Description:  pulumi.String("gabojago.com DNS"),
			ForceDestroy: pulumi.Bool(false),
			Visibility:   pulumi.String("public"),
		},
		pulumi.Protect(false),
	)
	if err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"gabojagocom-a",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("gabojago.com."),
			Type:        pulumi.String("A"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("216.239.32.21"),
				pulumi.String("216.239.34.21"),
				pulumi.String("216.239.36.21"),
				pulumi.String("216.239.38.21"),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"gabojagocom-aaaa",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("gabojago.com."),
			Type:        pulumi.String("AAAA"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("2001:4860:4802:32::15"),
				pulumi.String("2001:4860:4802:34::15"),
				pulumi.String("2001:4860:4802:36::15"),
				pulumi.String("2001:4860:4802:38::15"),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	return nil
}
