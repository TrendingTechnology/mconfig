## deploy_singleton


### Prepare

* mconfig
* mconfig-cli
* mconfig-go-sdk


### Deploy

1 start mconfig server

```shell
./mconfig-server  --registry=false --store_type=file
```

2 init mconfig data

such you want build an app named BookStore, you can...

```shell
./mconfig-server-cli init BookStore -t direct -r host:ip
```

3 edit the config, such as

> ./BookStore/config.json

```json
{"db":{"url":"127.0.0.1:3306","database":"bookstore","time_out": 20}}
```

> ./BookStore/schema.json

```json
{
    "type": "object",
    "properties": {
        "db": {
            "type": "object",
            "properties": {
               "url": {
                  "type": "string"
               },
               "database": {
                  "type": "string"
               },    
			   "time_out": {
                  "type": "integer"
               }           
            }
        }
    }
}
```

3 publish config to mconfig

```shell
./mconfig-server-cli publish  -c ./BookStore/config.json -s ./BookStore/schema.json  --app  BookStore  --config database -t direct -r  host:ip
```

4 use mconfig sdk to get config data

```go
	config := client.NewMconfig(
		client.DirectLinkAddress("127.0.0.1:8080"),
		client.AppKey("BookStore"),
		client.ConfigKey("database"),
		client.RetryTime(15 * time.Second),
	)
	url := config.String("db.url")
	db := config.String("db.database")
	timeout := config.Int("time_out")
	
```