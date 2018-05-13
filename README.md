# Mantis webhook
Webhook integration Gitlab -> Mantisbt

This webhook aims to integrate Gitlab with Mantisbt by closing issues listed on a commit message.
It also provides some nice features like email aliasing and project mapping.

# Getting started
All you need to run this webhook is Go (any version, but the newer the better) and a Postgresql database. You can download the prebuilt package with a docker configuration to quickly get it running.

This webhook use both soap and REST mantis apis, so make sure both of them are enabled before you continue.

# Configuration
This is an example of `config.json` you need to create on the root directory on the webhook.
```json
{
  "port": 8090,
  "secret": "secret_token",
  "database": {
    "host": "localhost",
    "database_name": "go-mantis",
    "user": "postgres",
    "password": "postgres"
  },
  "gitlab": {
    "token": "my_gitlab_token"
  },
  "mantis": {
    "host": "http://localhost:8989",
    "user": "administrator",
    "password": "root",
    "token": "QAei2Pozy2FHd_5fFx2Bw3SH8obe4e49"
  }
}
```

# Endpoints

### Webhook
All requests under `/webhook` are authorized by the token you set on your Gitlab repository. The webhook will check for it under the request's `X-Gitlab-Token` header and compare it with the token you set on the `config.json` under `gitlab.token`.

#### POST /webhook/push
The endpoint you need to register on your [Gitlab repository](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html).

### Application
All requests under `/app` are authorized by the token you set on you `config.json` under `secret`. The app will check for the token under the request's `Authorization` header.

#### POST /app/projects
Associates a new Gitlab project with a Mantis one.

Params:
```json
{
  "gitlab_project": "test/test",
  "mantis_project": "Test project"
}
```

#### POST /app/aliases
Registers an email alias for a Mantis user.

Params:
```json
{
  "email": "my.mantis.email@example.com",
  "alias": "my.other.email@example.com"
}
```
