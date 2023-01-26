{ "create" : { "_index": "metrics-aws.sqs-default" } }
{ "@timestamp": "2022-07-26T21:43:00.000Z", "agent": { "name": "docker-fleet-agent", "id": "2d4b09d0-cdb6-445e-ac3f-6415f87b9864", "type": "metricbeat", "ephemeral_id": "cdaaaabb-be7e-432f-816b-bda019fd7c15", "version": "8.3.2" }, "elastic_agent": { "id": "2d4b09d0-cdb6-445e-ac3f-6415f87b9864", "version": "8.3.2", "snapshot": false }, "cloud": { "provider": "aws", "region": "{{ .Region }}", "account": { "name": "elastic-beats", "id": "000000000000" } }, "ecs": { "version": "8.0.0" }, "service": { "type": "aws" }, "data_stream": { "namespace": "default", "type": "metrics", "dataset": "aws.sqs" }, "metricset": { "period": 300000, "name": "cloudwatch" }, "event": { "duration": {{ .EventDuration }}, "agent_id_status": "verified", "ingested": "{{ .EventIngested }}", "module": "aws", "dataset": "aws.sqs" }, "aws": { "cloudwatch": { "namespace": "AWS/SQS" }, "dimensions": { "QueueName": "{{ .QueueName }}" }, "sqs": { "metrics": { "ApproximateAgeOfOldestMessage": { "avg": {{ .OldestMessageAge }} }, "ApproximateNumberOfMessagesDelayed": { "avg": {{ .Delayed }} }, "ApproximateNumberOfMessagesNotVisible": { "avg": {{ .NotVisible }} }, "ApproximateNumberOfMessagesVisible": { "avg": {{ .Visible }} }, "NumberOfMessagesDeleted": { "avg": {{ .Deleted }} }, "NumberOfMessagesReceived": { "avg": {{ .Received }} }, "NumberOfMessagesSent": { "avg": {{ .Sent }} }, "NumberOfEmptyReceives": { "avg": {{ .EmptyReceives }} }, "SentMessageSize": { "avg": {{ .SentMessageSize }} }, } }, "tags": { "createdBy": "{{ .TagsCreatedBy }}" } }, "cloud": { "account": { "id": "000000000000", "name": "elastic-observability" }, "provider": "aws", "region": "{{ .Region }}" } }