# LogMe 🗄️

Application to receive log entries and to make them queryable

## Requirements 🐳

- [Docker](https://docs.docker.com/engine/install/)
- [Make](https://www.tutorialspoint.com/unix_commands/make.htm) is installed (usually via `apt install build-essential`)

## Running the Application 🚀

Create an empty file named `go.sum`

Copy the `.env.docker` to `.env`

Now run the following to bring up the application

```bash
make start
```

If successful you should be able to visit [http://localhost:8080](http://localhost:8080) and you should see the following:

```json
{"message":"Welcome to LogMe!"}
```

### Migrating the Databases

Use the built-in tool to migrate the database!

```bash
./logme-cli m
```

And migrate the test database 🐞

```bash
./logme-cli mt
```

### Currently Available Functionality


#### Add a log entry

```
POST /log/{accountId}
```

Example requests

```bash
curl -X POST -i -H "Content-Type: application/json" http://localhost:8080/log/321 \
     -d "{\"name\":\"error.log\", \"timestamp\":\"2022-12-31 12:34:56\", \"content\":\"this is another log entry\", \"account_id\":321}"
HTTP/1.1 201 Created
Content-Type: application/json
Date: Tue, 31 May 2022 12:14:18 GMT
Content-Length: 42

{"message":"Log successfully processed."}
```

3 values are required for the API call above:

- `name` the name of the log file
- `timestamp` when the entry occurred
- `content` the content of the log entry

#### List all logs

```
GET /log/{accountId}
```

Example request

```bash
curl -i -H "Content-Type: application/json" http://localhost:8080/log/321
HTTP/1.1 200 OK
Date: Tue, 31 May 2022 12:14:54 GMT
Content-Length: 152
Content-Type: text/plain; charset=utf-8

[{"uuid":"365aecd5-6bb5-4061-a941-50e2f99f9eaa","name":"error.log","account_id":321,"dt":"2022-12-31T12:34:56Z","content":"this is another log entry"}]
```

## Running tests 🧪

To run test, use the following command

```bash
make test
```

## TODO

- Add test database clean up
- Add account routes (create, update, etc.)
- Add route authentication
- Add concurrency to request handling
- Add log search
- Add meta data tags to enable better searches
