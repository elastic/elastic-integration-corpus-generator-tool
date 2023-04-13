{{- $period := generate "metricset.period" }}
{{- $timestamp := generate "timestamp" }}
{{- $agentId := generate "agent.id" }}
{{- $agentVersion := generate "agent.version" }}
{{- $agentName := generate "agent.name" }}
{{- $agentEphemeralid := generate "agent.ephemeral_id" }}
{{- $rxbytes := generate "container.network.ingress.bytes" | int }}
{{- $txbytes := generate "container.network.egress.bytes" | int }}
{{- $uId := uuidv4 }}
{{- $pod_uId := uuidv4 }}
{{- $suffix := split "-" $uId }}
{{- $offset := generate "Offset" | int }}
{{- $pct := generate "Percentage" | float64 }}
{"@timestamp": "{{$timestamp.Format "2006-01-02T15:04:05.999999Z07:00"}}",
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
         "hostname":"{{ $agentName }}.c.elastic-obs-integrations-dev.internal",
         "name":"{{ $agentName }}-{{ $suffix._0 }}",
         "labels":{
            "cloud_google_com/machine-family":"e2",
            "cloud_google_com/gke-nodepool":"kubernetes-scale-nl",
            "kubernetes_io/hostname":"{{ $agentName }}-{{ $uId }}",
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
         "start_time": "{{$timestamp.Format "2006-01-02T15:04:05.999999Z07:00"}}",
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
         "name":"demo-deployment-{{ $offset }}-{{ $suffix._0 }}",
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
      "namespace":"demo-{{ $offset }}",
      "namespace_uid":"demo-{{ $offset }}",
      "replicaset":{
         "name":"demo-deployment-{{ $offset }}-{{ $suffix._0 }}"
      },
      "namespace_labels":{
         "kubernetes_io/metadata_name":"demo-{{ $offset }}"
      },
      "labels":{
         "app":"demo",
         "pod-template-hash":"{{ $suffix._0 }}",
         "app-2":"demo-2",
         "app-1":"demo-1"
      },
      "deployment":{
         "name":"demo-deployment-{{ $offset }}"
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
        "snapshot": "false"
    },
    "metricset":{
        "period": "{{ $period }}" ,
        "name":"pod"
    },
    "event":{
        "duration": "{{generate "event.duration"}}",
        "agent_id_status": "verified",
        "ingested": "{{ $timestamp.Format "2006-01-02T15:04:05.999999Z07:00" }}",
        "module":"kubernetes",
        "dataset":"kubernetes.pod"
    }
}
