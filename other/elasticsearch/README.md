# Usage Metrics
Get usage metrics such as logins,logouts and new account creations. 
-   Configure a elasticsearch cluster and specify the credentials in `config.yaml` to get started.
-   Create a Index using the following elastic query
```
    PUT /sf_stats
    {
        "mappings": {
            "properties": {
                "Time": {
                    "type": "date"
                }
            }
        }
    }
```
-   Import the `dashboard.ndjson` file in kibana to get a overview of collected data.


