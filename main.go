package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func parseArgs() (*Listener, *Connector, error) {
	var connectMode, listenMode bool
	modeArgs := flag.NewFlagSet("mode", flag.ExitOnError)

	modeArgs.BoolVar(&connectMode, "C", false, "connect mode")
	modeArgs.BoolVar(&listenMode, "L", false, "listen mode")

	connector := &Connector{}
	connectArgs := flag.NewFlagSet("connect", flag.ExitOnError)
	connectArgs.StringVar(&connector.src, "s", "", "the host and port of the local server")
	connectArgs.StringVar(&connector.dst, "c", "", "the host and port of the Listen/Listen server")
	connectArgs.IntVar(&connector.interval, "i", 5, "interval when (re)connecting to either host in seconds, must be positive")

	clientCert := TLSCert{}
	connectArgs.StringVar(&clientCert.ca, "ca", "", "certificate authority (used to signed the server cert) to encrypt the connection to the Listen/Listen server, connection will be unencrypted if not specified")
	connectArgs.StringVar(&clientCert.crt, "crt", "", "client certificate signed by the CA provided to the server")
	connectArgs.StringVar(&clientCert.key, "key", "", "client key file")

	listener := &Listener{}
	listenArgs := flag.NewFlagSet("listen", flag.ExitOnError)
	listenArgs.StringVar(&listener.from, "p", "", "the port to listen on for actual client connection")
	listenArgs.StringVar(&listener.to, "q", "", "the port to listen on for connection from the other Connect/Connect node")

	serverCert := TLSCert{}
	listenArgs.StringVar(&serverCert.ca, "ca", "", "certificate authority (used to sign the client cert) to encrypt the connection to the Connect/Connect client, connection will be unencrypted if not specified")
	listenArgs.StringVar(&serverCert.crt, "crt", "", "server certificate signed by the CA provided to the client")
	listenArgs.StringVar(&serverCert.key, "key", "", "server key file")

	flag.Usage = func() {
		modeArgs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "For connect mode:\n")
		connectArgs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "For listen mode:\n")
		listenArgs.PrintDefaults()

	}

	modeArgs.Parse(os.Args[1:2])
	if !connectMode && !listenMode {
		return nil, nil, fmt.Errorf("no options")
	}

	if listenMode {
		listenArgs.Parse(os.Args[2:])
		if listener.from == "" || listener.to == "" {
			return nil, nil, fmt.Errorf("no listen options")
		}
		var err error
		if listener.tlsConfig, err = serverCert.CreateServerConfig(); err != nil {
			return nil, nil, err
		}
		return listener, nil, nil
	} else if connectMode {
		connectArgs.Parse(os.Args[2:])
		if connector.src == "" || connector.dst == "" {
			return nil, nil, fmt.Errorf("no connect options")
		}
		if connector.interval <= 0 {
			return nil, nil, fmt.Errorf("delay must be a positive integer")
		}
		if clientCert.ca != "" {
			var err error
			if connector.tlsConfig, err = clientCert.CreateClientConfig(); err != nil {
				return nil, nil, err
			}
		}

		return nil, connector, nil
	}
	return nil, nil, fmt.Errorf("too many options")
}

func main() {
	listener, connector, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("%v\n", err))
		flag.Usage()
		os.Exit(2)
	}
	if listener != nil {
		if err := listener.Listen(); err != nil {
			log.Fatalln(err)
		}
	} else if connector != nil {
		connector.Connect()
	}
}
