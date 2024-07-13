#!/bin/bash
appbox network add --name appbox0 --type bridge --ip 172.17.10.1
mkdir /etc/appbox/
cat > /etc/appbox/appbox-net.json <<EOF
{
  "IP":"172.17.10.1",
  "Name":"appbox0",
  "Type":"bridge",
  "BrdAttr":{
    "Name":"appbox0",
    "IP":"172.17.10.1"
  }
}
EOF