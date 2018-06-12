package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/jenkins-x/jx/pkg/jx/cmd/log"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
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
	IsKubernetesDashboardEnabled bool
	IsTillerEnabled              bool
	ServiceLbSubnetIds           string
	NodePoolName                 string
	NodeImageName                string
	NodeShape                    string
	SSHPublicKey                 string
	QuantityPerSubnet            string
	NodePoolSubnetIds            string
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

/*
type CreateNodePoolFlags struct {
	ClusterId         string
	CompartmentId     string
	NodePoolName      string
	KubernetesVersion string
	NodeImageName     string
	NodeShape         string
	SSHPublicKey      string
	QuantityPerSubnet int
}
*/

type PoolCustomOptions struct {
	NodePoolSubnetIds []string `json:"nodePoolSubnetIds"`
}

/*
type ClusterResourcesOutput struct {
	ActionType string `json:"action-type"`
	Entitytype string `json:"entity-type"`
	EntityUri  string `json:"entity-uri"`
	Identifier string `json:"identifier"`
}

type ClusterDataOutput struct {
	ClusterDataOutputCompartmentId string                 `json:"compartment-id"`
	ClusterDataOutputId            string                 `json:"id"`
	ClusterDataOutputOperationType string                 `json:"operation-type"`
	Identifier                     string                 `json:"identifier"`
	ClusterResourcesOutput         ClusterResourcesOutput `json:"clusterResourcesOutput"`

	ClusterDataOutputStatus       string `json:"status"`
	ClusterDataOutputTimeAccepted string `json:"time-accepted"`
	ClusterDataOutputTimeFinished string `json:"time-finished"`
	ClusterDataOutputTimeStarted  string `json:"time-started"`
}

type ClusterOutput struct {
	Etag              string            `json:"etag"`
	ClusterDataOutput ClusterDataOutput `json:"clusterDataOutput"`
}
*/

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
	cmd.Flags().BoolVarP(&options.Flags.IsKubernetesDashboardEnabled, "IsKubernetesDashboardEnabled", "", true, "Is KubernetesDashboard Enabled.")
	cmd.Flags().BoolVarP(&options.Flags.IsTillerEnabled, "IsTillerEnabled", "", false, "Is Tiller Enabled.")
	cmd.Flags().StringVarP(&options.Flags.ServiceLbSubnetIds, "ServiceLbSubnetIds", "", "", "Kubernetes Service LB Subnets.")
	cmd.Flags().StringVarP(&options.Flags.NodePoolName, "NodePoolName", "", "", "The  name  of  the  node pool.")
	cmd.Flags().StringVarP(&options.Flags.NodeImageName, "NodeImageName", "", "", "The name of the image running on the nodes in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.NodeShape, "NodeShape", "", "", "The name of the node shape of the nodes in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.SSHPublicKey, "SSHPublicKey", "", "", "The SSH public key to add to each node in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.QuantityPerSubnet, "QuantityPerSubnet", "", "", "The number of nodes to create in each subnet.")
	cmd.Flags().StringVarP(&options.Flags.NodePoolSubnetIds, "NodePoolSubnetIds", "", "", "The  OCIDs  of  the subnets in which to place nodes for this node pool.")

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
			Message: "The corresponding regional endpoint:",
			Default: "",
			Help:    "This is required environment variable",
		}

		survey.AskOne(prompt, &endpoint, nil)
	}
	fmt.Printf("Endpoint is %s\n", endpoint)
	os.Setenv("ENDPOINT", endpoint)

	if o.Flags.ClusterName == "" {
		o.Flags.ClusterName = strings.ToLower(randomdata.SillyName())
		log.Infof("No cluster name provided so using a generated one: %s\n", o.Flags.ClusterName)
	}

	compartmentId := o.Flags.CompartmentId
	if compartmentId == "" {
		prompt := &survey.Input{
			Message: "The OCID of the compartment in which to create the cluster:",
			Default: "",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &compartmentId, nil)
	}

	vcnId := o.Flags.VcnId
	if vcnId == "" {
		prompt := &survey.Input{
			Message: "The OCID of the virtual cloud network (VCN)  in  which  to  create  the cluster:",
			Default: "",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &vcnId, nil)
	}

	kubernetesVersion := o.Flags.KubernetesVersion
	if kubernetesVersion == "" {
		prompt := &survey.Input{
			Message: "The version  of  Kubernetes  to  install  into  the  cluster  masters:",
			Default: "v1.9.7",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &kubernetesVersion, nil)
	}

	//Start processing optional parameters
	serviceLbSubnetIds := o.Flags.ServiceLbSubnetIds
	if serviceLbSubnetIds == "" {
		prompt := &survey.Input{
			Message: "The OCIDs of Service load balance subnets:",
			Default: "",
			Help:    "This is optional parameter",
		}

		survey.AskOne(prompt, &serviceLbSubnetIds, nil)
	}

	serviceLbSubnetIdsArray := strings.Split(serviceLbSubnetIds, ",")

	isKubernetesDashboardEnabled := o.Flags.IsKubernetesDashboardEnabled

	isTillerEnabled := o.Flags.IsTillerEnabled

	podsCidr := o.Flags.PodsCidr
	if podsCidr == "" {
		prompt := &survey.Input{
			Message: "PODS CIDR BLOCK:",
			Default: "10.244.0.0/16",
			Help:    "This is optional parameter",
		}

		survey.AskOne(prompt, &podsCidr, nil)
	}

	servicesCidr := o.Flags.ServicesCidr
	if servicesCidr == "" {
		prompt := &survey.Input{
			Message: "KUBERNETES SERVICE CIDR BLOCK:",
			Default: "10.96.0.0/16",
			Help:    "This is optional parameter",
		}

		survey.AskOne(prompt, &servicesCidr, nil)
	}

	//Get node pool settings
	if o.Flags.NodePoolName == "" {
		o.Flags.NodePoolName = strings.ToLower(randomdata.SillyName())
		log.Infof("No node pool name provided so using a generated one: %s\n", o.Flags.NodePoolName)
	}

	nodeImageName := o.Flags.NodeImageName
	if nodeImageName == "" {
		prompt := &survey.Input{
			Message: "The name of the image running on the nodes in the node pool:",
			Default: "Oracle-Linux-7.4",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &nodeImageName, nil)
	}

	nodeShape := o.Flags.NodeShape
	if nodeShape == "" {
		prompt := &survey.Input{
			Message: "The name of the node shape of the nodes in the node pool:",
			Default: "VM.Standard1.1",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &nodeShape, nil)
	}

	nodePoolSubnetIds := o.Flags.NodePoolSubnetIds
	if nodePoolSubnetIds == "" {
		prompt := &survey.Input{
			Message: "The OCIDs of the subnets in which to place nodes for this node pool:",
			Default: "",
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &nodePoolSubnetIds, nil)
	}
	//nodePoolSubnetIdsArray := strings.Split(nodePoolSubnetIds, ",")

	sshPublicKey := o.Flags.SSHPublicKey
	if sshPublicKey == "" {
		prompt := &survey.Input{
			Message: "The SSH public key to add to each node in the node pool:",
			Default: "",
			Help:    "This is optional parameter",
		}

		survey.AskOne(prompt, &sshPublicKey, nil)
	}

	quantityPerSubnet := o.Flags.QuantityPerSubnet
	if quantityPerSubnet == "" {
		prompt := &survey.Input{
			Message: "The number of nodes to create in each subnet:",
			Default: "1",
			Help:    "This is optional parameter",
		}

		survey.AskOne(prompt, &quantityPerSubnet, nil)
	}
	/*
		quantityPerSubnetNum, err := strconv.Atoi(quantityPerSubnet)
		if err != nil {
			log.Errorf("error convert string to int %v", err)
			return err
		}
	*/
	args := []string{"ce", "cluster", "create",
		"--name", o.Flags.ClusterName,
		"--compartment-id", compartmentId,
		"--vcn-id", vcnId,
		"--kubernetes-version", kubernetesVersion}

	args = append(args, "--wait-for-state", "SUCCEEDED")

	resp := ClusterCustomOptions{
		ServiceLbSubnetIds: serviceLbSubnetIdsArray,
		AddOns: AddOns{
			IsKubernetesDashboardEnabled: isKubernetesDashboardEnabled,
			IsTillerEnabled:              isTillerEnabled,
		},
		KubernetesNetworkConfig: KubernetesNetworkConfig{
			PodsCidr:     podsCidr,
			ServicesCidr: servicesCidr,
		}}

	js, _ := json.Marshal(resp)

	err := ioutil.WriteFile("/tmp/oke_cluster_config.json", js, 0644)
	if err != nil {
		log.Errorf("error write file to /tmp file %v", err)
		return err
	}

	fmt.Printf("Cluster creation output json is %s\n", js)

	//if o.Flags.OKEOptions != "" {
	args = append(args, "--options", "file:///tmp/oke_cluster_config.json")
	//}

	fmt.Printf("Args are: %s\n", args)
	log.Info("Creating cluster...\n")
	//err = o.runCommandVerbose("oci", args...)
	output, err := o.getCommandOutput("", "oci", args...)
	if err != nil {
		return err
	}

	fmt.Printf("Create cluster output: %s\n", output)

	/*
		fmt.Printf("Cluster JSON: %s\n", output)
		clusterOutputJson := []byte(output)
		err = ioutil.WriteFile("/tmp/oke_cluster_info.json", clusterOutputJson, 0644)
		//get cluster id
		data := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(output))
		dec.Decode(&data)
		jq := jsonq.NewQuery(data)
		clusterinfo, err := jq.ArrayOfObjects("data", "resources")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		clusterId := clusterinfo[0]["identifier"]
	*/

	if strings.Contains(output, "identifier") {
		subClusterInfo := strings.Split(output, "identifier")
		clusterIdRaw := strings.Split(subClusterInfo[1], "}")
		clusterId := strings.TrimSpace(strings.Replace(clusterIdRaw[0][4:], "\"", "", -1))
		fmt.Printf("Cluster id: %s\n", clusterId)

		//setup the kube context
		log.Info("Setup kube context ...\n")
		var kubeconfigFile = ""
		if home := util.HomeDir(); home != "" {
			kubeconfigFile = filepath.Join(util.HomeDir(), "kubeconfig")
		} else {
			kubeconfigFile = filepath.Join("/tmp", "kubeconfig")
		}

		kubeContextArgs := []string{"ce", "cluster", "create-kubeconfig",
			"--cluster-id", clusterId,
			"--file", kubeconfigFile}

		err = o.runCommandVerbose("oci", kubeContextArgs...)
		if err != nil {
			return err
		}
		os.Setenv("KUBECONFIG", kubeconfigFile)

		//create node pool
		log.Info("Creating node pool ...\n")

		poolArgs := "ce node-pool create --name=" + o.Flags.NodePoolName + " --compartment-id=" + compartmentId + " --cluster-id=" + clusterId + " --kubernetes-version=" + kubernetesVersion + " --node-image-name=" + nodeImageName + " --node-shape=" + nodeShape + " --quantity-per-subnet=" + quantityPerSubnet + " --subnet-ids=file:///tmp/oke_pool_config.json" + " --wait-for-state=SUCCEEDED"

		/*
			args := []string{"ce", "node-pool", "create",
				"--name=", o.Flags.NodePoolName,
				"--compartment-id=", compartmentId,
				"--cluster-id=", clusterId,
				"--kubernetes-version=", kubernetesVersion,
				"--node-image-name=", nodeImageName,
				//"--ssh-public-key=", sshPublicKey,
				"--node-shape=", nodeShape,
				"--quantity-per-subnet=", quantityPerSubnet,
				"--subnet-ids=", "file:///tmp/oke_pool_config.json",
			}

			//args = append(args, "--wait-for-state=", "SUCCEEDED")

				resp := PoolCustomOptions{
					NodePoolSubnetIds: nodePoolSubnetIdsArray}

				js, _ := json.Marshal(resp)

				err := ioutil.WriteFile("/tmp/oke_pool_config.json", js, 0644)
				if err != nil {
					log.Errorf("error write file to /tmp file %v", err)
					return err
				}

				fmt.Printf("Pool creation output json is %s\n", js)

				args = append(args, "--options", "file:///tmp/oke_pool_config.json")
		*/

		fmt.Printf("Node pool creation args are: %s\n", poolArgs)

		log.Info("Creating Node Pool...\n")
		poolArgsArray := strings.Split(poolArgs, " ")
		//err = o.runCommandVerbose("oci", poolArgsArray...)
		poolCreationOutput, err := o.getCommandOutput("", "oci", poolArgsArray...)
		if err != nil {
			return err
		}

		//wait for node pool active
		if strings.Contains(poolCreationOutput, "identifier") {
			subPoolInfo := strings.Split(poolCreationOutput, "identifier")
			poolIdRaw := strings.Split(subPoolInfo[1], "}")
			poolId := strings.TrimSpace(strings.Replace(poolIdRaw[0][4:], "\"", "", -1))
			fmt.Printf("Node Pool id: %s\n", poolId)

			//get node pool status until they are active
			nodeQuantity, err := strconv.Atoi(quantityPerSubnet)
			if err != nil {
				return err
			}
			status := regexp.MustCompile("ACTIVE")
			for {
				getPoolStatusArgs := []string{"ce", "node-pool", "get", "--node-pool-id", poolId}
				poolStatusOutput, err := o.getCommandOutput("", "oci", getPoolStatusArgs...)
				if err != nil {
					return err
				}

				count := len(status.FindAllStringIndex(poolStatusOutput, -1))
				fmt.Printf("Now only %d nodes are ready\n", count)
				if count == nodeQuantity {
					break
				}

			}

			log.Info("Initialising cluster ...\n")

			return o.initAndInstall(OKE)
		}
	}
	return nil
}
