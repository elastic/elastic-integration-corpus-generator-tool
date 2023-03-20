{{- $startOffset := generate "StartOffset" }}
{{- $end := generate "End" }}
{{- $start := mul -1 $startOffset | int64 | duration | $end.Add }}
{{generate "Version"}} {{generate "AccountID"}} {{generate "InterfaceID"}} {{generate "SrcAddr"}} {{generate "DstAddr"}} {{generate "SrcPort"}} {{generate "DstPort"}} {{generate "Protocol"}}{{ $packets := generate "Packets" }} {{ $packets }} {{mul $packets 15 }} {{$start.Format "2006-01-02T15:04:05.999999Z07:00" }} {{$end.Format "2006-01-02T15:04:05.999999Z07:00"}} {{generate "Action"}}{{ if eq $packets 0 }} NODATA {{ else }} {{generate "LogStatus"}} {{ end }}
