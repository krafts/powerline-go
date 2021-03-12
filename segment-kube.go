package main

import (
	"fmt"
	pwl "github.com/justjanne/powerline-go/powerline"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

// KubeContext holds the kubernetes context
type KubeContext struct {
	Context struct {
		Cluster   string
		Namespace string
		User      string
	}
	Name string
}

// KubeConfig is the kubernetes configuration
type KubeConfig struct {
	Contexts       []KubeContext `yaml:"contexts"`
	CurrentContext string        `yaml:"current-context"`
}

type Environment struct {
	Type       string `yaml:"type"`
	Name       string `yaml:"name"`
	Contains   []string `yaml:"contains"`
	Prefix     []string `yaml:"prefix"`
	Suffix     []string `yaml:"suffix"`
}

type Environments struct {
	Environments []Environment `yaml:"environments"`
	Remove       []string `yaml:"remove"`
}

func getEnvironments() Environments {
	var k8sEnvs Environments
	k8sEnvsFile := path.Join(homePath(), ".kubectl_context_ps1_environments.yaml")
	stat, err := os.Stat(k8sEnvsFile)
	if err == nil && !stat.IsDir() {
		environments, err := ioutil.ReadFile(k8sEnvsFile)

		// unmarshall it
		err = yaml.Unmarshal(environments, &k8sEnvs)
		if err == nil {
		}
	}
	return k8sEnvs
}

func doesEnvMatch(cluster string, k8sEnvs *Environments, envType string) bool {
	match := false

	for i := 0; i < len(k8sEnvs.Environments); i++ {
		envMatch := (envType == k8sEnvs.Environments[i].Type)

		for j := 0; j < len(k8sEnvs.Environments[i].Contains); j++ {
			if strings.Contains(cluster, k8sEnvs.Environments[i].Contains[j]) && envMatch {
				return true
			}
		}
		for j := 0; j < len(k8sEnvs.Environments[i].Prefix); j++ {
			if strings.HasPrefix(cluster, k8sEnvs.Environments[i].Prefix[j]) && envMatch {
				return true
			}
		}
		for j := 0; j < len(k8sEnvs.Environments[i].Suffix); j++ {
			if strings.HasSuffix(cluster, k8sEnvs.Environments[i].Suffix[j]) && envMatch {
				return true
			}
		}
	}
	return match
}

func homePath() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	}
	return os.Getenv(env)
}

func readKubeConfig(config *KubeConfig, path string) (err error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	fileContent, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(fileContent, config)
	if err != nil {
		return
	}

	return
}

func segmentKube(p *powerline) {
	paths := append(strings.Split(os.Getenv("KUBECONFIG"), ":"), path.Join(homePath(), ".kube", "config"))
	config := &KubeConfig{}
	for _, configPath := range paths {
		temp := &KubeConfig{}
		if readKubeConfig(temp, configPath) == nil {
			config.Contexts = append(config.Contexts, temp.Contexts...)
			if config.CurrentContext == "" {
				config.CurrentContext = temp.CurrentContext
			}
		}
	}

	cluster := ""
	//namespace := ""
	//environment := ""
	for _, context := range config.Contexts {
		if context.Name == config.CurrentContext {
			cluster = context.Name
			//namespace = context.Context.Namespace
			break
		}
	}

	prod := false
	staging := false
	qat := false
	sandbox := false
	dev := false
	label_b := ""
	label_e := ""
	fg := p.theme.KubeClusterFg
	bg := p.theme.KubeClusterBg

	k8sEnvs := getEnvironments()

	// When you use gke your clusters may look something like gke_projectname_availability-zone_cluster-01
	// instead I want it to read as `cluster-01`
	// So we remove the first 3 segments of this string, if the flag is set, and there are enough segments
	if strings.HasPrefix(cluster, "gke") && *p.args.ShortenGKENames {
		segments := strings.Split(cluster, "_")
		if len(segments) > 3 {
			cluster = strings.Join(segments[3:], "_")
			//environment = segments[1]
		}
	}

	// With AWS EKS, cluster names are ARNs; it makes more sense to shorten them
	// so "eks-infra" instead of "arn:aws:eks:us-east-1:XXXXXXXXXXXX:cluster/eks-infra
	const arnRegexString string = "^arn:aws:eks:[[:alnum:]-]+:[[:digit:]]+:cluster/(.*)$"
	arnRe := regexp.MustCompile(arnRegexString)

	if arnMatches := arnRe.FindStringSubmatch(cluster); arnMatches != nil && *p.args.ShortenEKSNames {
		cluster = arnMatches[1]
	}

	if doesEnvMatch(cluster, &k8sEnvs, "PROD") {
		prod = true
		label_b = "PROD->>!!! "
		label_e = " !!!<<-PROD"
		fg = p.theme.KubeProdFg
		bg = p.theme.KubeProdBg
	} else if doesEnvMatch(cluster, &k8sEnvs, "Staging") {
		staging = true
		label_b = ""
		label_e = " STAGING"
		fg = p.theme.ShellVarFg
		bg = p.theme.ShellVarBg
	} else if doesEnvMatch(cluster, &k8sEnvs, "QAT") {
		qat = true
		label_b = ""
		label_e = " QAT"
		fg = p.theme.ShellVarFg
		bg = p.theme.ShellVarBg
		// fg = p.theme.GitNotStagedFg
		// bg = p.theme.GitNotStagedBg
	} else if doesEnvMatch(cluster, &k8sEnvs, "sandbox") {
		sandbox = true
		label_b = ""
		label_e = " SANDBOX"
		fg = p.theme.KubeClusterFg
		bg = p.theme.KubeClusterBg
	} else if doesEnvMatch(cluster, &k8sEnvs, "dev") {
		dev = true
		label_b = ""
		label_e = " DEV"
		fg = p.theme.KubeClusterFg
		bg = p.theme.KubeClusterBg
	}

	// Only draw the icon once
	//kubeIconHasBeenDrawnYet := false
	if cluster != "" {
		//kubeIconHasBeenDrawnYet = true
		if prod || staging || qat || sandbox || dev {
			for i := 0; i < len(k8sEnvs.Remove); i++ {
				cluster = strings.Replace(cluster, k8sEnvs.Remove[i], "", -1)
			}
			p.appendSegment("kube-cluster", pwl.Segment{
				Content:    fmt.Sprintf("%s%s%s", label_b, cluster, label_e),
				Foreground: fg,
				Background: bg,
			})
		}
	}

	// if namespace != "" {
	// 	content := namespace
	// 	if !kubeIconHasBeenDrawnYet {
	// 		content = fmt.Sprintf("⎈ %s", content)
	// 	}
	// 	p.appendSegment("kube-namespace", pwl.Segment{
	// 		Content:    content,
	// 		Foreground: p.theme.KubeNamespaceFg,
	// 		Background: p.theme.KubeNamespaceBg,
	// 	})
	// }
}



/*
sample of .kubectl_context_ps1_environments.yaml
---
environments:
  - type: PROD
    name: AWS PROD
    contains:
      - "111111111111"
    prefix:
      - np
    suffix:
  - type: Staging
    name: AWS Staging
    contains:
      - "222222222222"
    prefix:
    suffix:
  - type: sandbox
    name: AWS Sandbox
    contains:
      - "333333333333"
    prefix:
    suffix:
  - type: QAT
    name: QAT
    contains:
    prefix:
      - nq
    suffix:
  - type: dev
    name: dev
    contains:
    prefix:
      - nd
    suffix:
remove:
  - .k8s.net.com.yaml
  - arn:aws:eks:us-east-1:111111111111:cluster/
  - arn:aws:eks:us-west-1:222222222222:cluster/
  - arn:aws:eks:us-south-1:333333333333:cluster/
*/
