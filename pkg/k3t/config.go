package k3t

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type clusterKubeconfig struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			ClientCertificateData string `yaml:"client-certificate-data"`
			ClientKeyData         string `yaml:"client-key-data"`
		} `yaml:"user"`
	} `yaml:"users"`
}

type CertificateData struct {
	CertificateAuthorityData []byte
	ClientCertificateData    []byte
	ClientKeyData            []byte
}

func parseKubeconfig(data []byte) (CertificateData, error) {
	res := CertificateData{}
	config := clusterKubeconfig{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return res, errors.Wrap(err, "error unmarshalling kubeconfig")
	}
	cadStr := config.Clusters[0].Cluster.CertificateAuthorityData
	ccdStr := config.Users[0].User.ClientCertificateData
	ckdStr := config.Users[0].User.ClientKeyData

	cad, err := base64.StdEncoding.DecodeString(cadStr)
	if err != nil {
		return res, errors.Wrap(err, "error decoding CertificateAuthorityData")
	}
	res.CertificateAuthorityData = cad

	ccd, err := base64.StdEncoding.DecodeString(ccdStr)
	if err != nil {
		return res, errors.Wrap(err, "error decoding ClientCertificateData")
	}
	res.ClientCertificateData = ccd

	ckd, err := base64.StdEncoding.DecodeString(ckdStr)
	if err != nil {
		return res, errors.Wrap(err, "error decoding ClientKeyData")
	}
	res.ClientKeyData = ckd

	return res, nil
}
