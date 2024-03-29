package dns

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/dns"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	taehoioDNSManagedZones = "taehoio"
)

func RunTaehoioDNSRecordSets(ctx *pulumi.Context, project *organizations.Project) error {
	mz, err := dns.NewManagedZone(
		ctx,
		"taehoio",
		&dns.ManagedZoneArgs{
			Project:      project.ProjectId,
			Name:         pulumi.String(taehoioDNSManagedZones),
			DnsName:      pulumi.String("taeho.io."),
			Description:  pulumi.String("taeho.io DNS"),
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
		"taehoio-mx",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("taeho.io."),
			Type:        pulumi.String("MX"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("1 aspmx.l.google.com."),
				pulumi.String("5 alt1.aspmx.l.google.com."),
				pulumi.String("5 alt2.aspmx.l.google.com."),
				pulumi.String("10 alt3.aspmx.l.google.com."),
				pulumi.String("10 alt4.aspmx.l.google.com."),
			},
			Ttl: pulumi.Int(3600),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"taehoio-googleverification-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("vtumu3hh6p25.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("gv-vkfp5py2owa5gf.dv.googlehosted.com."),
			},
			Ttl: pulumi.Int(3600),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"api-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("api.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"api-staging-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("api.staging.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"1on1-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("1on1.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"1on1-staging-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("1on1.staging.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"taehoio-a",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("taeho.io."),
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
		"taehoio-aaaa",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("taeho.io."),
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

	if _, err := dns.NewRecordSet(
		ctx,
		"staging-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("staging.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	if _, err := dns.NewRecordSet(
		ctx,
		"karrot-staging-taehoio-cname",
		&dns.RecordSetArgs{
			Project:     project.ProjectId,
			ManagedZone: mz.Name,
			Name:        pulumi.String("karrot.staging.taeho.io."),
			Type:        pulumi.String("CNAME"),
			Rrdatas: pulumi.StringArray{
				pulumi.String("ghs.googlehosted.com."),
			},
			Ttl: pulumi.Int(5),
		},
		pulumi.Protect(false),
	); err != nil {
		return err
	}

	return nil
}
