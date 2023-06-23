{{- $period := generate "metricset.period" }}
{{- $agentId := generate "agent.id" }}
{{- $agentVersion := generate "agent.version" }}
{{- $agentName := generate "agent.name" }}
{{- $agentEphemeralid := generate "agent.ephemeral_id" }}
{{- $uId := uuidv4 }}
{{- $pod_uId := uuidv4 }}
{{- $container_uId := uuidv4 }}
{{- $timestamp := generate "timestamp" }}
{{- $fulltimestamp := $timestamp.Format "2006-01-02T15:04:05.999999Z07:00" }}
{{- $resttime := split ":" $fulltimestamp }}
{{- $picktimedate := generate "timedate" }}
{{- $timehour := generate "timehour" }}
{{- $faults := generate "faults" }}
{{- $pct := generate "Percentage" }}
{{- $rangeofid := generate "rangeofid" -}}
{{- $name :=  generate "container.name" }} 
{  "@timestamp": "{{$picktimedate}}T{{$timehour}}:{{ $resttime._1 }}:{{ $resttime._2 }}:{{ $resttime._3}}",
   "container":{
      "memory":{
         "usage": {{divf $pct 1000000}}
      },
      "name":"{{ $name }}",
      "runtime":"containerd",
      "cpu":{
         "usage": {{divf $pct 1000000}}
      },
      "id":"{{ $container_uId }}"
   },
   "kubernetes": {
      "container":{
         "start_time":"{{$picktimedate}}T{{$timehour}}:{{ $resttime._1 }}:{{ $resttime._2 }}:{{ $resttime._3}}",
         "memory":{
            "rss":{
               "bytes": {{generate "Bytes"}}
            },
            "majorpagefaults": {{ $faults }},
            "usage":{
               "node":{
                  "pct": {{divf $pct 1000000}}
               },
               "bytes": {{generate "Bytes"}},
               "limit":{
                  "pct": {{divf $pct 1000000}}
               }
            },
            "available":{
               "bytes": {{generate "Bytes"}}
            },
            "workingset":{
               "bytes": {{generate "Bytes"}},
               "limit":{
                  "pct": {{divf $pct 1000000}}
               }
            },
            "pagefaults": "{{ $faults }}"
         },
         "rootfs":{
            "inodes":{
               "used": {{ generate "kubernetes.container.rootfs.inodes.used" }}
            },
            "available":{
               "bytes": {{generate "Bytes"}}
            },
            "used":{
               "bytes": {{generate "Bytes"}}
            },
            "capacity":{
               "bytes": {{generate "Bytes"}}
            }
         },
         "name":"{{ $name }}",
         "cpu":{
            "usage":{
               "core":{
                  "ns": 41129679
               },
               "node":{
                  "pct": {{divf $pct 1000000}}
               },
               "nanocores":0,
               "limit":{
                  "pct": {{divf $pct 1000000}}
               }
            }
         },
         "logs":{
            "inodes":{
               "count": {{ generate "kubernetes.container.rootfs.inodes.used" }},
               "used":5,
               "free": {{ generate "kubernetes.container.rootfs.inodes.used" }}
            },
            "available":{
               "bytes": {{generate "Bytes"}}
            },
            "used":{
               "bytes": {{generate "Bytes"}}
            },
            "capacity":{
               "bytes": {{generate "Bytes"}}
            }
         }
      },
      "node":{
         "uid": "{{ $uId }}" ,
         "hostname":"{{ $agentName }}.c.elastic-obs-integrations-dev.internal",
         "name":"{{ $agentName }}-{{ $rangeofid }}",
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
         "ip":"{{generate "Ip"}}",
         "name":"demo-deployment-{{ $rangeofid }}",
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
         "url":"https://{{ generate "Ip" }}"
      }
   },
   "service":{
      "address": "https://{{ $agentName }}:10250/stats/summary",
      "type":"kubernetes"
   },
   "data_stream":{
      "namespace":"default",
      "type":"metrics",
      "dataset":"kubernetes.container"
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
      "dataset":"kubernetes.container"
   },
   "host":{
      "hostname":"{{ $agentName }}",
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
      "ip":[
         "10.10.0.186",
         "fe80::4001:aff:fe0a:ba",
         "169.254.123.1",
         "10.232.83.1",
         "fe80::18a2:5bff:fefd:4c7f",
         "10.232.83.1",
         "fe80::7c2c:b0ff:fe97:4fe6",
         "10.232.83.1",
         "fe80::701e:e8ff:fe55:57c8",
         "10.232.83.1",
         "fe80::94be:33ff:fe2c:709d",
         "10.232.83.1",
         "fe80::b86f:65ff:feb6:5f2e",
         "10.232.83.1",
         "fe80::b88a:4ff:fef8:d84e",
         "10.232.83.1",
         "fe80::80d7:1dff:fe53:2932",
         "10.232.83.1",
         "fe80::fc6a:d4ff:fe63:978d",
         "10.232.83.1",
         "fe80::8a9:12ff:fed9:160c",
         "10.232.83.1",
         "fe80::7cfc:9cff:fe70:7887",
         "10.232.83.1",
         "fe80::a053:5ff:fe15:d5c8",
         "10.232.83.1",
         "fe80::c05b:7bff:fec3:6bcb",
         "10.232.83.1",
         "fe80::f045:c9ff:fe18:35ea",
         "10.232.83.1",
         "fe80::2008:d8ff:fee0:2053",
         "10.232.83.1",
         "fe80::3444:ff:fea4:6267",
         "10.232.83.1",
         "fe80::10b1:fbff:fe0d:4c39",
         "10.232.83.1",
         "fe80::80ba:19ff:fe5a:5882",
         "10.232.83.1",
         "fe80::c094:4eff:fe05:705a",
         "10.232.83.1",
         "fe80::1095:69ff:fe35:77b6",
         "10.232.83.1",
         "fe80::88e3:4aff:fe82:3425",
         "10.232.83.1",
         "fe80::5084:b9ff:fedf:390d",
         "10.232.83.1",
         "fe80::50d7:cff:fe81:5f58",
         "10.232.83.1",
         "fe80::f09e:a0ff:fe1d:fa66",
         "10.232.83.1",
         "fe80::4464:f5ff:fe60:3654",
         "10.232.83.1",
         "fe80::b4bb:e9ff:fe22:e9c4",
         "10.232.83.1",
         "fe80::7052:28ff:fec0:bf4c",
         "10.232.83.1",
         "fe80::e4de:78ff:fe5c:aaa",
         "10.232.83.1",
         "fe80::140f:1ff:febe:aaa0",
         "10.232.83.1",
         "fe80::9c95:61ff:fe2e:6535",
         "10.232.83.1",
         "fe80::6079:1cff:fe48:4a7a",
         "10.232.83.1",
         "fe80::cc91:f0ff:fe0e:1e4b",
         "10.232.83.1",
         "fe80::e06b:7cff:fef4:9035",
         "10.232.83.1",
         "fe80::1c46:56ff:fed5:e314",
         "10.232.83.1",
         "fe80::4c99:7bff:fe9d:7a00",
         "10.232.83.1",
         "fe80::1c9c:a1ff:fe09:2a1d",
         "10.232.83.1",
         "fe80::58a1:50ff:feda:6ec5",
         "10.232.83.1",
         "fe80::b887:aeff:fe57:bfe7",
         "10.232.83.1",
         "fe80::b8ae:1eff:fea0:3e10",
         "10.232.83.1",
         "fe80::d8e4:2fff:fe50:6ee0",
         "10.232.83.1",
         "fe80::a422:4eff:febb:5976",
         "10.232.83.1",
         "fe80::1042:2aff:fe63:9a7d",
         "10.232.83.1",
         "fe80::6824:5aff:fe46:be3d",
         "10.232.83.1",
         "fe80::a037:a4ff:feb9:ebc7",
         "10.232.83.1",
         "fe80::a07d:aaff:fe82:4398",
         "10.232.83.1",
         "fe80::e035:f1ff:fec8:81a0",
         "10.232.83.1",
         "fe80::80b2:7ff:feb9:51c7",
         "10.232.83.1",
         "fe80::80de:ceff:fe61:210",
         "10.232.83.1",
         "fe80::a4cc:39ff:fe88:cec4",
         "10.232.83.1",
         "fe80::fcf1:90ff:fe50:28ad",
         "10.232.83.1",
         "fe80::90a0:37ff:fe79:e53",
         "10.232.83.1",
         "fe80::ac54:a8ff:fea4:811f",
         "10.232.83.1",
         "fe80::18ae:21ff:fefd:e840",
         "10.232.83.1",
         "fe80::48d9:98ff:fe0d:852",
         "10.232.83.1",
         "fe80::500b:3aff:fe6b:8717",
         "10.232.83.1",
         "fe80::54a0:b4ff:fec4:f5d5",
         "10.232.83.1",
         "fe80::d45f:99ff:fe9f:d6d8",
         "10.232.83.1",
         "fe80::b033:c7ff:fee9:abfa",
         "10.232.83.1",
         "fe80::c08c:52ff:fe6a:dfc4",
         "10.232.83.1",
         "fe80::3072:88ff:fe3f:bf92",
         "10.232.83.1",
         "fe80::80ed:72ff:fef5:a133",
         "10.232.83.1",
         "fe80::b4c3:ebff:febe:a01",
         "10.232.83.1",
         "fe80::858:8ff:fee1:a643",
         "10.232.83.1",
         "fe80::f4af:c9ff:feec:1a8a",
         "10.232.83.1",
         "fe80::94fa:6dff:feed:b3ab",
         "10.232.83.1",
         "fe80::42c:c2ff:fe0e:ff4b",
         "10.232.83.1",
         "fe80::684b:4eff:fec9:face",
         "10.232.83.1",
         "fe80::cc97:6cff:fe58:40e2",
         "10.232.83.1",
         "fe80::d806:9aff:fe33:79d7",
         "10.232.83.1",
         "fe80::a89c:f4ff:fef0:764",
         "10.232.83.1",
         "fe80::3032:c7ff:fe8c:5411",
         "10.232.83.1",
         "fe80::a1:c8ff:fe0b:5edf",
         "10.232.83.1",
         "fe80::1073:5fff:fe4e:81f7",
         "10.232.83.1",
         "fe80::54d2:97ff:feaa:9a02",
         "10.232.83.1",
         "fe80::d0db:58ff:fe48:af07",
         "10.232.83.1",
         "fe80::2891:aaff:fed5:a33a",
         "10.232.83.1",
         "fe80::44f1:c8ff:fee1:3f37",
         "10.232.83.1",
         "fe80::d4c7:6aff:feab:2495",
         "10.232.83.1",
         "fe80::d8da:dcff:fef5:8384",
         "10.232.83.1",
         "fe80::c494:b2ff:fed7:b34e",
         "10.232.83.1",
         "fe80::806a:73ff:fed9:eac",
         "10.232.83.1",
         "fe80::8879:72ff:fe4a:604a",
         "10.232.83.1",
         "fe80::a43a:53ff:fe19:d4fa",
         "10.232.83.1",
         "fe80::78a0:31ff:fe66:6aa4",
         "10.232.83.1",
         "fe80::a3:48ff:feec:2141",
         "10.232.83.1",
         "fe80::ecec:71ff:fef6:7f16",
         "10.232.83.1",
         "fe80::5c58:24ff:fe8f:101a",
         "10.232.83.1",
         "fe80::606f:dcff:fe9b:b9a5",
         "10.232.83.1",
         "fe80::ac11:3cff:fed5:d25",
         "10.232.83.1",
         "fe80::189f:8eff:fedf:a011",
         "10.232.83.1",
         "fe80::f406:54ff:fe94:b3a3",
         "10.232.83.1",
         "fe80::bc07:e2ff:fea4:ae01",
         "10.232.83.1",
         "fe80::1433:3bff:fe53:5b66",
         "10.232.83.1",
         "fe80::a89b:77ff:feae:bcc6",
         "10.232.83.1",
         "fe80::24a3:40ff:fe0d:6ba0",
         "10.232.83.1",
         "fe80::e050:4eff:fe47:58e2",
         "10.232.83.1",
         "fe80::28d6:97ff:fe96:5615",
         "10.232.83.1",
         "fe80::88b:57ff:fee3:bc1e",
         "10.232.83.1",
         "fe80::184d:f8ff:fe96:4109",
         "10.232.83.1",
         "fe80::c898:ebff:fe50:a559",
         "10.232.83.1",
         "fe80::8400:edff:fea3:e1ad",
         "10.232.83.1",
         "fe80::34f7:7aff:fe03:de89",
         "10.232.83.1",
         "fe80::829:49ff:fee5:3ea5",
         "10.232.83.1",
         "fe80::9cc1:ccff:fe21:8cd2",
         "10.232.83.1",
         "fe80::94cf:87ff:fe1f:7a92",
         "10.232.83.1",
         "fe80::5d:14ff:fec1:2d3a"
      ],
      "name": "{{ $agentName }}",
      "id": "{{ $uId }}",
      "mac":[
         "02-42-1F-C0-F0-D2",
         "02-5D-14-C1-2D-3A",
         "02-A1-C8-0B-5E-DF",
         "02-A3-48-EC-21-41",
         "06-2C-C2-0E-FF-4B",
         "0A-29-49-E5-3E-A5",
         "0A-58-08-E1-A6-43",
         "0A-8B-57-E3-BC-1E",
         "0A-A9-12-D9-16-0C",
         "12-42-2A-63-9A-7D",
         "12-73-5F-4E-81-F7",
         "12-95-69-35-77-B6",
         "12-B1-FB-0D-4C-39",
         "16-0F-01-BE-AA-A0",
         "16-33-3B-53-5B-66",
         "1A-4D-F8-96-41-09",
         "1A-9F-8E-DF-A0-11",
         "1A-A2-5B-FD-4C-7F",
         "1A-AE-21-FD-E8-40",
         "1E-46-56-D5-E3-14",
         "1E-9C-A1-09-2A-1D",
         "22-08-D8-E0-20-53",
         "26-A3-40-0D-6B-A0",
         "2A-91-AA-D5-A3-3A",
         "2A-D6-97-96-56-15",
         "32-32-C7-8C-54-11",
         "32-72-88-3F-BF-92",
         "36-44-00-A4-62-67",
         "36-F7-7A-03-DE-89",
         "42-01-0A-0A-00-BA",
         "46-64-F5-60-36-54",
         "46-F1-C8-E1-3F-37",
         "4A-D9-98-0D-08-52",
         "4E-99-7B-9D-7A-00",
         "52-0B-3A-6B-87-17",
         "52-84-B9-DF-39-0D",
         "52-D7-0C-81-5F-58",
         "56-A0-B4-C4-F5-D5",
         "56-D2-97-AA-9A-02",
         "5A-A1-50-DA-6E-C5",
         "5E-58-24-8F-10-1A",
         "62-6F-DC-9B-B9-A5",
         "62-79-1C-48-4A-7A",
         "6A-24-5A-46-BE-3D",
         "6A-4B-4E-C9-FA-CE",
         "72-1E-E8-55-57-C8",
         "72-52-28-C0-BF-4C",
         "7A-A0-31-66-6A-A4",
         "7E-2C-B0-97-4F-E6",
         "7E-FC-9C-70-78-87",
         "82-6A-73-D9-0E-AC",
         "82-B2-07-B9-51-C7",
         "82-BA-19-5A-58-82",
         "82-D7-1D-53-29-32",
         "82-DE-CE-61-02-10",
         "82-ED-72-F5-A1-33",
         "86-00-ED-A3-E1-AD",
         "8A-79-72-4A-60-4A",
         "8A-E3-4A-82-34-25",
         "92-A0-37-79-0E-53",
         "96-BE-33-2C-70-9D",
         "96-CF-87-1F-7A-92",
         "96-FA-6D-ED-B3-AB",
         "9E-95-61-2E-65-35",
         "9E-C1-CC-21-8C-D2",
         "A2-37-A4-B9-EB-C7",
         "A2-53-05-15-D5-C8",
         "A2-7D-AA-82-43-98",
         "A6-22-4E-BB-59-76",
         "A6-3A-53-19-D4-FA",
         "A6-CC-39-88-CE-C4",
         "AA-9B-77-AE-BC-C6",
         "AA-9C-F4-F0-07-64",
         "AE-11-3C-D5-0D-25",
         "AE-54-A8-A4-81-1F",
         "B2-33-C7-E9-AB-FA",
         "B6-BB-E9-22-E9-C4",
         "B6-C3-EB-BE-0A-01",
         "BA-6F-65-B6-5F-2E",
         "BA-87-AE-57-BF-E7",
         "BA-8A-04-F8-D8-4E",
         "BA-AE-1E-A0-3E-10",
         "BE-07-E2-A4-AE-01",
         "C2-5B-7B-C3-6B-CB",
         "C2-8C-52-6A-DF-C4",
         "C2-94-4E-05-70-5A",
         "C6-94-B2-D7-B3-4E",
         "CA-98-EB-50-A5-59",
         "CE-91-F0-0E-1E-4B",
         "CE-97-6C-58-40-E2",
         "D2-DB-58-48-AF-07",
         "D6-5F-99-9F-D6-D8",
         "D6-C7-6A-AB-24-95",
         "DA-06-9A-33-79-D7",
         "DA-DA-DC-F5-83-84",
         "DA-E4-2F-50-6E-E0",
         "E2-35-F1-C8-81-A0",
         "E2-50-4E-47-58-E2",
         "E2-6B-7C-F4-90-35",
         "E6-DE-78-5C-0A-AA",
         "EE-EC-71-F6-7F-16",
         "F2-45-C9-18-35-EA",
         "F2-9E-A0-1D-FA-66",
         "F6-06-54-94-B3-A3",
         "F6-AF-C9-EC-1A-8A",
         "FE-6A-D4-63-97-8D",
         "FE-F1-90-50-28-AD"
      ],
      "architecture":"x86_64"
   }
}
