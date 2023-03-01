{{- $ts := now }}
{{- $ip := generate "aws.ec2.ip_address" }}
{{- $pname := generate "process.name" }}
{
    "data_stream": {
        "namespace": "default",
        "type": "logs",
        "dataset": "aws.ec2_logs"
    },
    "process": {
        "name": "{{ $pname }}"
    },
    "message": "{{$ts | date "2006-01-02T15:04:05.000Z"}} {{$ts | date "Jan"}} {{$ts | date "02"}} {{$ts | date "15:04:05"}} {{printf "ip-%s" ($ip | splitList "." | join "-")}} {{$pname}}: {{generate "message"}}",
    "event": {
        "ingested": "{{ generate "event.ingested" | date "2006-01-02T15:04:05.000000000Z" }}",
     },
    "aws": {
        "ec2": {
            "ip_address": "{{ $ip }}"
        }
    },
    "tags": [
        "preserve_original_event"
    ]
}
