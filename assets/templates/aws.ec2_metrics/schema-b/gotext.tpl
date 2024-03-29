{{- /*  metadata */ -}}
{{- $Region := generate "Region" }}
{{- $eventIngested := generate "EventIngested" }}
{{- $eventDuration := generate "EventDuration" }}
{{- /*  availability zone */ -}}
{{- $AvailabilityZone := awsAZFromRegion $Region }}
{{- /*  dimensions */ -}}
{{- $AutoScalingGroupName := generate "AutoScalingGroupName" }}
{{- $ImageId := generate "ImageId" }}
{{- $InstanceId := generate "InstanceId" }}
{{- $instanceTypeIdx := generate "instanceTypeIdx" }}
{{- $InstanceTypeValues := generate "InstanceType" }}
{{- $InstanceType := index $InstanceTypeValues $instanceTypeIdx }}
{{- /*  metrics */ -}}
{{- $StatusCheckFailed_InstanceAvg := generate "StatusCheckFailed_InstanceAvg" }}
{{- $StatusCheckFailed_SystemAvg := generate "StatusCheckFailed_SystemAvg" }}
{{- $StatusCheckFailedAvg := generate "StatusCheckFailedAvg" }}
{{- $CPUUtilizationAvg := generate "CPUUtilizationAvg" }}
{{- $NetworkPacketsInSum := generate "NetworkPacketsInSum" }}
{{- $NetworkPacketsOutSum := generate "NetworkPacketsOutSum" }}
{{- $NetworkInSum := mul $NetworkPacketsInSum 15 }}
{{- $NetworkOutSum := mul $NetworkPacketsOutSum 15 }}
{{- $CPUCreditBalanceAvg := generate "CPUCreditBalanceAvg" }}
{{- $CPUSurplusCreditBalanceAvg := generate "CPUSurplusCreditBalanceAvg" }}
{{- $CPUSurplusCreditsChargedAvg := generate "CPUSurplusCreditsChargedAvg" }}
{{- $CPUCreditUsageAvg := generate "CPUCreditUsageAvg" }}
{{- $DiskReadBytesSum := generate "DiskReadBytesSum" }}
{{- $DiskReadOpsSum := generate "DiskReadOpsSum" }}
{{- $DiskWriteBytesSum := generate "DiskWriteBytesSum" }}
{{- $DiskWriteOpsSum := generate "DiskWriteOpsSum" }}
{{- /*  instance data */ -}}
{{- $instanceCoreCountValues := generate "instanceCoreCount" }}
{{- $instanceCoreCount := index $instanceCoreCountValues $instanceTypeIdx }}
{{- $instanceThreadPerCoreValues := generate "instanceThreadPerCore" }}
{{- $instanceThreadPerCore := index $instanceThreadPerCoreValues $instanceTypeIdx }}
{{- $instanceImageId := generate "instanceImageId" }}
{{- $instanceMonitoringState := generate "instanceMonitoringState" }}
{{- $instancePrivateIP := generate "instancePrivateIP" }}
{{- $instancePrivateDnsEmpty := generate "instancePrivateDnsEmpty" }}
{{- $instancePublicIP := generate "instancePublicIP" }}
{{- $instancePublicDnsEmpty := generate "instancePublicDnsEmpty" }}
{{- $instanceStateName := generate "instanceStateName" }}
{{- $instanceStateCode := 16 }}
{{- if eq $instanceStateName "running" }}
{{- $instanceStateCode = 48 }}
{{- end}}
{{- $cloudInstanceName := generate "cloudInstanceName" }}
{{- /* rate period */ -}}
{{- $period := 60. }}
{{- if eq $instanceMonitoringState "disabled" }}
{{- $period = 300. }}
{{- end}}
{{- /*  ip */ -}}
{{- $instancePrivateDns := "" }}
{{- if eq $instancePrivateDnsEmpty "fromPrivateIP" }}
{{- $instancePrivateDnsPrefix := $instancePrivateIP | replace "." "-" }}
{{- $instancePrivateDns = printf "%s.%s.compute.internal" $instancePrivateDnsPrefix $Region }}
{{- end}}
{{- $instancePublicDns := "" }}
{{- if eq $instancePublicDnsEmpty "fromPublicIP" }}
{{- $instancePublicDnsPrefix := $instancePublicIP | replace "." "-" }}
{{- $instancePublicDns = printf "e2-%s.compute-1.amazonaws.com" $instancePublicDnsPrefix }}
{{- end}}
{{- /*  tags */ -}}
{{- $partOfAutoScalingGroup := generate "partOfAutoScalingGroup" | mod 20 }}{{- /*  5% chance the instance is part of an autoscaling group */ -}}
{{- /*  events */ -}}
{
    "@timestamp": "{{ $eventIngested.Format "2006-01-02T15:04:05.999999Z07:00" }}",
    "ecs.version": "8.0.0",
    "agent": {
        "name": "docker-fleet-agent",
        "id": "2d4b09d0-cdb6-445e-ac3f-6415f87b9864",
        "type": "metricbeat",
        "ephemeral_id": "cdaaaabb-be7e-432f-816b-bda019fd7c15",
        "version": "8.3.2"
    },
    "elastic_agent": {
        "id": "2d4b09d0-cdb6-445e-ac3f-6415f87b9864",
        "version": "8.3.2",
        "snapshot": false
    },
    "cloud": {
        "provider": "aws",
        "region": "{{ $Region }}",
        "account": {
            "name": "elastic-beats",
            "id": "000000000000"
        }
    },
    "ecs": {
        "version": "8.0.0"
    },
    "service": {
        "type": "aws"
    },
    "data_stream": {
        "namespace": "default",
        "type": "metrics",
        "dataset": "aws.ec2_metrics"
    },
    "metricset": {
        "period": 3600000,
        "name": "cloudwatch"
    },
    "event": {
        "duration": {{ $eventDuration }},
        "agent_id_status": "verified",
        "ingested": "{{ $eventIngested.Format "2006-01-02T15:04:05.999999Z07:00" }}",
        "module": "aws",
        "dataset": "aws.ec2_metrics"
    },
    "aws": {
        "cloudwatch": {
            "namespace": "AWS/EC2"
        }
    },
{{ $dimensionType := generate "dimensionType"  }}
{{ if eq $dimensionType "AutoScalingGroupName" }}
    "aws.dimensions.AutoScalingGroupName": "{{ $AutoScalingGroupName }}",
    "aws.ec2.metrics.CPUCreditBalance.avg": {{ $CPUCreditBalanceAvg }},
    "aws.ec2.metrics.CPUCreditUsage.avg": {{ $CPUCreditUsageAvg }},
    "aws.ec2.metrics.CPUSurplusCreditBalance.avg": {{ $CPUSurplusCreditBalanceAvg }},
    "aws.ec2.metrics.CPUSurplusCreditsCharged.avg": {{ $CPUSurplusCreditsChargedAvg }},
    "aws.ec2.metrics.CPUUtilization.avg": {{ $CPUUtilizationAvg }},
    "aws.ec2.metrics.NetworkIn.sum": {{ $NetworkInSum }},
    "aws.ec2.metrics.NetworkOut.sum": {{ $NetworkOutSum }},
    "aws.ec2.metrics.NetworkPacketsIn.sum": {{ $NetworkPacketsInSum }},
    "aws.ec2.metrics.NetworkPacketsOut.sum": {{ $NetworkPacketsOutSum }},
    "aws.ec2.metrics.StatusCheckFailed_Instance.avg": {{ $StatusCheckFailed_InstanceAvg }},
    "aws.ec2.metrics.StatusCheckFailed_System.avg": {{ $StatusCheckFailed_SystemAvg }},
    "aws.ec2.metrics.StatusCheckFailed.avg": {{ $StatusCheckFailedAvg }}
{{ else if eq $dimensionType "ImageId" }}
    "aws.dimensions.ImageId": "{{ $ImageId }}",
    "aws.ec2.metrics.CPUUtilization.avg": {{ $CPUUtilizationAvg }},
    "aws.ec2.metrics.DiskReadBytes.sum": {{ $DiskReadBytesSum }},
    "aws.ec2.metrics.DiskReadOps.sum": {{ $DiskReadOpsSum }},
    "aws.ec2.metrics.DiskWriteBytes.sum": {{ $DiskWriteBytesSum }},
    "aws.ec2.metrics.DiskWriteOps.sum": {{ $DiskWriteOpsSum }},
    "aws.ec2.metrics.NetworkIn.sum": {{ $NetworkInSum }},
    "aws.ec2.metrics.NetworkOut.sum": {{ $NetworkOutSum }}
{{ else if eq $dimensionType "InstanceId" }}
    "aws.dimensions.InstanceId": "{{ $InstanceId }}",
    "aws.ec2.instance.core.count": {{ $instanceCoreCount }},
    "aws.ec2.instance.image.id": "{{ $instanceImageId }}",
    "aws.ec2.instance.monitoring.state": "{{ $instanceMonitoringState }}",
    "aws.ec2.instance.private.dns_name": "{{ $instancePrivateDns }}",
    "aws.ec2.instance.private.ip": "{{ $instancePrivateIP }}",
    "aws.ec2.instance.public.dns_name": "{{ $instancePublicDns }}",
    "aws.ec2.instance.public.ip": "{{ $instancePublicIP }}",
    "aws.ec2.instance.state.code": {{ $instanceStateCode }},
    "aws.ec2.instance.state.name": "{{ $instanceStateName }}",
    "aws.ec2.instance.threads_per_core": {{ $instanceThreadPerCore }},
    "aws.ec2.metrics.CPUCreditBalance.avg": {{ $CPUCreditBalanceAvg }},
    "aws.ec2.metrics.CPUCreditUsage.avg": {{ $CPUCreditUsageAvg }},
    "aws.ec2.metrics.CPUSurplusCreditBalance.avg": {{ $CPUSurplusCreditBalanceAvg }},
    "aws.ec2.metrics.CPUSurplusCreditsCharged.avg": {{ $CPUSurplusCreditsChargedAvg }},
    "aws.ec2.metrics.CPUUtilization.avg": {{ $CPUUtilizationAvg }},
    "aws.ec2.metrics.DiskReadBytes.rate": {{ divf $DiskReadBytesSum $period }},
    "aws.ec2.metrics.DiskReadBytes.sum": {{ $DiskReadBytesSum }},
    "aws.ec2.metrics.DiskReadOps.rate": {{ divf $DiskReadOpsSum $period }},
    "aws.ec2.metrics.DiskReadOps.sum": {{ $DiskReadOpsSum }},
    "aws.ec2.metrics.DiskWriteBytes.rate": {{ divf $DiskWriteBytesSum $period }},
    "aws.ec2.metrics.DiskWriteBytes.sum": {{ $DiskWriteBytesSum }},
    "aws.ec2.metrics.DiskWriteOps.rate": {{ divf $DiskWriteOpsSum $period }},
    "aws.ec2.metrics.DiskWriteOps.sum": {{ $DiskWriteOpsSum }},
    "aws.ec2.metrics.NetworkIn.rate": {{ divf $NetworkInSum $period }},
    "aws.ec2.metrics.NetworkIn.sum": {{ $NetworkInSum }},
    "aws.ec2.metrics.NetworkOut.rate": {{ divf $NetworkOutSum $period }},
    "aws.ec2.metrics.NetworkOut.sum": {{ $NetworkOutSum }},
    "aws.ec2.metrics.NetworkPacketsIn.rate": {{ divf $NetworkPacketsInSum $period }},
    "aws.ec2.metrics.NetworkPacketsIn.sum": {{ $NetworkPacketsInSum }},
    "aws.ec2.metrics.NetworkPacketsOut.rate": {{ divf $NetworkPacketsOutSum $period }},
    "aws.ec2.metrics.NetworkPacketsOut.sum": {{ $NetworkPacketsOutSum }},
    "aws.ec2.metrics.StatusCheckFailed_Instance.avg": {{ $StatusCheckFailed_InstanceAvg }},
    "aws.ec2.metrics.StatusCheckFailed_System.avg": {{ $StatusCheckFailed_SystemAvg }},
    "aws.ec2.metrics.StatusCheckFailed.avg": {{ $StatusCheckFailedAvg }},
    "aws.tags.Name": "{{ $cloudInstanceName }}",
    {{ if eq $partOfAutoScalingGroup 0 }}
    "aws.tags.aws:autoscaling:groupName": "{{ $AutoScalingGroupName }}",
    {{ end }}
    "cloud.availability_zone": "{{ $AvailabilityZone }}",
    "cloud.instance.id": "{{ $InstanceId }}",
    "cloud.instance.name": "{{ $cloudInstanceName }}",
    "cloud.machine.type": "{{ $InstanceType }}",
    "host.cpu.usage": {{ $CPUUtilizationAvg }},
    "host.disk.read.bytes": {{ $DiskReadBytesSum }},
    "host.disk.write.bytes": {{ $DiskWriteBytesSum }},
    "host.id": "{{ $InstanceId }}",
    "host.name": "{{ $cloudInstanceName }}",
    "host.network.egress.bytes": {{ $NetworkOutSum }},
    "host.network.egress.packets": {{ $NetworkPacketsOutSum }},
    "host.network.ingress.bytes": {{ $NetworkInSum }},
    "host.network.ingress.packets": {{ $NetworkPacketsInSum }}
{{ else if eq $dimensionType "InstanceType" }}
    "aws.dimensions.InstanceType": "{{ $InstanceType }}",
    "aws.ec2.metrics.CPUUtilization.avg": {{ $CPUUtilizationAvg }},
    "aws.ec2.metrics.DiskReadBytes.sum": {{ $DiskReadBytesSum }},
    "aws.ec2.metrics.DiskReadOps.sum": {{ $DiskReadOpsSum }},
    "aws.ec2.metrics.DiskWriteBytes.sum": {{ $DiskWriteBytesSum }},
    "aws.ec2.metrics.DiskWriteOps.sum": {{ $DiskWriteOpsSum }},
    "aws.ec2.metrics.NetworkIn.sum": {{ $NetworkInSum }},
    "aws.ec2.metrics.NetworkOut.sum": {{ $NetworkOutSum }}
{{ else }}
    "aws.ec2.metrics.CPUUtilization.avg": {{ $CPUUtilizationAvg }},
    "aws.ec2.metrics.DiskReadBytes.sum": {{ $DiskReadBytesSum }},
    "aws.ec2.metrics.DiskReadOps.sum": {{ $DiskReadOpsSum }},
    "aws.ec2.metrics.DiskWriteBytes.sum": {{ $DiskWriteBytesSum }},
    "aws.ec2.metrics.DiskWriteOps.sum": {{ $DiskWriteOpsSum }},
    "aws.ec2.metrics.NetworkIn.sum": {{ $NetworkInSum }},
    "aws.ec2.metrics.NetworkOut.sum": {{ $NetworkOutSum }}
{{ end }}
}