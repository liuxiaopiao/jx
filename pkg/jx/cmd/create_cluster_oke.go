package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	ClusterName                  string
	CompartmentId                string
	VcnId                        string
	KubernetesVersion            string
	WaitForState                 string
	Endpoint                     string
	PodsCidr                     string
	ServicesCidr                 string
	IsKubernetesDashboardEnabled string
	IsTillerEnabled              string
	ServiceLbSubnetIds           string
}

type KubernetesNetworkConfig struct {
	PodsCidr     string `json:"podsCidr"`
	ServicesCidr string `json:"servicesCidr"`
}

type AddOns struct {
	IsKubernetesDashboardEnabled bool `json:"isKubernetesDashboardEnabled"`
	IsTillerEnabled              bool `json:"isTillerEnabled"`
}

type ClusterCustomOptions struct {
	ServiceLbSubnetIds      []string                `json:"serviceLbSubnetIds"`
	AddOns                  AddOns                  `json:"addOns"`
	KubernetesNetworkConfig KubernetesNetworkConfig `json:"kubernetesNetworkConfig"`
}

type CreateNodePoolFlags struct {
	ClusterId         string
	CompartmentId     string
	KubernetesVersion string
	NodeImageName     string
	NodeShape         string
	//KubernetesNetworkConfig string
	//InitialNodeLabels       string
	SSHPublicKey      string
	QuantityPerSubnet int
	SubnetIds         string
}

type PoolCustomOptions struct {
	SubnetIds []string `json:"subnetIds"`
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
		Use:     "oke",
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
	cmd.Flags().StringVarP(&options.Flags.Endpoint, "Endpoint", "", "", "Endpoint for the environment.")
	cmd.Flags().StringVarP(&options.Flags.WaitForState, "wait-for-state", "", "SUCCEEDED", "Specify this  option to perform the action and then wait until the work request reaches a certain state.")
	cmd.Flags().StringVarP(&options.Flags.PodsCidr, "PodsCidr", "", "", "PODS CIDR Block.")
	cmd.Flags().StringVarP(&options.Flags.ServicesCidr, "ServicesCidr", "", "", "Kubernetes Service CIDR Block.")
	cmd.Flags().StringVarP(&options.Flags.IsKubernetesDashboardEnabled, "IsKubernetesDashboardEnabled", "", "true", "Is KubernetesDashboard Enabled.")
	cmd.Flags().StringVarP(&options.Flags.IsTillerEnabled, "IsTillerEnabled", "", "true", "Is Tiller Enabled.")
	cmd.Flags().StringVarP(&options.Flags.ServiceLbSubnetIds, "ServiceLbSubnetIds", "", "", "Kubernetes Service LB Subnets.")
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
	//we assume user has prepared the oci config file under ~/.oci/
	//need to set the environment variable first
	endpoint := o.Flags.Endpoint
	if endpoint == "" {
		prompt := &survey.Input{
			Message: "The corresponding regional endpoint",
			Default: "",
			Help:    "This is required environment variable",
		}

		survey.AskOne(prompt, &endpoint, nil)
	}
	os.Setenv("ENDPOINT", endpoint)

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
			Message: "The  version  of  Kubernetes  to  install  into  the  cluster  masters",
			Default: "v1.9.7",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &kubernetesVersion, nil)
	}

	// mandatory flags are name,compartment-id,vcn-id,kubernetesVersion
	// need to figure out KubernetesNetworkConfig, ServiceLbSubnetIds
	//		"--endpoint", o.Flags.Endpoint,
	args := []string{"ce", "cluster", "create",
		"--name", o.Flags.ClusterName,
		"--compartment-id", compartmentId,
		"--vcn-id", vcnId,
		"--kubernetes-version", kubernetesVersion}

	args = append(args, "--wait-for-state", "SUCCEEDED")

	resp := ClusterCustomOptions{
		ServiceLbSubnetIds: []string{"ocid1.subnet.oc1.phx.aaaaaaaavnbpcy4cvbnsfvntzrcrgrmralkbo4ysbewwhnoq4okatjato27a", "ocid1.subnet.oc1.phx.aaaaaaaahvymb43jigjsolmkkvtnfu5kckw3c72vuurqcknv3pvxxk5euhba"},
		AddOns: AddOns{
			IsKubernetesDashboardEnabled: true,
			IsTillerEnabled:              true,
		},
		KubernetesNetworkConfig: KubernetesNetworkConfig{
			PodsCidr:     "10.244.0.0/16",
			ServicesCidr: "10.96.0.0/16",
		}}

	js, _ := json.Marshal(resp)

	err := ioutil.WriteFile("/tmp/oke_cluster_config.json", js, 0644)
	if err != nil {
		log.Errorf("error write file to /tmp file %v", err)
		return err
	}
	fmt.Printf("%s", js)

	//if o.Flags.OKEOptions != "" {
	args = append(args, "--options", "file:///tmp/oke_cluster_config.json")
	//}

	fmt.Printf("%s", args)
	log.Info("\nCreating cluster...\n")
	err = o.runCommandVerbose("oci", args...)
	//output, err := o.getCommandOutput("oci", args...)
	if err != nil {
		return err
	}

	//create node pool

	getCredentials := []string{"aks", "get-credentials", "--resource-group", resourceName, "--name", clusterName}

	err = o.runCommand("az", getCredentials...)

	if err != nil {
		return err
	}

	log.Info("Initialising cluster ...\n")
	// to be added
	log.Info("Creating node pool ...\n")
	return nil
}
