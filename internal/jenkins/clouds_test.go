package jenkins

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const sampleRootConfig = `<?xml version='1.1' encoding='UTF-8'?>
<hudson>
  <clouds>
    <org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud plugin="kubernetes@4258.v1234">
      <name>kubernetes</name>
    </org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud>
    <io.jenkins.plugins.ec2.EC2Cloud>
      <name>ec2-builders</name>
    </io.jenkins.plugins.ec2.EC2Cloud>
  </clouds>
</hudson>`

func TestListClouds(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config.xml" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(sampleRootConfig))
	})
	defer stop()
	clouds, err := c.ListClouds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(clouds) != 2 {
		t.Fatalf("got %d clouds: %+v", len(clouds), clouds)
	}
	if clouds[0].Name != "kubernetes" || !strings.Contains(clouds[0].Kind, "Kubernetes") {
		t.Errorf("clouds[0]=%+v", clouds[0])
	}
	if clouds[1].Name != "ec2-builders" {
		t.Errorf("clouds[1]=%+v", clouds[1])
	}
}

func TestGetCloudConfig(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manage/cloud/kubernetes/config.xml" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`<KubernetesCloud><name>kubernetes</name></KubernetesCloud>`))
	})
	defer stop()
	b, err := c.GetCloudConfig(context.Background(), "kubernetes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<KubernetesCloud>") {
		t.Errorf("got %s", string(b))
	}
}
