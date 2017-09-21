# TCP Gender Changer

TCP Gender Changer is a small utility to connect/connect on a local service to a remote server using a listen/listen setup. 
More information on [wikipedia](https://en.wikipedia.org/wiki/TCP_Gender_Changer).

This is heaviliy inspired by http://tgcd.sourceforge.net

## Installation

```bash
$ go get -u github.com/jfyuen/tgc
```

## Usage

On the node on which to access a server on a LAN, run the connect/connect setup:
```bash
$ tgc -C -s ${LOCAL_ADDRESS} -c ${REMOTE_ADDRESS}
```

On the remote node on which you can connect to, run the listen/listen setup:
```bash
$ tgc -L -q ${EXTERNAL_PORT} -p ${LOCAL_PORT}
```

The remote node now has access to the local server on the LAN via `${LOCAL_PORT}`

Additionally, a secure connection can be established between the Connect/Connect node and the Listen/Listen node by using the following options for both modes:
- `ca`: CA of the connecting party, on the server side, provide the client CA; on the client side, provide the server CA. They can be the same CA
- `crt`: x509 cert file signed by the CA provided to the remote node
- `key`: key file

### Example

Forward local port 8080 to a remote server on port 8000 via port 80. 
On the local server:
```bash
$ tgc -C -s 127.0.0.1:8080 -c remote:80
```

On the remote server:
```bash
$ tgc -L -q 80 -p 127.0.0.1:8000
```