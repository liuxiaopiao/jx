package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/jenkins-x/jx/pkg/jx/cmd/oce"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

// CreateClusterOptions the flags for running create cluster
type CreateClusterOCEOptions struct {
	CreateClusterOptions
	Flags CreateClusterOCEFlags
}

type CreateClusterOCEFlags struct {
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
	ClusterMaxWaitSeconds        string
	ClusterWaitIntervalSeconds   string
	InitialNodeLabels            string
	PoolMaxWaitSeconds           string
	PoolWaitIntervalSeconds      string
}

var (
	createClusterOCELong = templates.LongDesc(`
		This command creates a new kubernetes cluster on OCE, installing required local dependencies and provisions the
		Jenkins X platform

		You can see a demo of this command here: [http://jenkins-x.io/demos/create_cluster_oce/](http://jenkins-x.io/demos/create_cluster_oce/)

	  Oracle Cloud Infrastructure Container Engine for Kubernetes is a fully-managed, scalable, and highly available
	  service that you can use to deploy your containerized applications to the cloud.

		Oracle build the best of what we learn into Kubernetes, the industry-leading open source container orchestrator
		which powers Kubernetes Engine.

`)

	createClusterOCEExample = templates.Examples(`

		jx create cluster oce

`)
)

// NewCmdGet creates a command object for the generic "init" action, which
// installs the dependencies required to run the jenkins-x platform on a kubernetes cluster.
func NewCmdCreateClusterOCE(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := CreateClusterOCEOptions{
		CreateClusterOptions: createCreateClusterOptions(f, out, errOut, OCE),
	}
	cmd := &cobra.Command{
		Use:     "oce",
		Short:   "Create a new kubernetes cluster on OCE: Runs on Oracle Cloud",
		Long:    createClusterOCELong,
		Example: createClusterOCEExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}

	options.addCreateClusterFlags(cmd)
	options.addCommonFlags(cmd)

	cmd.Flags().StringVarP(&options.Flags.ClusterName, "name", "", "", "The name of the cluster. Avoid entering confidential information.")
	cmd.Flags().StringVarP(&options.Flags.CompartmentId, "compartmentId", "", "", "The OCID of the compartment in which to create the cluster.")
	cmd.Flags().StringVarP(&options.Flags.VcnId, "vcnId", "", "", "The OCID of the virtual cloud network (VCN) in which to create the cluster.")
	cmd.Flags().StringVarP(&options.Flags.KubernetesVersion, "kubernetesVersion", "", "", "The version of Kubernetes to install into the cluster masters.")
	cmd.Flags().StringVarP(&options.Flags.Endpoint, "endpoint", "", "", "Endpoint for the environment.")
	cmd.Flags().StringVarP(&options.Flags.WaitForState, "waitForState", "", "SUCCEEDED", "Specify this option to perform the action and then wait until the work request reaches a certain state.")
	cmd.Flags().StringVarP(&options.Flags.PodsCidr, "podsCidr", "", "", "PODS CIDR Block.")
	cmd.Flags().StringVarP(&options.Flags.ServicesCidr, "servicesCidr", "", "", "Kubernetes Service CIDR Block.")
	cmd.Flags().BoolVarP(&options.Flags.IsKubernetesDashboardEnabled, "isKubernetesDashboardEnabled", "", true, "Is KubernetesDashboard Enabled.")
	cmd.Flags().BoolVarP(&options.Flags.IsTillerEnabled, "isTillerEnabled", "", false, "Is Tiller Enabled.")
	cmd.Flags().StringVarP(&options.Flags.ServiceLbSubnetIds, "serviceLbSubnetIds", "", "", "Kubernetes Service LB Subnets.")
	cmd.Flags().StringVarP(&options.Flags.NodePoolName, "nodePoolName", "", "", "The name of the node pool.")
	cmd.Flags().StringVarP(&options.Flags.NodeImageName, "nodeImageName", "", "", "The name of the image running on the nodes in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.NodeShape, "nodeShape", "", "", "The name of the node shape of the nodes in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.SSHPublicKey, "sshPublicKey", "", "", "The SSH public key to add to each node in the node pool.")
	cmd.Flags().StringVarP(&options.Flags.QuantityPerSubnet, "quantityPerSubnet", "", "", "The number of nodes to create in each subnet.")
	cmd.Flags().StringVarP(&options.Flags.NodePoolSubnetIds, "nodePoolSubnetIds", "", "", "The OCIDs of the subnets in which to place nodes for this node pool.")
	cmd.Flags().StringVarP(&options.Flags.ClusterMaxWaitSeconds, "clusterMaxWaitSeconds", "", "", "The maximum time to wait for the work request to reach the state defined by --wait-for-state. Defaults to 1200 seconds.")
	cmd.Flags().StringVarP(&options.Flags.ClusterWaitIntervalSeconds, "clusterWaitIntervalSeconds", "", "", "Check every --wait-interval-seconds to see whether the work request to see if it has reached the state defined by --wait-for-state.")
	cmd.Flags().StringVarP(&options.Flags.InitialNodeLabels, "initialNodeLabels", "", "", "A list of key/value pairs to add to nodes after they join the Kubernetes cluster.")
	cmd.Flags().StringVarP(&options.Flags.PoolMaxWaitSeconds, "poolMaxWaitSeconds", "", "", "The maximum time to wait for the work request to reach the state defined by --wait-for-state. Defaults to 1200 seconds.")
	cmd.Flags().StringVarP(&options.Flags.PoolWaitIntervalSeconds, "poolWaitIntervalSeconds", "", "", "Check every --wait-interval-seconds to see whether the work request to see if it has reached the state defined by --wait-for-state.")

	return cmd
}

func (o *CreateClusterOCEOptions) Run() error {
	err := o.installRequirements(OCE)
	if err != nil {
		return err
	}

	err = o.createClusterOCE()
	if err != nil {
		log.Errorf("error creating cluster %v", err)
		return err
	}

	return nil
}

func (o *CreateClusterOCEOptions) createClusterOCE() error {
	//we assume user has prepared the oci config file under ~/.oci/

	imagesArray, kubeVersionsArray, shapesArray, latestKubeVersion, err := oce.GetOptionValues()
	if err != nil {
		fmt.Println("error")
	}
	fmt.Println("Image array is %s\n", imagesArray)
	fmt.Println("kubeVersionsArray array is %s\n", kubeVersionsArray)
	fmt.Println("shapesArray array is %s\n", shapesArray)

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
		prompt := &survey.Select{
			Message: "The version  of  Kubernetes  to  install  into  the  cluster  masters:",
			Options: kubeVersionsArray,
			Default: latestKubeVersion,
			Help:    "This is required parameter",
		}

		survey.AskOne(prompt, &kubernetesVersion, nil)
	}

	//Get node pool settings
	if o.Flags.NodePoolName == "" {
		o.Flags.NodePoolName = strings.ToLower(randomdata.SillyName())
		log.Infof("No node pool name provided so using a generated one: " + o.Flags.NodePoolName + "\n")
	}

	nodeImageName := o.Flags.NodeImageName
	if nodeImageName == "" {
		prompt := &survey.Select{
			Message:  "The name of the image running on the nodes in the node pool:",
			Options:  imagesArray,
			Default:  "Oracle-Linux-7.4",
			Help:     "This is required parameter",
			PageSize: 10,
		}

		survey.AskOne(prompt, &nodeImageName, nil)
	}

	nodeShape := o.Flags.NodeShape
	if nodeShape == "" {
		prompt := &survey.Select{
			Message:  "The name of the node shape of the nodes in the node pool:",
			Options:  shapesArray,
			Default:  "VM.Standard1.1",
			Help:     "This is required parameter",
			PageSize: 10,
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
	nodePoolSubnetIdsArray := strings.Split(nodePoolSubnetIds, ",")
	for i := range nodePoolSubnetIdsArray {
		nodePoolSubnetIdsArray[i] = "\"" + nodePoolSubnetIdsArray[i] + "\""
	}
	nodePoolSubnetIdsTemp := "[" + strings.Join(nodePoolSubnetIdsArray, ",") + "]"
	err = ioutil.WriteFile("/tmp/oce_pool_config.json", []byte(nodePoolSubnetIdsTemp), 0644)
	if err != nil {
		fmt.Printf("error write file to /tmp file %v", err)
	}

	args := []string{"ce", "cluster", "create",
		"--name", o.Flags.ClusterName,
		"--compartment-id", compartmentId,
		"--vcn-id", vcnId,
		"--kubernetes-version", kubernetesVersion}

	args = append(args, "--wait-for-state", "SUCCEEDED")

	//Start processing optional parameters
	serviceLbSubnetIds := o.Flags.ServiceLbSubnetIds
	if serviceLbSubnetIds != "" {
		/*
			serviceLbSubnetIdsArray := strings.Split(serviceLbSubnetIds, ",")
			resp := ClusterCustomOptions{
				ServiceLbSubnetIds: serviceLbSubnetIdsArray,
			}

			js, _ := json.Marshal(resp)

			err := ioutil.WriteFile("/tmp/oce_cluster_config.json", js, 0644)
			if err != nil {
				log.Errorf("error write file to /tmp file %v", err)
				return err
			}
		*/
		//serviceLbSubnetIdsArray := strings.Split(serviceLbSubnetIds, ",")

		serviceLbSubnetIdsArray := strings.Split(serviceLbSubnetIds, ",")
		for i := range serviceLbSubnetIdsArray {
			serviceLbSubnetIdsArray[i] = "\"" + serviceLbSubnetIdsArray[i] + "\""
		}

		serviceLbSubnetIdsTemp := "[" + strings.Join(serviceLbSubnetIdsArray, ",") + "]"

		err := ioutil.WriteFile("/tmp/oce_cluster_config.json", []byte(serviceLbSubnetIdsTemp), 0644)
		if err != nil {
			fmt.Printf("error write file to /tmp file %v", err)
		}

		args = append(args, "--service-lb-subnet-ids", "file:///tmp/oce_cluster_config.json")
	}

	isKubernetesDashboardEnabled := o.Flags.IsKubernetesDashboardEnabled
	if !isKubernetesDashboardEnabled {
		args = append(args, "--dashboard-enabled", "false")
	}

	isTillerEnabled := o.Flags.IsTillerEnabled
	if !isTillerEnabled {
		args = append(args, "--tiller-enabled", "false")
	}

	podsCidr := o.Flags.PodsCidr
	if podsCidr != "" {
		args = append(args, "--pods-cidr", podsCidr)
	}

	servicesCidr := o.Flags.ServicesCidr
	if servicesCidr != "" {
		args = append(args, "--services-cidr", servicesCidr)
	}

	clusterMaxWaitSeconds := o.Flags.ClusterMaxWaitSeconds
	if clusterMaxWaitSeconds != "" {
		args = append(args, "--max-wait-seconds", clusterMaxWaitSeconds)
	}

	clusterWaitIntervalSeconds := o.Flags.ClusterWaitIntervalSeconds
	if clusterWaitIntervalSeconds != "" {
		args = append(args, "--wait-interval-seconds", clusterWaitIntervalSeconds)
	}

	fmt.Printf("Args are: %s\n", args)
	log.Info("Creating cluster...\n")
	output, err := o.getCommandOutput("", "oci", args...)
	if err != nil {
		return err
	}

	fmt.Printf("Create cluster output: %s\n", output)

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

		poolArgs := "ce node-pool create --name=" + o.Flags.NodePoolName + " --compartment-id=" + compartmentId + " --cluster-id=" + clusterId + " --kubernetes-version=" + kubernetesVersion + " --node-image-name=" + nodeImageName + " --node-shape=" + nodeShape + " --subnet-ids=file:///tmp/oce_pool_config.json" + " --wait-for-state=SUCCEEDED"

		quantityPerSubnet := o.Flags.QuantityPerSubnet
		quantityPerSubnet = (map[bool]string{true: quantityPerSubnet, false: "1"})[quantityPerSubnet != ""]
		log.Info("Will create " + quantityPerSubnet + " node per subnet ...\n")
		poolArgs = poolArgs + " --quantity-per-subnet=" + quantityPerSubnet

		initialNodeLabels := o.Flags.InitialNodeLabels
		if initialNodeLabels != "" {
			initialNodeLabelsJson := "[" + initialNodeLabels + "]"
			err := ioutil.WriteFile("/tmp/oce_pool_labels_config.json", []byte(initialNodeLabelsJson), 0644)
			if err != nil {
				fmt.Printf("error write file to /tmp file %v", err)
			}
			poolArgs = poolArgs + " --initial-node-labels=file:///tmp/oce_pool_labels_config.json"
		}

		poolMaxWaitSeconds := o.Flags.PoolMaxWaitSeconds
		if poolMaxWaitSeconds != "" {

			poolArgs = poolArgs + " --max-wait-seconds=" + poolMaxWaitSeconds
		}

		poolWaitIntervalSeconds := o.Flags.PoolWaitIntervalSeconds
		if poolWaitIntervalSeconds != "" {
			poolArgs = poolArgs + " --wait-interval-seconds=" + poolWaitIntervalSeconds
		}

		log.Info("Creating Node Pool...\n")
		poolArgsArray := strings.Split(poolArgs, " ")

		if o.Flags.SSHPublicKey != "" {
			sshPubKey := "--ssh-public-key=" + o.Flags.SSHPublicKey
			poolArgsArray = append(poolArgsArray, sshPubKey)
		}

		fmt.Printf("Pool creation args are: %s\n", poolArgsArray)
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

			err = o.waitForNodeToComeUp(nodeQuantity*len(nodePoolSubnetIdsArray), poolId)
			if err != nil {
				return fmt.Errorf("Failed to wait for Kubernetes cluster node to be ready: %s\n", err)
			}

			if isTillerEnabled {
				//need to wait for tiller pod is running
				fmt.Printf("Wait for tiller pod is running\n")
				err = o.waitForTillerComeUp()
				if err != nil {
					return fmt.Errorf("Failed to wait for Tiller to be ready: %s\n", err)
				}
			}

			err = util.DeleteFile("/tmp/oce_cluster_config.json")
			if err != nil {
				return err
			}
			err = util.DeleteFile("/tmp/oce_pool_config.json")
			if err != nil {
				return err
			}
			err = util.DeleteFile("/tmp/oce_pool_labels_config.json")
			if err != nil {
				return err
			}
			log.Info("Initialising cluster ...\n")

			return o.initAndInstall(OCE)
		}
	}
	return nil
}

func (o *CreateClusterOCEOptions) waitForNodeToComeUp(nodeQuantity int, poolId string) error {
	attempts := 1000
	status := regexp.MustCompile("ACTIVE")
	getPoolStatusArgs := []string{"ce", "node-pool", "get", "--node-pool-id", poolId}
	for i := 0; ; i++ {
		poolStatusOutput, err := o.getCommandOutput("", "oci", getPoolStatusArgs...)
		if err != nil {
			return err
		}

		count := len(status.FindAllStringIndex(poolStatusOutput, -1))
		fmt.Printf("Now only %d nodes are ready\n", count)
		if count == nodeQuantity {
			break
		}
		time.Sleep(time.Second * 5)
		if i >= attempts {
			return fmt.Errorf("Retry %d times and nodes are still not ready. Please check it manually.", attempts)
		}
	}
	return nil
}

func (o *CreateClusterOCEOptions) waitForTillerComeUp() error {
	f := func() error {
		tillerStatus := "kubectl get --namespace=kube-system deployment/tiller-deploy  | tail -n +2 | awk '{print $5}' | grep 1"
		return o.runCommandQuietly("bash", "-c", tillerStatus)
	}
	return o.retryQuiet(2000, time.Second*10, f)
}
