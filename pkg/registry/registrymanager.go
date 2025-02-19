package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/containers/common/pkg/auth"
	"github.com/containers/image/v5/pkg/sysregistriesv2"
	"github.com/containers/image/v5/types"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type Endpoint struct {
	// The endpoint's remote location. Can be empty iff Prefix contains
	// wildcard in the format: "*.example.com" for subdomain matching.
	// Please refer to FindRegistry / PullSourcesFromReference instead
	// of accessing/interpreting `Location` directly.
	Location string `toml:"location,omitempty"`
	// If true, certs verification will be skipped and HTTP (non-TLS)
	// connections will be allowed.
	Insecure bool `toml:"insecure,omitempty"`
}

type Registry struct {
	// Prefix is used for matching images, and to translate one namespace to
	// another.  If `Prefix="example.com/bar"`, `location="example.com/foo/bar"`
	// and we pull from "example.com/bar/myimage:latest", the image will
	// effectively be pulled from "example.com/foo/bar/myimage:latest".
	// If no Prefix is specified, it defaults to the specified location.
	// Prefix can also be in the format: "*.example.com" for matching
	// subdomains. The wildcard should only be in the beginning and should also
	// not contain any namespaces or special characters: "/", "@" or ":".
	// Please refer to FindRegistry / PullSourcesFromReference instead
	// of accessing/interpreting `Prefix` directly.
	Prefix string
	// A registry is an Endpoint too
	Endpoint
}

type registryManagerOpts struct {
	RegistriesDirPath        string
	SystemRegistriesConfPath string
}

type RegistryManager struct {
	opts registryManagerOpts
}

type RegistryManagerOption func(opts *registryManagerOpts) error

func WithRegistryManagerRegistriesDirPath(registriesDirPath string) RegistryManagerOption {
	return func(opts *registryManagerOpts) error {
		opts.RegistriesDirPath = registriesDirPath

		return nil
	}
}

func WithRegistryManagerSystemRegistriesConfPath(systemRegistriesConfPath string) RegistryManagerOption {
	return func(opts *registryManagerOpts) error {
		opts.SystemRegistriesConfPath = systemRegistriesConfPath

		return nil
	}
}

type (
	// LoginOption allows specifying various settings on login
	LoginOption func(*loginOperation) error

	loginOperation struct {
		Credential            orasauth.Credential
		PlainHTTP             bool
		TLSInsecureSkipVerify bool
	}
)

// LoginOptBasicAuth returns a function that sets the username/password settings on login
func LoginOptBasicAuth(username string, password string) LoginOption {
	return func(o *loginOperation) error {
		o.Credential = orasauth.Credential{Username: username, Password: password}

		return nil
	}
}

// LoginOptPlainText returns a function that allows plaintext (HTTP) login
func LoginOptPlainText(isPlainText bool) LoginOption {
	return func(o *loginOperation) error {
		o.PlainHTTP = isPlainText

		return nil
	}
}

// LoginOptInsecure returns a function that sets the insecure setting on login
func LoginOptInsecure(insecure bool) LoginOption {
	return func(o *loginOperation) error {
		o.TLSInsecureSkipVerify = insecure

		return nil
	}
}

// LoginOptTLSClientConfig returns a function that sets the TLS settings on login.
func LoginOptTLSClientConfig(certFile, keyFile, caFile string) LoginOption {
	return func(o *loginOperation) error {
		if (certFile == "" || keyFile == "") && caFile == "" {
			return nil
		}

		tlsConfig, err := ensureTLSConfig(o.client.authorizer)
		if err != nil {
			panic(err)
		}

		if certFile != "" && keyFile != "" {
			authCert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				panic(err)
			}
			tlsConfig.Certificates = []tls.Certificate{authCert}
		}

		if caFile != "" {
			certPool := x509.NewCertPool()
			ca, err := os.ReadFile(caFile)
			if err != nil {
				panic(err)
			}
			if !certPool.AppendCertsFromPEM(ca) {
				panic(fmt.Errorf("unable to parse CA file: %q", caFile))
			}
			tlsConfig.RootCAs = certPool
		}

		return nil
	}
}

type (
	// LogoutOption allows specifying various settings on logout
	LogoutOption func(*logoutOperation)

	logoutOperation struct{}
)

func NewRegistryManager(options ...RegistryManagerOption) (*RegistryManager, error) {

	opts := registryManagerOpts{}

	errs := []error{}
	for _, o := range options {
		err := o(&opts)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("failed to create RegistryManager: %w", err)
	}

	result := &RegistryManager{
		opts: opts,
	}

	return result, nil
}

func (r *RegistryManager) sysContext() *types.SystemContext {
	return &types.SystemContext{
		RegistriesDirPath:        r.opts.RegistriesDirPath,
		SystemRegistriesConfPath: r.opts.SystemRegistriesConfPath,
	}
}

func (r *RegistryManager) Login(registryPath string, opts ...LoginOption) error {

	operation := loginOperation{}
	for _, option := range opts {
		option(&operation)
	}

	sys := r.sysContext()

	loginOpts := &auth.LoginOptions{
		AcceptRepositories:        false,
		Stdin:                     os.Stdin,
		Stdout:                    os.Stdout,
		AcceptUnspecifiedRegistry: false,
		NoWriteBack:               false,
	}
	args := []string{registryPath}

	err := auth.Login(context.Background(), sys, loginOpts, args)

	return err
}

func (r *RegistryManager) Logout(registryPath string, opts ...LogoutOption) error {

	operation := logoutOperation{}
	for _, opt := range opts {
		opt(&operation)
	}

	sys := r.sysContext()

	opts := &auth.LogoutOptions{
		AcceptRepositories:        false,
		Stdout:                    os.Stdout,
		AcceptUnspecifiedRegistry: false,
	}
	args := []string{registryPath}

	err := auth.Logout(sys, opts, args)

	return err
}

func (r *RegistryManager) ReadRegistryCredentials(refPath string) (*Registry, error) {

	sys := r.sysContext()

	registry, err := sysregistriesv2.FindRegistry(sys, refPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find registry: %w", err)
	}

	result := &Registry{
		Prefix: registry.Prefix,
		Endpoint: Endpoint{
			Location: registry.Location,
			Insecure: registry.Insecure,
		},
	}

	return result, nil
}

func ensureTLSConfig(client *auth.Client) (*tls.Config, error) {
	var transport *http.Transport

	switch t := client.Client.Transport.(type) {
	case *http.Transport:
		transport = t
	case *retry.Transport:
		switch t := t.Base.(type) {
		case *http.Transport:
			transport = t
		}
	}

	if transport == nil {
		// we don't know how to access the http.Transport, most likely the
		// auth.Client.Client was provided by API user
		return nil, fmt.Errorf("unable to access TLS client configuration, the provided HTTP Transport is not supported, given: %T", client.Client.Transport)
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}

	return transport.TLSClientConfig, nil
}
