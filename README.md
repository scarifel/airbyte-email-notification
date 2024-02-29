## Airbyte Email Notification 

Build
```
docker build -t test-notification . 
```

Run
```
docker run -d -p 8080:8080 --env-file=.env --name=test-mail test-notification 
```