package main

import (
	"context"
	"testing"
)

func TestCluster(t *testing.T) {
	bg := context.Background()

	cluster := NewClusterFromEnv(7443)
	data, err := cluster.Start(bg)
	if err != nil {
		t.Error(err)
	}

	if len(data.CertificateAuthorityData) == 0 {
		t.Errorf("CertificateAuthorityData should not be empty")
	}

	if len(data.ClientCertificateData) == 0 {
		t.Errorf("ClientCertificateData should not be empty")
	}

	if len(data.ClientKeyData) == 0 {
		t.Errorf("ClientKeyData should not be empty")
	}

	cluster.logger.Info().
		Msg("shutting down")
	if err := cluster.Stop(bg); err != nil {
		t.Error(err)
	}
}
