## Getting Started
### With docker
#### Run dev with watch files and auto restart
```sh
make dev
```

#### Run production
```sh
make run
```

### Without docker
Edit config.ini (dev)
```ini
[Database]
User = root
Password = secret
Host = localhost
Port = 3306
Name = orderdb
```

Run
```sh
go get
air
```