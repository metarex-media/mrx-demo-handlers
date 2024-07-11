# mrx-demo-handlers

A demo using the [MetaRex register][1] as a no/low code solution of processing
metadata.

This repo provides the processing and the mrx-demo-svc provides the web
service interface to invoke these processing functions.

This is work in progress as we find a best-practice way to publish and invoke
metadata services in a low-effort media workflow environment.

The current API transformations are listed on [swaggerhub][2]. Complete with a
schema of expected metadata inputs and outputs.

## The MetaRex Register / How does it work

Each MetaRex file, comes with a ID for each type of data it contains. This ID
is unique for each data type, the properties can than be accessed by the
globally available MetaRex register. Data can then be processed using the APIs
available in the register, be known to the system or discarded.

## steps to take

### Build and Run the API

The API can be built with the following commands

```cmd
cd api
go build
./api
```

Check out the OpenAPI [specification](./mrxhandle/openAPI.yaml)
for the API.

### Building and running the MRX handler

This section will be updated. In reality its most likely to be used as a data
handler for users to input their own metadata and play with.

This section will be more about running it as a command line. For the sake of
this demo all the data output will be in JSON Format.

### THE MRX object for logging

A recursive body of the MRX transformation history can be built. The ID of
which is saved as part of the logging, this id is chain of parent to child of
the MetaRex IDs, with the origin of the oldest mrx as well. This is then hashed
using xxh64 hash, with the purpose of preserving Ids over different functions
where you don't need to pass all the mrx details.

[1]: https://metarex.media/ui/reg/
[2]: https://app.swaggerhub.com/apis/TRISTAN_9/MetarexAPIDemo/0.1.0#/default/post_3dTransform