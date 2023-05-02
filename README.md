<a href="https://github.com/SENERGY-Platform/device-selection/actions/workflows/tests.yml" rel="nofollow">
    <img src="https://github.com/SENERGY-Platform/device-selection/actions/workflows/tests.yml/badge.svg?branch=master" alt="Tests" />
</a>

This service allowes a user to find devices (and services) which match given filter criteria.

the endpoint expects a jwt in the Authorization header.


## Filter By Simple Query-Parameter

if the request correlates to a single bpmn process task (or event), the user can use the following query-parameters:
- function_id
- device_class_id
- aspect_id


**request:**
```
GET /selectables?function_id=someId1&aspect_id=someOtherId
```

**response:**
```
[{
   "device":{
      "id":"example-id",
      "local_id":"example-local-id",
      "name":"example-name",
      "device_type_id":"example-device-type"
      "creator": "creator",
      "shared": true,
      "permission": {"r": true, "w": false, "x": true", "a": false"}
   },
   "services":[
      {
         "id":"example-service-id",
         "local_id":"example-local-id",
         "name":"example-name",
         ...
      }
   ]
}]
```

## Filter By JSON

if the request correlates to multiple bpmn tasks in a bpmn lane, the simple query parameters and powerful enough.
in this case the user may send a json list of the filter-criteria to the endpoint.
the json has to be query-encoded and set to the 'json' query-parameter. 

**json example:**
```
[
   {
      "function_id":"some-function-id",
      "device_class_id":"some-device-clsass-id",
      "aspect_id":"some-aspect-id"
   }
]
```  

**request:**
```
GET /selectables?json=%5B%7B%22function_id%22%3A%22some-function-id%22%2C%22device_class_id%22%3A%22some-device-clsass-id%22%2C%22aspect_id%22%3A%22some-aspect-id%22%7D%5D
```

## Filter By Base64 JSON
the json may also be send base64 encoded

**request:**
```
GET /selectables?base64=W3siZnVuY3Rpb25faWQiOiJzb21lLWZ1bmN0aW9uLWlkIiwiZGV2aWNlX2NsYXNzX2lkIjoic29tZS1kZXZpY2UtY2xzYXNzLWlkIiwiYXNwZWN0X2lkIjoic29tZS1hc3BlY3QtaWQifV0%3D
```

## Protocol-Block-List

the user may filter the results additionally with a blocklist of brotocol ids, by using the query-parameter 'filter_protocols'.
filter_protocols may be a ',' separated list of protocol-ids.

**request:**
```
GET /selectables?function_id=someId1&aspect_id=someOtherId&filter_protocols=id1,id2
```

the list of blocked protocols may also be set implicitly by using the query-parameter 'filter_interaction'.
this creates a list of unwanted protocols, that use the given interaction.

the value of 'filter_interaction' may be one of the following values:
- event
- request
- event+request

**request:**
```
GET /selectables?function_id=someId1&aspect_id=someOtherId&filter_interaction=event
```

## Bulk Request

**request:**
```
POST /bulk/selectables
[
   {
      "id":"1",
      "filter_interaction":null,
      "filter_protocols":[
         "mqtt"
      ],
      "criteria":[
         {
            "function_id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
            "device_class_id":"dc1",
            "aspect_id":"a1"
         }
      ]
   },
   {
      "id":"2",
      "filter_interaction":"event",
      "filter_protocols":null,
      "criteria":[
         {
            "function_id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
            "device_class_id":"dc1",
            "aspect_id":"a1"
         }
      ]
   },
   {
      "id":"3",
      "filter_interaction":null,
      "filter_protocols":[
         "mqtt",
         "pid"
      ],
      "criteria":[
         {
            "function_id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
            "device_class_id":"unknown",
            "aspect_id":"a1"
         }
      ]
   }
]
```


**response:**
```
[
   {
      "id":"1",
      "selectables":[
         {
            "device":{
               "id":"1",
               "name":"1",
               "device_type_id":"dt1",
               "permissions":{
                  "r":true,
                  "w":false,
                  "x":true,
                  "a":false
               },
               "shared":false,
               "creator":""
            },
            "services":[
               {
                  "id":"11",
                  "local_id":"11_l",
                  "name":"11_name",
                  "aspects":[
                     {
                        "id":"a1",
                        "name":"",
                        "rdf_type":""
                     }
                  ],
                  "protocol_id":"pid",
                  "functions":[
                     {
                        "id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
                        "name":"",
                        "concept_id":"",
                        "rdf_type":"https://senergy.infai.org/ontology/MeasuringFunction"
                     }
                  ]
               }
            ]
         }
      ]
   },
   {
      "id":"2",
      "selectables":[
         {
            "device":{
               "id":"1",
               "name":"1",
               "device_type_id":"dt1",
               "permissions":{
                  "r":true,
                  "w":false,
                  "x":true,
                  "a":false
               },
               "shared":false,
               "creator":""
            },
            "services":[
               {
                  "id":"11",
                  "local_id":"11_l",
                  "name":"11_name",
                  "aspects":[
                     {
                        "id":"a1",
                        "name":"",
                        "rdf_type":""
                     }
                  ],
                  "protocol_id":"pid",
                  "functions":[
                     {
                        "id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
                        "name":"",
                        "concept_id":"",
                        "rdf_type":"https://senergy.infai.org/ontology/MeasuringFunction"
                     }
                  ]
               }
            ]
         }
      ]
   },
   {
      "id":"3",
      "selectables":null
   }
]
```


## Bulk Request Combined Devices

similar to a request to '/bulk/selectables' but returns only the found (distinct) devices of all found selectables.

**request:**
```
POST /bulk/selectables/combined/devices
[
   {
      "id":"1",
      "filter_interaction":null,
      "filter_protocols":[
         "mqtt"
      ],
      "criteria":[
         {
            "function_id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
            "device_class_id":"dc1",
            "aspect_id":"a1"
         }
      ]
   },
   {
      "id":"2",
      "filter_interaction":"event",
      "filter_protocols":null,
      "criteria":[
         {
            "function_id":"https://senergy.infai.org/ontology/MeasuringFunction_1",
            "device_class_id":"dc1",
            "aspect_id":"a1"
         }
      ]
   }
]
```


**response:**
```
[
   {
      "id":"1",
      "name":"1",
      "device_type_id":"dt1",
      "permissions":{
         "r":true,
         "w":false,
         "x":true,
         "a":false
      },
      "shared":false,
      "creator":""
   }
]
```

## Completed Services

by default the '/selectables' and '/bulk/selectables' endpoints return the services as known by the semantic repository. For completed services the query-parameter 'complete_services' can be set to true. In this case the additional field servicePathOptions is returned for each selectable.

**examples:**
```
GET /selectables?complete_services=true&function_id=someId1&aspect_id=someOtherId
```

```
POST /bulk/selectables/combined/devices?complete_services=true
...
```
