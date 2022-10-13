# Mail Sender

This service is only responsible for validating sending requests and interacting to the AWS/SES to send email requests, it also publishes 
success or failures events regarding the sending attempts. 

Note that a success event only means the email has been queued for sending successfully, not that it arrived in the repicient(s) inbox, 
for that see the mail events microservice

---

## Rabbitmq

This services consumes a single queue defined by the `RMQ_QUEUE` env var (defaults to `mail_requests`) where it expects the following message
body: 

```json
{
    "uuid": "2221e2de-7385-433a-ac63-21ce013a6436",
    "to": ["bruce.wayne@gmail.com"],
    "cc": [],
    "bcc": [],
    "reply_to_addresses": [],
    "subject_text": "you got mail",
    "body_html": "<h1>hello !</h1>",
    "body_text": "hello !"
}
```

if the amqp delivery `correlation id` and `reply to` properties are set feedback regarding the send operation
will be send to the queue on the `reply to` property, with the same `correlation id`, a feedback has the following
body.

```json
{
    "success": true,                     // false if error
    "message": "email sent successfully" // if success is false, this will be the error description
}
```

---

## Configuration

configuration is set by a yml config file and enviroment variables, each variable on the yaml file can be overwrittern by a env var,
check `config/config.yml` for details.

when developing its easier to use the yml equivalent of those variables on your `config/config.dev.yml` file and running the service
with `make run_dev` or `go run cmd/main.go --config-file="./config/config.dev.yml"`

