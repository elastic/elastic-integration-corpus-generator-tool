{{- $ts := generate "timestamp" }}
{{- $syslogts := generate "syslog_timestamp" }}
{{$ts.Format "2006-1-2T15:04:05Z"}} {{$syslogts.Format "Jan 2 15:04:05"}} {{generate "iporhost"}} {{generate "process_name"}} {{generate "process_pid"}} {{generate "message"}}
