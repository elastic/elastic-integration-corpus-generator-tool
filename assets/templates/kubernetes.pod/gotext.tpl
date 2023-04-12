{{- $period := generate "metricset.period" }}
{{- $timestamp := generate "timestamp" }}
{{- $agentId := generate "agent.id" }}
{{- $agentVersion := generate "agent.version" }}
{{- $agentName := generate "agent.name" }}
{{- $agentEphemeralid := generate "agent.ephemeral_id" }}
{{- $rxbytes := generate "container.network.ingress.bytes" }}
{{- $txbytes := generate "container.network.egress.bytes" }}

{       
    "@timestamp": "{{$timestamp.Format "2006-01-02T15:04:05.999999Z07:00"}}",
    "container":{
      "network":{
         "ingress":{
            "bytes": {{ $rxbytes }} 
         },
         "egress":{
            "bytes": {{ $txbytes}} 
         }
      }
   },
   "kubernetes": {
    "node":{
         "uid":"56f352ee-ea23-4299-9c71-cd74c523b0f6",
         "hostname":"gke-kubernetes-scale-kubernetes-scale-0f73d58f-2qqz.c.elastic-obs-integrations-dev.internal",
         "name":"gke-kubernetes-scale-kubernetes-scale-0f73d58f-2qqz",
         "labels":{
            "cloud_google_com/machine-family":"e2",
            "cloud_google_com/gke-nodepool":"kubernetes-scale-nl",
            "kubernetes_io/hostname":"gke-kubernetes-scale-kubernetes-scale-0f73d58f-2qqz",
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
         "uid":"390ec2bb-1bc3-492e-980f-37fc91a6eca8",
         "start_time":"2023-03-21T09:50:54Z",
         "memory":{
            "rss":{
               "bytes":1216512
            },
            "major_page_faults":0,
            "usage":{
               "node":{
                  "pct":0.00011638926817746606
               },
               "bytes":1953792,
               "limit":{
                  "pct":0.00011638926817746606
               }
            },
            "available":{
               "bytes":0
            },
            "page_faults":1386,
            "working_set":{
               "bytes":1953792,
               "limit":{
                  "pct":0.00011638926817746606
               }
            }
         },
         "ip":"10.232.83.124",
         "name":"demo-deployment-8-78469f5d79-qtx69",
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
               "bytes":132,
               "errors":0
            },
            "rx":{
               "bytes":446,
               "errors":0
            }
         }
      },
      "namespace":"demo-8",
      "namespace_uid":"df8f15a6-a622-4e96-a867-d63a52986a95",
      "replicaset":{
         "name":"demo-deployment-8-78469f5d79"
      },
      "namespace_labels":{
         "kubernetes_io/metadata_name":"demo-8"
      },
      "labels":{
         "app":"demo",
         "pod-template-hash":"78469f5d79",
         "app-2":"demo-2",
         "app-1":"demo-1"
      },
      "deployment":{
         "name":"demo-deployment-8"
      }
   },
    "cloud": {
        "provider": "gcp",
        "availability_zone": "europe-west1-d",
        "instance":{
         "name":  {{ $agentName }} ,
         "id": {{ $agentId }}
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
            "url":"https://10.10.0.2"
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
        "id":  {{ $agentId}},
        "name": {{ $agentName }} ,
        "type": "metricbeat",
        "version": {{ $agentVersion }},
        "ephemeral_id": {{ $agentEphemeralid }}
    },
    "elastic_agent": {
        "id": {{ $agentId }} ,
        "version": {{ $agentVersion }}",
        "snapshot": "false"
    },
    "metricset":{
        "period": {{ $period }} ,
        "name":"pod"
    },
    "event":{
        "duration": "{{generate "event.duration"}}",
        "agent_id_status": "verified",
        "ingested": {{ $timestamp.Format "2006-01-02T15:04:05.999999Z07:00" }},
        "module":"kubernetes",
        "dataset":"kubernetes.pod"
    }
}
