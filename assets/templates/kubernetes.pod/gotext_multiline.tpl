{{- $period := generate "metricset.period" }}
{{- $timestamp := generate "timestamp" }}
{{- $fulltimestamp := $timestamp.Format "2006-01-02T15:04:05.999999Z07:00" }}
{{- $resttime := split ":" $fulltimestamp }}
{{- $picktimedate := generate "timedate" }}
{{- $timehour := generate "timehour" }}
{{- $agentId := generate "agent.id" }}
{{- $agentVersion := generate "agent.version" }}
{{- $agentName := generate "agent.name" }}
{{- $agentEphemeralid := generate "agent.ephemeral_id" }}
{{- $rxbytes := generate "container.network.ingress.bytes" }}
{{- $txbytes := generate "container.network.egress.bytes" }}
{{- $uId := uuidv4 }}
{{- $pod_uId := uuidv4 }}
{{- $rangeofid := generate "rangeofid" }}
{{- $nodeid := generate "nodeid" -}}
{{- $pct := generate "Percentage" }}
{  "@timestamp": "{{$picktimedate}}T{{$timehour}}:{{ $resttime._1 }}:{{ $resttime._2 }}:{{ $resttime._3}}",
   "container":{
      "network":{
         "ingress":{
            "bytes": {{ $rxbytes }} 
         },
         "egress":{
            "bytes": {{ $txbytes }} 
         }
      }
   },
   "kubernetes": {
      "node":{
         "uid": "{{ $uId }}" ,
         "hostname":"node-{{ $nodeid }}",
         "name":"node-{{ $nodeid }}",
         "labels":{
            "cloud_google_com/machine-family":"e2",
            "cloud_google_com/gke-nodepool":"kubernetes-scale-nl",
            "kubernetes_io/hostname":"node-{{ $nodeid }}",
            "cloud_google_com/gke-os-distribution":"cos",
            "topology_kubernetes_io/zone":"europe-west1-d",
            "topology_gke_io/zone":"europe-west1-d",
            "topology_kubernetes_io/region":"europe-west1",
            "kubernetes_io/arch":"amd64",
            "cloud_google_com/gke-cpu-scaling-level":"4",
            "env":"kubernetes-scale",
            "failure-domain_beta_kubernetes_io/region":"europe-west1",
            "cloud_google_com/gke-max-pods-per-node":"110",
            "cloud_google_com/gke-container-runtime":"containerd",
            "beta_kubernetes_io/instance-type":"e2-standard-4",
            "failure-domain_beta_kubernetes_io/zone":"europe-west1-d",
            "node_kubernetes_io/instance-type":"e2-standard-4",
            "beta_kubernetes_io/os":"linux",
            "cloud_google_com/gke-boot-disk":"pd-balanced",
            "kubernetes_io/os":"linux",
            "cloud_google_com/private-node":"false",
            "cloud_google_com/gke-logging-variant":"DEFAULT",
            "beta_kubernetes_io/arch":"amd64"
         }
      },
      "pod":{
         "uid": "{{ $pod_uId }}",
         "start_time": "{{$picktimedate}}T{{$timehour}}:{{ $resttime._1 }}:{{ $resttime._2 }}:{{ $resttime._3}}",
         "memory":{
            "rss":{
               "bytes":"{{generate "Bytes"}}"
            },
            "major_page_faults":0,
            "usage":{
               "node":{
                  "pct": "{{divf $pct 1000000}}"
               },
               "bytes": "{{generate "Bytes"}}",
               "limit":{
                  "pct":"{{divf $pct 1000000}}"
               }
            },
            "available":{
               "bytes":0
            },
            "page_faults":1386,
            "working_set":{
               "bytes": "{{generate "Bytes"}}",
               "limit":{
                  "pct": "{{divf $pct 1000000}}"
               }
            }
         },
         "ip":"{{generate "Ip"}}",
         "name":"demo-deployment-{{ $rangeofid }}",
         "cpu":{
            "usage":{
               "node":{
                  "pct":0
               },
               "nanocores":0,
               "limit":{
                  "pct":0
               }
            }
         },
         "network":{
            "tx":{
               "bytes": {{ $txbytes }},
               "errors":0
            },
            "rx":{
               "bytes": {{ $rxbytes }},
               "errors":0
            }
         }
      },
      "namespace":"demo-{{ $rangeofid }}",
      "namespace_uid":"demo-{{ $rangeofid }}",
      "replicaset":{
         "name":"demo-deployment-{{ $rangeofid }}"
      },
      "namespace_labels":{
         "kubernetes_io/metadata_name":"demo-{{ $rangeofid }}"
      },
      "labels":{
         "app":"demo",
         "pod-template-hash":"{{ $rangeofid }}",
         "app-2":"demo-2",
         "app-1":"demo-1"
      },
      "deployment":{
         "name":"demo-deployment-{{ $rangeofid }}"
      }
   },
   "cloud": {
      "provider": "gcp",
      "availability_zone": "europe-west1-d",
      "instance":{
      "name":  "{{ $agentName }}" ,
      "id": "{{ $agentId }}"
      },
      "machine":{
         "type":"e2-standard-4"
      },
      "service":{
         "name":"GCE"
      },
      "project":{
         "id":"elastic-obs-integrations-dev"
      },
      "account":{
         "id":"elastic-obs-integrations-dev"
      }
   },
   "orchestrator":{
      "cluster":{
         "name":"kubernetes-scale",
         "url":"https://{{generate "Ip"}}"
      }
   },
   "service":{
      "address": "https://{{ $agentName }}:10250/stats/summary",
      "type":"kubernetes"
   },
   "data_stream":{
      "namespace":"default",
      "type":"metrics",
      "dataset":"kubernetes.pod"
   },
   "ecs": {
      "version": "8.2.0"
   },
   "agent": {
      "id":  "{{ $agentId}}",
      "name": "{{ $agentName }}" ,
      "type": "metricbeat",
      "version": "{{ $agentVersion }}",
      "ephemeral_id": "{{ $agentEphemeralid }}"
   },
   "elastic_agent": {
      "id": "{{ $agentId }}" ,
      "version": "{{ $agentVersion }}",
      "snapshot": {{ generate "agent.snapshot" }}
   },
   "metricset":{
      "period": "{{ $period }}" ,
      "name":"pod"
   },
   "event":{
      "duration": "{{generate "event.duration"}}",
      "agent_id_status": "verified",
      "ingested": "{{$picktimedate}}T{{$timehour}}:{{ $resttime._1 }}:{{ $resttime._2 }}:{{ $resttime._3}}",
      "module":"kubernetes",
      "dataset":"kubernetes.pod"
   },
   "host":{
      "hostname":"node-{{ $nodeid }}",
      "os":{
         "kernel":"5.10.161+",
         "codename":"focal",
         "name":"Ubuntu",
         "type":"linux",
         "family":"debian",
         "version":"20.04.5 LTS (Focal Fossa)",
         "platform":"ubuntu"
      },
      "containerized":false,
      "name": "{{- $nodeid := generate "nodeid" -}}",
      "id": "{{ $uId }}",
      "architecture":"x86_64"
   }
}