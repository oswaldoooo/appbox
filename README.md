# APPBOX

### Use Example
```shell
# run with redis box config file
appbox run -f redis-box.json
appbox run redis-server --path /etc/redis --path /var/log --static-path /home/user/.data 
```
### Create Network Zone
```shell
# create a network bridge at current namespace
appbox network add --name appbox1 --ip 172.20.10.1 --type bridge
# create on target namespace
appbox network add --name appbox1 --ip 172.20.10.1 --type bridge --pid 3424
```
### Base appbox on target network namespace
config json add `LinkNet` field
```json
{
  "LinkNet":"3234"
}
```