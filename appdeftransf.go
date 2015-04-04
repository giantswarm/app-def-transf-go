package appdeftransf

import (
	"path/filepath"

	"github.com/giantswarm/app-service/service/instance-service"
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

type Conf struct{}

type Deps struct {
	Logger          *logging.Logger
	InstanceService instanceservice.InstanceServiceI
}

// AppDefTransf stands for App-Definition-Transformer and is responsible for
// transforming data structures to other data structures.
type AppDefTransf struct {
	Conf
	Deps
}

func NewAppDefTransf(c Conf, d Deps) *AppDefTransf {
	return &AppDefTransf{
		Conf: c,
		Deps: d,
	}
}

// TODO
func (adt *AppDefTransf) V1toV2(appDefV1 userconfig.AppDefinition) userconfig.V2AppDefinition {
	return userconfig.V2AppDefinition{}
}

type Metadata struct {
	Organization string
	Environment  string
}

func (adt *AppDefTransf) V2toInstanceGroups(metadata Metadata, nodes userconfig.V2AppDefinition) ([]*instanceservice.InstanceGroup, error) {
	// find scaling nodes
	scalingNodes := []userconfig.Node{}
	for _, node := range nodes {
		if node.IsScalingNode() {
			scalingNodes = append(scalingNodes, node)
		}
	}

	instanceGroups := []*instanceservice.InstanceGroup{}
	for _, node := range nodes {
		// ignore scaling nodes
		if node.IsScalingNode() {
			continue
		}

		// find desired scale by user defined scaling nodes
		scale := userconfig.ScalingPolicyConfig{}
		for _, scalingNode := range scalingNodes {
			if scalingNode.Name.String() == node.Name.String() {
				scale = scalingNode.Scale
				break
			}
		}

		// apply scale
		for i := 0; i < scale.MinScale(); i++ {
			cfg, err := adt.createInstanceConfig(metadata, node, nodes)
			if err != nil {
				return nil, errgo.Mask(err)
			}

			instanceGroup, err := adt.InstanceService.New(cfg)
			if err != nil {
				return nil, errgo.Mask(err)
			}

			instanceGroups = append(instanceGroups, instanceGroup)
		}
	}

	return instanceGroups, nil
}

func (adt *AppDefTransf) createInstanceConfig(metadata Metadata, node userconfig.Node, allNodes []userconfig.Node) (instanceservice.InstanceGroupConfig, error) {
	volumes, err := adt.volumeConfigToInstanceVolumes(node.Volumes)
	if err != nil {
		return instanceservice.InstanceGroupConfig{}, errgo.Mask(err)
	}

	cfg := instanceservice.InstanceGroupConfig{
		Metadata: instanceservice.Metadata{
			Company:     metadata.Organization,
			Environment: metadata.Environment,
			Application: node.Name.Root(),
			Service:     node.Name.Dir(),
			Component:   node.Name.Base(),
		},
		Container: instanceservice.ContainerConfig{
			Image:   node.Image,
			Args:    node.Args,
			Env:     node.Env,
			Volumes: volumes,

			Discovery: instanceservice.DiscoveryConfig{
				// NOTE: For now we simply use the componentId
				// if we change this, we also need to build the Dependencies differently below!
				Contexts: []string{context(metadata, node)},

				Domains: node.Domains,
				Ports:   node.Ports,

				Dependencies: adt.nodeLinksToInstanceDependencies(metadata, node, allNodes),
			},
		},
		Scheduling: instanceservice.SchedulingConfig{},
	}

	return cfg, nil
}

func (adt *AppDefTransf) volumeConfigToInstanceVolumes(volumes []userconfig.VolumeConfig) ([]instanceservice.Volume, error) {
	result := []instanceservice.Volume{}
	for _, vol := range volumes {
		size, err := vol.Size.SizeInGB()
		if err != nil {
			return nil, errgo.Mask(err)
		}
		result = append(result, instanceservice.Volume{
			MountPoint: vol.Path,
			Size:       size,
		})
	}
	return result, nil
}

func (adt *AppDefTransf) nodeLinksToInstanceDependencies(metadata Metadata, node userconfig.Node, nodes []userconfig.Node) []instanceservice.Dependency {
	lookup := map[string]userconfig.Node{}
	for _, n := range nodes {
		lookup[n.Name.String()] = n
	}

	// Create a copy and then modify the Dependencies array
	dependencies := []instanceservice.Dependency{}
	for _, link := range node.Links {
		depNode, ok := lookup[link.Node.String()]
		if !ok {
			panic("unknown dependency: " + link.Node.String())
		}

		name := link.Alias
		if name == "" {
			name = depNode.Name.Base()
		}

		newDep := instanceservice.Dependency{
			Context: context(metadata, node),
			Name:    name,
			Port:    link.Port,
		}
		dependencies = append(dependencies, newDep)
	}

	return dependencies
}

func context(metadata Metadata, node userconfig.Node) string {
	return filepath.Join(metadata.Organization, metadata.Environment, node.Name.String())
}
