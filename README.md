# Install Go
```
sudo rm -fr /usr/local/go
Download & Install MacOS pkg from https://golang.org/dl/
export PATH=$PATH:/usr/local/go/bin
```

# Sharding Config
sharding.yml
```
datasources:
  - root:root@tcp(127.0.0.1:3306)/my_db_0
  - root:root@tcp(127.0.0.1:3306)/my_db_1
  - root:root@tcp(127.0.0.1:3306)/my_db_2
  - root:root@tcp(127.0.0.1:3306)/my_db_3

table_number: 256
sharding_column: uid
```

# Compile
```

```

# Run Test
```

```