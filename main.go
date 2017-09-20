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
	connectArgs.StringVar(&connector.src, "s", "", "The host and port of the local server")
	connectArgs.StringVar(&connector.dst, "c", "", "The host and port of the Listen/Listen server")

	listener := &Listener{}
	listenArgs := flag.NewFlagSet("listen", flag.ExitOnError)
	listenArgs.StringVar(&listener.from, "p", "", "The port to listen on for actual client connection")
	listenArgs.StringVar(&listener.to, "q", "", "The port to listen on for connection from the other Connect/Connect node")
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
		return listener, nil, nil
	} else if connectMode {
		connectArgs.Parse(os.Args[2:])
		if connector.src == "" || connector.dst == "" {
			return nil, nil, fmt.Errorf("no connect options")
		}
		return nil, connector, nil
	}
	return nil, nil, fmt.Errorf("too many options")
}

func main() {
	listener, connector, err := parseArgs()
	if err != nil {
		flag.Usage()
	}
	if listener != nil {
		if err := listener.Listen(); err != nil {
			log.Fatalln(err)
		}
	} else if connector != nil {
		connector.Connect()
	}
}
