# LogMe 🗄️

Application to receive log entries and to make them queryable

## Requirements 🐳

- [Docker](https://docs.docker.com/engine/install/)

## Running the Application 🚀

Create 2 empty files `.env` and `go.sum`

Now run the following to bring up the application

```bash
docker-compose up
```

If successful you should be able to visit [http://localhost:8080](http://localhost:8080) and you should see the following:

```json
{"message":"Welcome to LogMe!"}
```

### Currently Available Functionality


#### Add a log entry

```
POST /log
```

Example requests

```bash
curl -X POST -i -H "Content-Type: application/json" http://localhost:8080/log \
     -d "{\"name\":\"error.log\", \"timestamp\":\"2022-12-31 12:34:56\", \"content\":\"this is a log entry\"}"
HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 26 May 2022 12:56:21 GMT
Content-Length: 42

{"message":"Log successfully processed."}
```

3 values are required for the API call above:

- `name` the name of the log file
- `timestamp` when the entry occurred
- `content` the content of the log entry

#### Read a log file

```
GET /log/{name}
```

Example request

```bash
curl -i -H "Content-Type: application/json" http://localhost:8080/log/error.log
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Thu, 26 May 2022 12:56:52 GMT
Content-Length: 20

this is a log entry
```

## Running tests 🧪

To run test, use the following command

```bash
go test
```
