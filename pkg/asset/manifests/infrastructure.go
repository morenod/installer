package manifests

import (
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/installer/pkg/types/aws"
	"github.com/openshift/installer/pkg/types/libvirt"
	"github.com/openshift/installer/pkg/types/none"
	"github.com/openshift/installer/pkg/types/openstack"
	"github.com/openshift/installer/pkg/types/vsphere"
)

var (
	infraCrdFilename = filepath.Join(manifestDir, "cluster-infrastructure-01-crd.yaml")
	infraCfgFilename = filepath.Join(manifestDir, "cluster-infrastructure-02-config.yml")
)

// Infrastructure generates the cluster-infrastructure-*.yml files.
type Infrastructure struct {
	FileList []*asset.File
}

var _ asset.WritableAsset = (*Infrastructure)(nil)

// Name returns a human friendly name for the asset.
func (*Infrastructure) Name() string {
	return "Infrastructure Config"
}

// Dependencies returns all of the dependencies directly needed to generate
// the asset.
func (*Infrastructure) Dependencies() []asset.Asset {
	return []asset.Asset{
		&installconfig.InstallConfig{},
	}
}

// Generate generates the Infrastructure config and its CRD.
func (i *Infrastructure) Generate(dependencies asset.Parents) error {
	installConfig := &installconfig.InstallConfig{}
	dependencies.Get(installConfig)

	var platform configv1.PlatformType
	switch installConfig.Config.Platform.Name() {
	case aws.Name:
		platform = configv1.AWSPlatform
	case none.Name:
		platform = configv1.NonePlatform
	case libvirt.Name:
		platform = configv1.LibvirtPlatform
	case openstack.Name:
		platform = configv1.OpenStackPlatform
	case vsphere.Name:
		platform = configv1.VSpherePlatform
	default:
		platform = configv1.NonePlatform
	}

	config := &configv1.Infrastructure{
		TypeMeta: metav1.TypeMeta{
			APIVersion: configv1.SchemeGroupVersion.String(),
			Kind:       "Infrastructure",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
			// not namespaced
		},
		Status: configv1.InfrastructureStatus{
			Platform:            platform,
			APIServerURL:        getAPIServerURL(installConfig.Config),
			EtcdDiscoveryDomain: getEtcdDiscoveryDomain(installConfig.Config),
		},
	}

	configData, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal config: %#v", config)
	}

	i.FileList = []*asset.File{
		{
			Filename: infraCfgFilename,
			Data:     configData,
		},
	}

	return nil
}

// Files returns the files generated by the asset.
func (i *Infrastructure) Files() []*asset.File {
	return i.FileList
}

// Load returns false since this asset is not written to disk by the installer.
func (i *Infrastructure) Load(f asset.FileFetcher) (bool, error) {
	return false, nil
}
