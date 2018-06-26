package oce

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func GetOptionValues() ([]string, []string, []string, string, error) {
	jsonString, err := exec.Command("oci", "ce", " cluster-options", "get", "--cluster-option-id", "all").Output()
	if err != nil {
		return nil, nil, nil, "", err
	}
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &dat); err != nil {
		fmt.Println("error")
		return nil, nil, nil, "", err
	}

	originalStrs := dat["data"].(map[string]interface{})

	kubeVersions := fmt.Sprintf("%v", originalStrs["kubernetes-versions"])
	kubeVersions = strings.TrimPrefix(kubeVersions, "[")
	kubeVersions = strings.TrimSuffix(kubeVersions, "]")
	kubeVersionsArray := strings.Split(kubeVersions, " ")

	images := fmt.Sprintf("%v", originalStrs["images"])
	images = strings.TrimPrefix(images, "[")
	images = strings.TrimSuffix(images, "]")
	imagesArray := strings.Split(images, " ")

	shapes := fmt.Sprintf("%v", originalStrs["shapes"])
	shapes = strings.TrimPrefix(shapes, "[")
	shapes = strings.TrimSuffix(shapes, "]")
	shapesArray := strings.Split(shapes, " ")

	sort.Strings(kubeVersionsArray)

	return imagesArray, kubeVersionsArray, shapesArray, kubeVersionsArray[0], nil
}

func GetOracleShapes() []string {

	return []string{
		"VM.Standard1.1",
		"VM.Standard1.2",
		"VM.Standard1.4",
		"VM.Standard1.8",
		"VM.Standard1.16",
		"VM.DenseIO1.4",
		"VM.DenseIO1.8",
		"VM.DenseIO1.16",
		"VM.Standard2.1",
		"VM.Standard2.2",
		"VM.Standard2.4",
		"VM.Standard2.8",
		"VM.Standard2.16",
		"VM.Standard.2.24",
		"VM.DenseIO2.8",
		"VM.DenseIO2.16",
		"VM.DenseIO2.24",
		"BM.Standard1.36",
		"BM.DenseIO1.36",
		"BM.Standard2.52",
		"BM.DenseO2.52",
	}
}
