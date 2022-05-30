# LogMe 🗄️

Application to receive log entries and to make them queryable

## Requirements 🐳

- [Docker](https://docs.docker.com/engine/install/)
- [Make](https://www.tutorialspoint.com/unix_commands/make.htm) is installed (usually via `apt install build-essential`)

## Running the Application 🚀

Create 2 empty files `.env` and `go.sum`

Now run the following to bring up the application

```bash
make start
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
     -d "{\"name\":\"error.log\", \"timestamp\":\"2022-12-31 12:34:56\", \"content\":\"this is a log entry\", \"account_id\":321}"
HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 30 May 2022 12:46:29 GMT
Content-Length: 42

{"message":"Log successfully processed."}
```

3 values are required for the API call above:

- `name` the name of the log file
- `timestamp` when the entry occurred
- `content` the content of the log entry

#### List all logs

```
GET /log
```

Example request

```bash
curl -i -H "Content-Type: application/json" http://localhost:8080/log
HTTP/1.1 200 OK
Date: Mon, 30 May 2022 12:56:25 GMT
Content-Length: 306
Content-Type: text/plain; charset=utf-8

[{"Uuid":"a76038ed-3a1f-4529-876a-72960f043b32","Name":"error.log","AccountId":321,"DateTime":"2022-12-31T12:34:56Z","Content":"this is a log entry"}]
```

## Running tests 🧪

To run test, use the following command

```bash
go test
```
