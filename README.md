# IAM

This service periodically collects GCP IAM roles and their permissions and stores them in a local DB. Roles and/or Permissions can be queried via the api.


## API

To retrieve details about a particular role: 

```shell
curl --location --request GET 'v1/role/named' \
--header 'Content-Type: application/json' \
--data-raw '{
    "named": "role/name"
}'
```

To retrieve all roles with the provided permissions:

```shell
curl --location --request GET 'v1/role/permissions' \
--header 'Content-Type: application/json' \
--data-raw '{
    "permissions": ["permission_1", "permissions_2"]
}'
```