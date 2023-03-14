{{- $startOffset := generate "StartOffset" }}
{{- $startOffsetInSecond := mul -1 1000000000 $startOffset }}
{{- $startOffsetDuration := $startOffsetInSecond | int64 | duration}}
{{- $end := generate "End" }}
{{- $start := $end.Add $startOffsetDuration}}
{{generate "Version"}} {{generate "AccountID"}} {{generate "InterfaceID"}} {{generate "SrcAddr"}} {{generate "DstAddr"}} {{generate "SrcPort"}} {{generate "DstPort"}} {{generate "Protocol"}}{{ $packets := generate "Packets" }} {{ $packets }} {{mul $packets 15 }} {{$start.Format "2006-01-02T15:04:05.999999Z07:00" }} {{$end.Format "2006-01-02T15:04:05.999999Z07:00"}} {{generate "Action"}}{{ if eq $packets 0 }} NODATA {{ else }} {{generate "LogStatus"}} {{ end }}
