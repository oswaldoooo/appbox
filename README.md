# APPBOX

### **Use Example**
```shell
# run with redis box config file
appbox run -f redis-box.json
appbox run redis-server --path /etc/redis --path /var/log --static-path /home/user/.data 
```
### **Create Network Zone**
```shell
# create a network bridge at current namespace
appbox network add --name appbox1 --ip 172.20.10.1 --type bridge
# create on target namespace
appbox network add --name appbox1 --ip 172.20.10.1 --type bridge --pid 3424
```
### **Execute Command In Target Box Namespace**
```shell
appbox nsexec ip a -n --target 3343
appbox nsexec ls / -m --target 3343
```
### Base appbox on target network namespace
config json add `LinkNet` field. Link Namespace is Confilect on Standlone Namespace. Only choose one between them
```json
{
  "LinkNet":"3234"
}
```

## **Base Class Flag**
### Base flag like `--base-net-pid`,`--base-mnt-pid`. It's base on the target pid namespace target resource run. You can build multiply network

## **Init a existed box namespace to other box use**
```shell
apptool network init appbox1 --pid 3343 --ip 172.17.20.1/24
```

## **Boxd**
### Io Manager
io manager bind on tcp *:5678. make sure your host's tcp 5678 is not be used before

*box will bind their self stdout and stderr to boxd(if boxd is live),then boxd copy their stdout,stderr to local file. and it also will copy one to client(if there have client connect)*
![](io-manage.png)