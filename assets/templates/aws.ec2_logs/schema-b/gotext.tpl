{{- $ts := now }}
{{- $ip := generate "aws.ec2.ip_address" }}
{{- $pname := generate "process.name" }}
{{- $logstream := generate "aws.cloudwatch.log_stream" }}
{{- $hostname := generate "host.name" }}
{{- $agentId := generate "agent.id" }}
{
  "@timestamp": "{{ $ts }}",
  "aws.cloudwatch": {
    "log_stream": "{{$logstream}}",
    "ingestion_time": "{{ $ts | date "2006-01-02T15:04:05.000Z" }}",
    "log_group": "/var/log/messages"
  },
  "cloud": {
    "region": ""
  },
  "log.file.path": "/var/log/messages/{{$logstream}}",
  "input": {
    "type": "aws-cloudwatch"
  },
  "data_stream": {
    "namespace": "default",
    "type": "logs",
    "dataset": "generic"
  },
  "process": {
    "name": "{{ $pname }}"
  },
  "message": "{{$ts | date "2006-01-02T15:04:05.000Z"}} {{$ts | date "Jan"}} {{$ts | date "02"}} {{$ts | date "15:04:05"}} {{printf "ip-%s" ($ip | splitList "." | join "-")}} {{$pname}}: {{generate "message"}}",
  "event": {
    "id": "{{ generate "event.id" }}",
    "ingested": "{{ generate "event.ingested" | date "2006-01-02T15:04:05.000000000Z" }}",
    "dataset": "generic"
  },
  "host": {
    "name": "{{$hostname}}"
  },
  "agent": {
    "id": "{{$agentId}}",
    "name": "{{$hostname}}",
    "type": "filebeat",
    "version": "8.8.0",
    "ephemeral_id": "{{$agentId}}"
  },
  "tags": [
    "preserve_original_event"
  ]
}
