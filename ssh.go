package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	Client                *ssh.Client
	Connected             *atomic.Bool
	ControlTerminal       *ssh.Session
	ControlTerminalStdin  *io.WriteCloser
	ControlTerminalActive *atomic.Bool
	Host                  string
	Port                  string
	Username              string
	Password              string
	UseSSHKey             bool
	SSHKeyPath            string
	ClientIpAddress       string
	Secret                string
	ForwardedConnections  map[uint16]*net.Conn
	Timeout               time.Duration
}

func (sshConnection *SSHConnection) StartSSHConnection() error {
	// get host public key
	hostKey := getHostKey(sshConnection.Host)

	// ssh client config
	config := &ssh.ClientConfig{
		User: sshConnection.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshConnection.Password),
		},
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
		Timeout: sshConnection.Timeout,
	}

	if sshConnection.UseSSHKey {
		authMethod, merr := GetSSHPrivateKeyAuthMethod(sshConnection.SSHKeyPath)
		if merr == nil {
			config.Auth = append(config.Auth, authMethod)
		} else {
			log.Println("couldn't enable SSH key auth ", merr.Error())
		}
	}

	// connect
	client, err := ssh.Dial("tcp", sshConnection.Host+":"+sshConnection.Port, config)
	if err != nil {
		log.Println(err)
		return err
	}

	sshConnection.Client = client
	controlTerminal, cterr := client.NewSession()
	if cterr != nil {
		client.Close()
		return cterr
	}

	controlTerminal.Setenv("SECRET", sshConnection.Secret)
	controlTerminal.Setenv("REMOTE_ADDR", sshConnection.ClientIpAddress)
	sshConnection.ControlTerminal = controlTerminal
	sshConnection.Connected.Store(true)
	go sshConnection.SetupControlTerminal()
	return nil
}

func (sshConnection *SSHConnection) StartTerminal() (Session *ssh.Session,
	StdIn *io.WriteCloser, StdOut *io.Reader, StdErr *io.Reader, Error error) {
	if sshConnection.Connected.Load() {

		sess, err := sshConnection.Client.NewSession()
		if err != nil {
			return nil, nil, nil, nil, err
		}

		sess.RequestPty("xterm-256color", 80, 80, ssh.TerminalModes{
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
			ssh.ECHO:          0,
			ssh.ECHOCTL:       0,
		})

		st1err := sess.Setenv("SECRET", sshConnection.Secret)
		if st1err != nil {
			log.Println(st1err, "SECRET")
		}
		st2err := sess.Setenv("REMOTE_ADDR", sshConnection.ClientIpAddress)
		if st2err != nil {
			log.Println(st2err, "REMOTE_ADDR")
		}

		stdin, sierr := sess.StdinPipe()
		if sierr != nil {
			return nil, nil, nil, nil, sierr
		}
		stdout, soerr := sess.StdoutPipe()
		if soerr != nil {
			return nil, nil, nil, nil, soerr
		}
		stderr, seerr := sess.StderrPipe()
		if seerr != nil {
			return nil, nil, nil, nil, seerr
		}

		serr := sess.Shell()

		if serr != nil {
			sess.Close()
			return nil, nil, nil, nil, serr
		}

		return sess, &stdin, &stdout, &stderr, nil
	}
	return nil, nil, nil, nil, errors.New("connection is not active yet")
}

func (sshConnection *SSHConnection) StopSSHConnection() error {
	if sshConnection.Connected.Load() {
		for _, forwardedConn := range sshConnection.ForwardedConnections {
			(*forwardedConn).Close()
		}
		sshConnection.Client.Close()
		return sshConnection.Client.Wait()
	}
	return errors.New("connection is not active")
}

func (sshConnection *SSHConnection) WaitForConnection(tries int, checkDelay time.Duration) error {
	for tries > 0 {
		if sshConnection.Connected.Load() {
			return nil
		}
		tries -= 1
		time.Sleep(checkDelay)
	}

	return errors.New("SSH Connection timeout")
}

func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file
	// ssh or use ssh-keyscan to get initial key
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		log.Fatalf("no hostkey found for %s", host)
	}

	return hostKey
}

func (sshConnection *SSHConnection) SetupControlTerminal() {
	err := sshConnection.ControlTerminal.RequestPty("xterm-256color", 80, 80, ssh.TerminalModes{
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		ssh.ECHO:          0,
		ssh.ECHOCTL:       0,
	})
	if err != nil {
		log.Println(err)
		return
	}

	stdin, err := sshConnection.ControlTerminal.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}

	err = sshConnection.ControlTerminal.Shell()
	if err != nil {
		log.Println(err)
		return
	}

	sshConnection.ControlTerminalStdin = &stdin
	sshConnection.ControlTerminalActive.Store(true)
}

func (sshConnection *SSHConnection) RunControlCommand(command string) error {
	if sshConnection.ControlTerminalActive.Load() {
		stdin := *sshConnection.ControlTerminalStdin
		n, err := stdin.Write(append([]byte(command), 10, 13)) // append /n/c to the end
		if err != nil {
			log.Println(err)
			return err
		}
		if n < len(command) {
			log.Println("short write, exec failed")
			return errors.New("short write, exec failed")
		}
		return nil
	}

	return errors.New("control terminal not active")
}

func (sshConnection *SSHConnection) ForwardRemotePort(port uint16, cache bool) (*net.Conn, error) {
	conn, err := sshConnection.Client.DialTCP("tcp4", nil, net.TCPAddrFromAddrPort(
		netip.AddrPortFrom(
			netip.AddrFrom4(
				[4]byte{127, 0, 0, 1},
			),
			port,
		),
	))
	return &conn, err
}

func GetSSHPrivateKeyAuthMethod(keyFilePath string) (ssh.AuthMethod, error) {
	filebytes, ferr := os.ReadFile(keyFilePath)
	if ferr != nil {
		return nil, ferr
	}

	signer, serr := ssh.ParsePrivateKey(filebytes)
	if serr != nil {
		return nil, serr
	}

	return ssh.PublicKeys(signer), nil
}
