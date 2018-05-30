package cmd

import (
	"io"

	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/jenkins-x/jx/pkg/jx/cmd/log"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

// CreateClusterOptions the flags for running create cluster
type CreateClusterOKEOptions struct {
	CreateClusterOptions

	Flags CreateClusterOKEFlags
}

type CreateClusterOKEFlags struct {
	ClusterName             string
	CompartmentId           string
	VcnId                   string
	KubernetesVersion       string
	OKEOptions              string
	WaitForState            string
	KubernetesNetworkConfig string
	ServiceLbSubnetIds      string
}

type CreateNodePoolFlags struct {
	ClusterName             string
	ClusterId               string
	CompartmentId           string
	KubernetesVersion       string
	NodeImageName           string
	NodeShape               string
	KubernetesNetworkConfig string
	InitialNodeLabels       string
	SSHPublicKey            string
	QuantityPerSubnet       int
	SubnetIds               string
}

var (
	createClusterOKELong = templates.LongDesc(`
		This command creates a new kubernetes cluster on OKE, installing required local dependencies and provisions the
		Jenkins X platform

		You can see a demo of this command here: [http://jenkins-x.io/demos/create_cluster_oke/](http://jenkins-x.io/demos/create_cluster_oke/)

	  Oracle Cloud Infrastructure Container Engine for Kubernetes is a fully-managed, scalable, and highly available
	  service that you can use to deploy your containerized applications to the cloud.

		Oracle build the best of what we learn into Kubernetes, the industry-leading open source container orchestrator
		which powers Kubernetes Engine.

`)

	createClusterOKEExample = templates.Examples(`

		jx create cluster oke

`)
)

// NewCmdGet creates a command object for the generic "init" action, which
// installs the dependencies required to run the jenkins-x platform on a kubernetes cluster.
func NewCmdCreateClusterOKE(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := CreateClusterOKEOptions{
		CreateClusterOptions: createCreateClusterOptions(f, out, errOut, OKE),
	}
	cmd := &cobra.Command{
		Use:     "oci",
		Short:   "Create a new kubernetes cluster on OKE: Runs on Oracle Cloud",
		Long:    createClusterOKELong,
		Example: createClusterOKEExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}

	options.addCreateClusterFlags(cmd)
	options.addCommonFlags(cmd)

	cmd.Flags().StringVarP(&options.Flags.ClusterName, "name", "", "", "The  name  of  the  cluster.  Avoid  entering confidential information.")
	cmd.Flags().StringVarP(&options.Flags.CompartmentId, "compartment-id", "", "", "The OCID of the compartment in which to create the cluster.")
	cmd.Flags().StringVarP(&options.Flags.VcnId, "vcn-id", "", "", "The OCID of the virtual cloud network (VCN)  in  which  to  create  the cluster.")
	cmd.Flags().StringVarP(&options.Flags.KubernetesVersion, "kubernetes-version", "", "", "The  version  of  Kubernetes  to  install  into  the  cluster  masters.")
	cmd.Flags().StringVarP(&options.Flags.OKEOptions, "options", "", "", "Optional attributes for the cluster.")
	cmd.Flags().StringVarP(&options.Flags.WaitForState, "wait-for-state", "", "SUCCEEDED", " Specify this  option to perform the action and then wait until the work request reaches a certain state.")
	//KubernetesNetworkConfig and ServiceLbSubnetIds
	return cmd
}

func (o *CreateClusterOKEOptions) Run() error {
	err := o.installRequirements(OKE)
	if err != nil {
		return err
	}

	err = o.createClusterOKE()
	if err != nil {
		log.Errorf("error creating cluster %v", err)
		return err
	}

	return nil
}

func (o *CreateClusterOKEOptions) createClusterOKE() error {
	var err error
	//we assume user has prepared the oci config file under ~/.oci/
	if o.Flags.ClusterName == "" {
		o.Flags.ClusterName = strings.ToLower(randomdata.SillyName())
		log.Infof("No cluster name provided so using a generated one: %s\n", o.Flags.ClusterName)
	}

	compartmentId := o.Flags.CompartmentId
	if compartmentId == "" {
		prompt := &survey.Input{
			Message: "The OCID of the compartment in which to create the cluster.",
			Default: "",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &compartmentId, nil)
	}

	vcnId := o.Flags.VcnId
	if vcnId == "" {
		prompt := &survey.Input{
			Message: "The OCID of the virtual cloud network (VCN)  in  which  to  create  the cluster",
			Default: "",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &vcnId, nil)
	}

	kubernetesVersion := o.Flags.KubernetesVersion
	if kubernetesVersion == "" {
		prompt := &survey.Input{
			Message: "The OCID of the virtual cloud network (VCN)  in  which  to  create  the cluster",
			Default: "v1.9.7",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &kubernetesVersion, nil)
	}
	// mandatory flags are name,compartment-id,vcn-id,kubernetesVersion
	// need to figure out KubernetesNetworkConfig, ServiceLbSubnetIds
	args := []string{"ce", "clusters", "create",
		"--name", o.Flags.ClusterName,
		" --compartment-id", compartmentId,
		"--vcn-id", vcnId,
		"--kubernetes-version", kubernetesVersion}

	if o.Flags.OKEOptions != "" {
		args = append(args, "--options", o.Flags.OKEOptions)
	}

	args = append(args, "--wait-for-state", "SUCCEEDED")

	log.Info("Creating cluster...\n")
	err = o.runCommand("oci", args...)
	if err != nil {
		return err
	}

	log.Info("Initialising cluster ...\n")
	// to be added
	return nil
}
