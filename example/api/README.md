# Sending data to API

To send data to the API, you can use the following commands:

```console
export KONSUMO_USR="YourAccessKey"
export KONSUMO_PWD="YourSecretKey"

$ curl -k -X POST -H 'Content-Type: application/json' -d @example.json https://$KONSUMO_USR:$KONSUMO_PWD@127.0.0.1:8080/konsumo/api/add/water

$ curl -k -X POST -H 'Content-Type: application/json' -d '{ "date":"2023-03-16", "value1": 796 }' https://$KONSUMO_USR:$KONSUMO_PWD@127.0.0.1:8080/konsumo/api/add/electricity


$ curl -k -X POST -H 'Content-Type: application/json' -d @bundle.json https://$KONSUMO_USR:$KONSUMO_PWD@127.0.0.1:8080/konsumo/api/addbundle/gazoline
```