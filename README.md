### Load Simulator

- This repo contains code for simulating load on various mechanism by which a service can receive requests
- All the loads have few common parameters like
    - `ratePerSec` : The rate at which we want to simulate the load
    - `duration` : the duration for which we want to simulate the load
    - `concurrentRequests` : the number of workers across which total load will be distributed
    - `fileName` : except for `GET` request the path of the payload for posting
- For making HTTP Get & Post calls we assume an OAuth mechanism
- For posting the message to Kafka topics we assume either an OAuth mechanism or Scram mechanism. Scram `SHA-512`
- For generating an OAuth token we require 4 parameters:
    - `clientId`
    - `clientSecret`
    - `scope`
    - `OauthUrl` : We have used `TokenUrl` for HTTP calls and `OauthUrl` for Kafka OAuth. In case they are same we can
      keep the same value for both
- All the configs are kept under `assets/configs` folder and data which we want to post is kept under `data` folder
- The parameters which are specific for each type of load is mentioned below
- logs for individual scenarios are generated under `logs/` directory and app.log contains main load run log.

#### HTTP Get Calls

```yaml
getByPathVariable:
  method: "GET"
  baseUrl: "https://baseUrl.com/"
  endpoint: "api/v1/{id}"
  contentType: "application/json" # expected response content's type
  expectedStatusCode: 200
  pathVariables: # value of path variables if any
    - key: "id"
      value: AB12345
```

#### HTTP Post Calls

```yaml
postWithoutReplacement:
  method: "POST"
  baseUrl: "https://baseUrl.com/"
  endpoint: "api/v1/"
  contentType: "application/xml"
  expectedStatusCode: 202
  # the below is optional. If you have to ensure that there are certain fields which need to randomized or add some
  # custom mechanism for populating we can mention them here
  replaceParams:
    - key: "{{random}}"
      value: "UUID"
```

#### S3 Upload

```yaml
s3Upload:
  bucket: 'name_of_s3_bucket'
  key: 'foo/bar/abc/'
  extension: '_load.json' # this is added so to identify the payloads which were sent as load and can be deleted later
  region: 'us-east-1'
```

#### SQS Message

```yaml
sendToSqs:
  queue: 'name_of_the_queue'
  region: 'us-east-1'
  # message attributes and their value and types (string, number or binary)
  messageAttributes:
    - name: "orderNumber"
      value: "1234566"
      type: "string"
```

#### Kafka Producer

```yaml
kafkaOauth:
  topic: "name_of_the_topic"
  broker: "broker_url"
  authentication: "oauth"

kafkaScram:
  username: "user_name"
  password: "pass_word"
  topic: "topic_name"
  broker: "broker_url"
  authentication: "scram"
```

### Building and running

- Makefile has different commands to execute the respective scenarios
- For building one can use `make linux` or `make darwin` for arm64.