package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Loggers should be given as fields of struct and Pipe for better composability, but it is still ok like that
var debugLog *log.Logger = log.New(ioutil.Discard, "", log.Ldate|log.Ltime)
var errorLog *log.Logger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)

func newDebugLog() *log.Logger {
	return log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
}

func parseArgs() (*Listener, *Connector, error) {
	var connectMode, listenMode, logDebugConnect, logDebugListen bool
	modeArgs := flag.NewFlagSet("mode", flag.ExitOnError)

	modeArgs.BoolVar(&connectMode, "C", false, "connect mode")
	modeArgs.BoolVar(&listenMode, "L", false, "listen mode")

	connector := &Connector{}
	connectArgs := flag.NewFlagSet("connect", flag.ExitOnError)
	connectArgs.StringVar(&connector.src, "s", "", "the host and port of the local server")
	connectArgs.StringVar(&connector.dst, "c", "", "the host and port of the Listen/Listen server")
	connectArgs.IntVar(&connector.interval, "i", 5, "interval when (re)connecting to either host in seconds, must be positive")
	connectArgs.BoolVar(&logDebugConnect, "d", false, "activate debug logs")

	clientCert := TLSCert{}
	connectArgs.StringVar(&clientCert.ca, "ca", "", "certificate authority (used to signed the server cert) to encrypt the connection to the Listen/Listen server, connection will be unencrypted if not specified")
	connectArgs.StringVar(&clientCert.crt, "crt", "", "client certificate signed by the CA provided to the server")
	connectArgs.StringVar(&clientCert.key, "key", "", "client key file")

	listener := &Listener{}
	listenArgs := flag.NewFlagSet("listen", flag.ExitOnError)
	listenArgs.StringVar(&listener.from, "p", "", "the port to listen on for actual client connection")
	listenArgs.StringVar(&listener.to, "q", "", "the port to listen on for connection from the other Connect/Connect node")
	listenArgs.BoolVar(&logDebugListen, "d", false, "activate debug logs")

	serverCert := TLSCert{}
	listenArgs.StringVar(&serverCert.ca, "ca", "", "certificate authority (used to sign the client cert) to encrypt the connection to the Connect/Connect client, connection will be unencrypted if not specified")
	listenArgs.StringVar(&serverCert.crt, "crt", "", "server certificate signed by the CA provided to the client")
	listenArgs.StringVar(&serverCert.key, "key", "", "server key file")

	flag.Usage = func() {
		modeArgs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "For connect mode:")
		connectArgs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "For listen mode:")
		listenArgs.PrintDefaults()

	}
	if len(os.Args) >= 2 {
		modeArgs.Parse(os.Args[1:2])
	}
	if !connectMode && !listenMode {
		return nil, nil, fmt.Errorf("no options")
	}

	if listenMode {
		listenArgs.Parse(os.Args[2:])
		if listener.from == "" || listener.to == "" {
			return nil, nil, fmt.Errorf("no listen options")
		}
		var err error
		if serverCert.ca != "" {
			if listener.tlsConfig, err = serverCert.CreateServerConfig(); err != nil {
				return nil, nil, err
			}
		}
		if logDebugListen {
			debugLog = newDebugLog()
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
		if logDebugConnect {
			debugLog = newDebugLog()
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
