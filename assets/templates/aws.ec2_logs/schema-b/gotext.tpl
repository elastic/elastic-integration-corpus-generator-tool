{{- $currency := generate "aws.billing.currency"}}
{
    "@timestamp": "{{generate "timestamp"}}",
    "cloud": {
        "provider": "aws",
        "region": "{{generate "cloud.region"}}",
        "account": {
            "id": "{{generate "cloud.account.id"}}",
            "name": "{{generate "cloud.account.name"}}"
        }
    },
    "event": {
        "dataset": "aws.billing",
        "module": "aws",
        "duration": {{generate "event.duration"}}
    },
    "metricset": {
        "name": "billing",
        "period": {{generate "metricset.period"}}
    },
    "ecs": {
        "version": "1.5.0"
    },
    "aws": {
        "billing": {
            "Currency": "{{$currency}}",
            "EstimatedCharges": {{generate "aws.billing.EstimatedCharges"}},
            "ServiceName": "{{generate "aws.billing.ServiceName"}}",
            "AmortizedCost": {
                "amount": {{generate "aws.billing.AmortizedCost.amount"}},
                "unit": "USD"
            },
            "BlendedCost": {
                "amount": {{generate "aws.billing.BlendedCost.amount"}},
                "unit": "{{$currency}}"
            },
            "NormalizedUsageAmount": {
                "amount": {{generate "aws.billing.NormalizedUsageAmount.amount"}},
                "unit": "N/A"
            },
            "UnblendedCost": {
                "amount": {{generate "aws.billing.UnblendedCost.amount"}},
                "unit": "{{$currency}}"
            },
            "UsageQuantity": {
                "amount": {{generate "aws.billing.UsageQuantity.amount"}},
                "unit": "N/A"
            }
        }
    },
    "service": {
        "type": "aws"
    },
    "agent": {
        "id": "{{generate "agent.id"}}",
        "name": "{{generate "agent.name"}}",
        "type": "metricbeat",
        "version": "8.0.0",
        "ephemeral_id": "{{generate "agent.ephemeral_id"}}"
    }
}
