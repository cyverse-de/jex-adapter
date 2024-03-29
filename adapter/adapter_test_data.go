package adapter

const testCondorLaunchJSON = `
{
    "description":"this is a description",
    "email":"wregglej@iplantcollaborative.org",
    "name":"Word Count analysis1@@",
    "username":"test@this is a test",
    "app_id":"c7f05682-23c8-4182-b9a2-e09650a5f49b",
    "user_id" : "00000000-0000-0000-0000-000000000000",
    "steps":[
        {
            "component":{
                "container":{
                    "name" : "test-name",
                    "network_mode" : "none",
                    "cpu_shares" : 2048,
                    "memory_limit" : 2048,
                    "entrypoint" : "/bin/true",
                    "id":"16fd2a16-3ac6-11e5-a25d-2fa4b0893ef1",
                    "image":{
                        "id":"fc210a84-f7cd-4067-939c-a68ec3e3bd2b",
                        "url":"https://registry.hub.docker.com/u/discoenv/backwards-compat",
                        "tag":"latest",
                        "name":"gims.iplantcollaborative.org:5000/backwards-compat",
                        "osg_image_path":"/path/to/image"
                    },
                    "container_volumes_from" : [
                      {
                        "name":"vf-name1",
                        "name_prefix":"vf-prefix1",
                        "tag":"vf-tag1",
                        "url":"vf-url1",
                        "read_only":true,
                        "host_path":"/host/path1",
                        "container_path":"/container/path1"
                      },
                      {
                        "name":"vf-name2",
                        "name_prefix":"vf-prefix2",
                        "tag":"vf-tag2",
                        "url":"vf-url2",
                        "read_only":true,
                        "host_path":"/host/path2",
                        "container_path":"/container/path2"
                      }
                    ],
                    "container_volumes" : [
                      {
                        "host_path" : "/host/path1",
                        "container_path" : "/container/path1"
                      },
                      {
                        "host_path" : "",
                        "container_path" : "/container/path2"
                      }
                    ],
                    "container_devices" : [
                      {
                        "host_path" : "/host/path1",
                        "container_path" : "/container/path1"
                      },
                      {
                        "host_path" : "/host/path2",
                        "container_path" : "/container/path2"
                      }
                    ],
                    "working_directory" : "/work"
                },
                "type":"executable",
                "name":"wc_wrapper.sh",
                "location":"/usr/local3/bin/wc_tool-1.00",
                "description":"Word Count"
            },
            "environment":{
                "food" : "banana",
                "foo" : "bar"
            },
            "config":{
                "input":[
                    {
                        "id":"2f58fce9-8183-4ab5-97c4-970592d1c35a",
                        "multiplicity":"single",
                        "name":"Acer-tree.txt",
                        "property":"Acer-tree.txt",
                        "retain":true,
                        "type":"FileInput",
                        "value":"/iplant/home/wregglej/Acer-tree.txt"
                    }
                ],
                "output":[
                    {
                        "multiplicity":"single",
                        "name":"wc_out.txt",
                        "property":"wc_out.txt",
                        "qual-id":"67781636-854a-11e4-b715-e70c4f8db0dc_e7721c78-56c9-41ac-8ff5-8d46093f1fb1",
                        "retain":true,
                        "type":"File"
                    },
                    {
                        "multiplicity":"collection",
                        "name":"logs",
                        "property":"logs",
                        "type":"File",
                        "retain":true
                    }
                ],
                "params":[
                    {
                        "id":"e7721c78-56c9-41ac-8ff5-8d46093f1fb1",
                        "name":"param0",
                        "order":2,
                        "value":"wc_out.txt"
                    },
                    {
                        "id":"2f58fce9-8183-4ab5-97c4-970592d1c35a",
                        "name":"param1",
                        "order":1,
                        "value":"Acer-tree.txt"
                    }
                ]
            },
            "stdin" : "/path/to/stdin",
            "stdout" : "/path/to/stdout",
            "stderr" : "/path/to/stderr",
            "log-file" : "log-file-name",
            "type":"condor"
        }
    ],
    "file-metadata" : [
      {
        "attr" : "attr1",
        "value" : "value1",
        "unit" : "unit1"
      },
      {
        "attr" : "attr2",
        "value" : "value2",
        "unit" : "unit2"
      }
    ],
    "create_output_subdir":true,
    "request_type":"submit",
    "app_description":"this is an app description",
    "output_dir":"/iplant/home/wregglej/analyses/Word_Count_analysis1-2015-09-17-21-42-20.9",
    "wiki_url":"https://pods.iplantcollaborative.org/wiki/display/DEapps/WordCount",
    "uuid":"07b04ce2-7757-4b21-9e15-0b4c2f44be26",
    "notify":true,
    "execution_target":"condor",
    "app_name":"Word Count"
}
`

const testCondorCustomLaunchJSON = `
{
    "description":"this is a description",
    "email":"wregglej@iplantcollaborative.org",
    "name":"Word Count analysis1@@",
    "username":"test@this is a test",
    "app_id":"c7f05682-23c8-4182-b9a2-e09650a5f49b",
    "user_id" : "00000000-0000-0000-0000-000000000000",
    "steps":[
        {
            "component":{
                "container":{
					"max_cpu_cores" : 64,
                    "name" : "test-name",
                    "network_mode" : "none",
                    "cpu_shares" : 2048,
                    "memory_limit" : 2048,
                    "entrypoint" : "/bin/true",
                    "id":"16fd2a16-3ac6-11e5-a25d-2fa4b0893ef1",
                    "image":{
                        "id":"fc210a84-f7cd-4067-939c-a68ec3e3bd2b",
                        "url":"https://registry.hub.docker.com/u/discoenv/backwards-compat",
                        "tag":"latest",
                        "name":"gims.iplantcollaborative.org:5000/backwards-compat",
                        "osg_image_path":"/path/to/image"
                    },
                    "container_volumes_from" : [
                      {
                        "name":"vf-name1",
                        "name_prefix":"vf-prefix1",
                        "tag":"vf-tag1",
                        "url":"vf-url1",
                        "read_only":true,
                        "host_path":"/host/path1",
                        "container_path":"/container/path1"
                      },
                      {
                        "name":"vf-name2",
                        "name_prefix":"vf-prefix2",
                        "tag":"vf-tag2",
                        "url":"vf-url2",
                        "read_only":true,
                        "host_path":"/host/path2",
                        "container_path":"/container/path2"
                      }
                    ],
                    "container_volumes" : [
                      {
                        "host_path" : "/host/path1",
                        "container_path" : "/container/path1"
                      },
                      {
                        "host_path" : "",
                        "container_path" : "/container/path2"
                      }
                    ],
                    "container_devices" : [
                      {
                        "host_path" : "/host/path1",
                        "container_path" : "/container/path1"
                      },
                      {
                        "host_path" : "/host/path2",
                        "container_path" : "/container/path2"
                      }
                    ],
                    "working_directory" : "/work"
                },
                "type":"executable",
                "name":"wc_wrapper.sh",
                "location":"/usr/local3/bin/wc_tool-1.00",
                "description":"Word Count"
            },
            "environment":{
                "food" : "banana",
                "foo" : "bar"
            },
            "config":{
                "input":[
                    {
                        "id":"2f58fce9-8183-4ab5-97c4-970592d1c35a",
                        "multiplicity":"single",
                        "name":"Acer-tree.txt",
                        "property":"Acer-tree.txt",
                        "retain":true,
                        "type":"FileInput",
                        "value":"/iplant/home/wregglej/Acer-tree.txt"
                    }
                ],
                "output":[
                    {
                        "multiplicity":"single",
                        "name":"wc_out.txt",
                        "property":"wc_out.txt",
                        "qual-id":"67781636-854a-11e4-b715-e70c4f8db0dc_e7721c78-56c9-41ac-8ff5-8d46093f1fb1",
                        "retain":true,
                        "type":"File"
                    },
                    {
                        "multiplicity":"collection",
                        "name":"logs",
                        "property":"logs",
                        "type":"File",
                        "retain":true
                    }
                ],
                "params":[
                    {
                        "id":"e7721c78-56c9-41ac-8ff5-8d46093f1fb1",
                        "name":"param0",
                        "order":2,
                        "value":"wc_out.txt"
                    },
                    {
                        "id":"2f58fce9-8183-4ab5-97c4-970592d1c35a",
                        "name":"param1",
                        "order":1,
                        "value":"Acer-tree.txt"
                    }
                ]
            },
            "stdin" : "/path/to/stdin",
            "stdout" : "/path/to/stdout",
            "stderr" : "/path/to/stderr",
            "log-file" : "log-file-name",
            "type":"condor"
        }
    ],
    "file-metadata" : [
      {
        "attr" : "attr1",
        "value" : "value1",
        "unit" : "unit1"
      },
      {
        "attr" : "attr2",
        "value" : "value2",
        "unit" : "unit2"
      }
    ],
    "create_output_subdir":true,
    "request_type":"submit",
    "app_description":"this is an app description",
    "output_dir":"/iplant/home/wregglej/analyses/Word_Count_analysis1-2015-09-17-21-42-20.9",
    "wiki_url":"https://pods.iplantcollaborative.org/wiki/display/DEapps/WordCount",
    "uuid":"07b04ce2-7757-4b21-9e15-0b4c2f44be26",
    "notify":true,
    "execution_target":"condor",
    "app_name":"Word Count"
}
`
