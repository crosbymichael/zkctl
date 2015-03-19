# zkctl - zookeeper cli client

`zkctl` is a easy to use cli to zookeeper that allows you to 
create, list, get, set, and delete nodes within zookeeper.


```bash
NAME:
   zkctl - cli appliction for zookeeper

USAGE:
   zkctl [global options] command [command options] [arguments...]

VERSION:
   1

AUTHOR:
  @crosbymichael - <crosbymichael@gmail.com>

COMMANDS:
   create   create a node
   set      set a key's value
   get      get a key's value
   ls       list nodes
   delete   delete a key
   help, h  Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --debug, -D              enable debug output
   -t, --timeout "2s"           connection timeout
   --zk [--zk option --zk option]   zookeeper node addresses
   --help, -h               show help
   --version, -v            print the version
   
```


### Usage:

#### Create a new node

```bash
NAME:
   create - create a node

USAGE:
   command create [command options] [arguments...]

OPTIONS:
   -v, --version "0"                node version
   --world, -w                  set world ACL
   --acl, -a [--acl option --acl option]    ACL used on create. default is all. all|create|write|read|delete|admin
   

zkctl create /name koye
```

#### Get a node's value
```
NAME:
   get - get a key's value

USAGE:
   command get [arguments...]


zkctl get /name
koye
```

#### Set the value for an exist node

Automatically get the node's version for ease of use in changing the value by using 
the `--auto` flag.  You can also use the `-v` flag to set the version to a specific
value.

```
NAME:
   set - set a key's value

USAGE:
   command set [command options] [arguments...]

OPTIONS:
   -v, --version "0"    node version
   --auto       automatically set the version based on the current node
   
zkctl set --auto /name mike
```

#### List nodes

```
NAME:
   ls - list nodes

USAGE:
   command ls [arguments...]


zkctl ls /
/mesos
/name
/zookeeper
/marathon
```

#### Delete a node

`delete` also accepts `--auto` and `-v` for version handling.

```
NAME:
   delete - delete a key

USAGE:
   command delete [command options] [arguments...]

OPTIONS:
   -v, --version "0"    node version
   --auto       automatically set the version based on the current node
   --recursive, -r  recursively delete keys within the path
   

zkctl delete --auto /name
```
