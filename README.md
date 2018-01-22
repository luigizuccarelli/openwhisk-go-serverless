# openwhisk-go-serverless
An openwhisk go serverless function that reads and writes iot custom data from/to mongodb

## Development and Debug Tips

### Install and start openwhisk on local dev
* Refer to the openwhisk repo [ https://github.com/apache/incubator-openwhisk ]

### Download and start the mongodb docker container
```bash
# I made use of the jessie 3.6 image
docker run --name mongodb -p 27017:27017 -v /tmp/data:/data/db -d mongo:3.6.0-jessie

# use docker inspect to get the ip address

```

### Build the go file/s

```bash
# build go executable ensure libraries are statically linked (if not an error /action/exec can't be found)
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' exec.go

```

### Test it locally - repeat the code change and build steps until its all working
```bash
./exec '{"ip": "localhost","db": "test", "action": "insert", "payload": {"channels": [1,0,0,0,0,0,0,1],"temperatures":[40.3,40.5,42.5,45.8,50.1,50.2,53.5,54.5]}}'

# read the data from mongodb
./exec '{"ip": "localhost","db": "test", "action": "read", "payload": {"channels": [1,0,0,0,0,0,0,1],"temperatures":[40.3,40.5,42.5,45.8,50.1,50.2,53.5,54.5]}}'

```

### Build docker file (change it accordingly to match your docker hub account)
```bash
docker build -t lzuccarelli/go-serverless .

docker push lzuccarelli/go-serverless

```
### Start the serverless docker image just created
```bash
# run the docker file linking in the mongodb (where xxxxx is the newly created image)
docker run -p 8080:8080 --link mongodb -d xxxxx "/bin/bash" "-c" "cd actionProxy && python -u actionproxy.py"

```

### Test the docker image (use the invoke.py script)
```bash
python invoke.py init
OK

python invoke.py run '{"ip": "localhost","db": "test", "action": "read", "payload": {"channels": [1,0,0,0,0,0,0,1],"temperatures":[40.3,40.5,42.5,45.8,50.1,50.2,53.5,54.5]}}'

```

### Test using wsk command line
```bash
bin/wsk action invoke  lmz-test --result -p ip "172.17.0.17" -p db "test" -p action "read" --insecure --blocking 

```

## TODO
* Make use of envars for mongodb connection string rather than passing ip
* Stress test on an openwhisk cluster


