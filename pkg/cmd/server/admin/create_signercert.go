package admin

import (
	"errors"
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kcmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"

	"github.com/openshift/origin/pkg/cmd/server/crypto"
)

const CreateSignerCertCommandName = "create-signer-cert"

type CreateSignerCertOptions struct {
	CertFile   string
	KeyFile    string
	SerialFile string
	Name       string
	Output     cmdutil.Output

	Overwrite bool
}

func BindSignerCertOptions(options *CreateSignerCertOptions, flags *pflag.FlagSet, prefix string) {
	flags.StringVar(&options.CertFile, prefix+"cert", "openshift.local.config/master/ca.crt", "The certificate file.")
	flags.StringVar(&options.KeyFile, prefix+"key", "openshift.local.config/master/ca.key", "The key file.")
	flags.StringVar(&options.SerialFile, prefix+"serial", "openshift.local.config/master/ca.serial.txt", "The serial file that keeps track of how many certs have been signed.")
	flags.StringVar(&options.Name, prefix+"name", DefaultSignerName(), "The name of the signer.")
	flags.BoolVar(&options.Overwrite, prefix+"overwrite", options.Overwrite, "Overwrite existing cert files if found.  If false, any existing file will be left as-is.")
}

const create_signer_long = `
Create a self-signed CA key/cert for signing certificates used by
OpenShift components.

This is mainly intended for development/trial deployments as production
deployments of OpenShift should utilize properly signed certificates
(generated separately) or start with a properly signed CA.
`

func NewCommandCreateSignerCert(commandName string, fullName string, out io.Writer) *cobra.Command {
	options := &CreateSignerCertOptions{Overwrite: true, Output: cmdutil.Output{out}}

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a signer (certificate authority/CA) certificate",
		Long:  create_signer_long,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if _, err := options.CreateSignerCert(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.SetOutput(out)

	BindSignerCertOptions(options, cmd.Flags(), "")

	return cmd
}

func (o CreateSignerCertOptions) Validate(args []string) error {
	if len(args) != 0 {
		return errors.New("no arguments are supported")
	}
	if len(o.CertFile) == 0 {
		return errors.New("cert must be provided")
	}
	if len(o.KeyFile) == 0 {
		return errors.New("key must be provided")
	}
	if len(o.SerialFile) == 0 {
		return errors.New("serial must be provided")
	}
	if len(o.Name) == 0 {
		return errors.New("name must be provided")
	}

	return nil
}

func (o CreateSignerCertOptions) CreateSignerCert() (*crypto.CA, error) {
	glog.V(2).Infof("Creating a signer cert with: %#v", o)
	var ca *crypto.CA
	var err error
	written := true
	if o.Overwrite {
		ca, err = crypto.MakeCA(o.CertFile, o.KeyFile, o.SerialFile, o.Name)
	} else {
		ca, written, err = crypto.EnsureCA(o.CertFile, o.KeyFile, o.SerialFile, o.Name)
	}
	if written {
		fmt.Fprintf(o.Output.Get(), "Generated new CA for %s: cert in %s and key in %s\n", o.Name, o.CertFile, o.KeyFile)
	} else {
		fmt.Fprintf(o.Output.Get(), "Keeping existing CA cert at %s and key at %s\n", o.CertFile, o.KeyFile)
	}
	return ca, err
}
