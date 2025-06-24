Redix
=====
> `redix` is a key-value datastore with pluggable storage engines and redis protocol as interface, [documentation & learn more](https://redix.alash3al.com/)

Contributions
=============
> You're welcome!


# Run in docker
```bash
docker run -d -v $(pwd)/redix.hcl:/etc/redix/redix.hcl --link redixdb -p 6380:6380 refrag/redix
```
