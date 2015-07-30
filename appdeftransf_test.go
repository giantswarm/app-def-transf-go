package appdeftransf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestMigrateV1ToV2(t *testing.T) {
	v1AppDef := userconfig.AppDefinition{
		AppName: "app_name",
		Services: []userconfig.ServiceConfig{
			userconfig.ServiceConfig{
				ServiceName: "service_name",
				Components: []userconfig.ComponentConfig{
					userconfig.ComponentConfig{
						ComponentName: "component_name",
						InstanceConfig: userconfig.InstanceConfig{
							Image: generictypes.MustParseDockerImage("image"),
							Dependencies: []userconfig.DependencyConfig{
								userconfig.DependencyConfig{
									Name:  "component_name2",
									Port:  generictypes.MustParseDockerPort("80/tcp"),
									Alias: "myalias",
								},
							},
						},
					},
					userconfig.ComponentConfig{
						ComponentName: "component_name2",
						InstanceConfig: userconfig.InstanceConfig{
							Image: generictypes.MustParseDockerImage("image"),
							Ports: []generictypes.DockerPort{
								generictypes.MustParseDockerPort("80/tcp"),
							},
						},
					},
				},
			},
		},
	}

	v2AppDef, err := V1GiantSwarmToV2GiantSwarm(v1AppDef)
	if err != nil {
		t.Fatalf("V1GiantSwarmToV2GiantSwarm failed: %#v", err)
	}

	if err := v2AppDef.Validate(nil); err != nil {
		t.Fatalf("v2AppDef.Validate failed: %#v", err)
	}

	node, err := v2AppDef.Nodes.NodeByName("service_name/component_name")
	if err != nil {
		t.Fatalf("node service_name/component_name node found: %#v", err)
	}

	if len(node.Links) != 1 {
		t.Fatalf("node.Links should contain 1 link, got %v", len(node.Links))
	}
	if !node.Links[0].Node.Equals("service_name/component_name2") {
		t.Fatalf("node.Links[0].Node should be service_name/component_name2, got %s", node.Links[0].Node)
	}
	if !node.Links[0].TargetPort.Equals(generictypes.MustParseDockerPort("80/tcp")) {
		t.Fatalf("node.Links[0].TargetPort should be 80/tcp, got %#v", node.Links[0].TargetPort)
	}
	if node.Links[0].Alias != "myalias" {
		t.Fatalf("node.Links[0].Alias should be myalias, got %s", node.Links[0].Alias)
	}
}

func TestMigratePodsV1ToV2(t *testing.T) {
	v1AppDef := userconfig.AppDefinition{
		AppName: "app_name",
		Services: []userconfig.ServiceConfig{
			userconfig.ServiceConfig{
				ServiceName: "service_name",
				Components: []userconfig.ComponentConfig{
					userconfig.ComponentConfig{
						ComponentName: "component_name1",
						InstanceConfig: userconfig.InstanceConfig{
							Image: generictypes.MustParseDockerImage("image"),
						},
						PodConfig: userconfig.PodConfig{
							PodName: "pod1",
						},
					},
					userconfig.ComponentConfig{
						ComponentName: "component_name2",
						InstanceConfig: userconfig.InstanceConfig{
							Image: generictypes.MustParseDockerImage("image"),
						},
						PodConfig: userconfig.PodConfig{
							PodName: "pod1",
						},
					},
				},
			},
		},
	}

	v2AppDef, err := V1GiantSwarmToV2GiantSwarm(v1AppDef)
	if err != nil {
		t.Fatalf("V1GiantSwarmToV2GiantSwarm failed: %#v", err)
	}

	if err := v2AppDef.Validate(nil); err != nil {
		t.Fatalf("v2AppDef.Validate failed: %#v", err)
	}
}

var _ = Describe("AppDefTransf", func() {
	var (
		err error
		n   string
		t   DefType
	)

	BeforeEach(func() {
		err = nil
		t = ""
		n = ""
	})

	Describe("ParseTypeFromBytes()", func() {
		Describe("invalid", func() {
			BeforeEach(func() {
				b := []byte(`{
					"foo": "bar"
				}`)

				t, err = ParseTypeFromBytes(b)
			})

			It("should throw error", func() {
				Expect(err).NotTo(BeNil())
				Expect(IsInvalidDefType(err)).To(BeTrue())
			})

			It("should not parse to any known type", func() {
				Expect(t).To(Equal(DefType("")))
			})
		})

		Describe("DefTypeV1GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"app_name": "an",
						"services": [
							{
								"service_name": "sn",
								"components": [
									{ "component_name": "cn" }
								]
							}
						]
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV1GiantSwarm))
				})
			})

			Describe("broken", func() {
				BeforeEach(func() {
					// wrong keys are "Component_name" and "appname"
					// spaces ensure edge case regex parsing works
					b := []byte(`{
						"appname": "an",
						"services"	 : [
							{
								"service_name": "sn",
								"components": [
									{ "Component_name": "cn" }
								]
							}
						]
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV1GiantSwarm))
				})
			})

			Describe("broken JSON", func() {
				BeforeEach(func() {
					// the first comma (,) is missing
					// spaces ensure edge case regex parsing works
					b := []byte(`{
						"appname": "an"
						"services"	: [
							{
								"service_name": "sn",
								"components": [
									{ "Component_name": "cn" }
								]
							}
						]
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV1GiantSwarm))
				})
			})

			Describe("unparsed variables", func() {
				BeforeEach(func() {
					b := []byte(`{
						"appname": "an",
						"services": [
							{
								"service_name": "$serviceName",
								"components": [
									{ "Component_name": "cn" }
								]
							}
						]
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV1GiantSwarm))
				})
			})
		})

		Describe("DefTypeV2GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"nodes": {
							"node/name/foo": {
								"image": "fancy/image:latest",
								"ports": [ 8080 ]
							}
						}
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV2GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV2GiantSwarm))
				})
			})

			Describe("broken field names", func() {
				BeforeEach(func() {
					// wrong keys are "Nodes" and "portS"
					b := []byte(`{
						"Nodes": {
							"node/name/foo": {
								"image": "fancy/image:latest",
								"portS": [ 8080 ]
							}
						}
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV2GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV2GiantSwarm))
				})
			})

			Describe("broken JSON, missing comma", func() {
				BeforeEach(func() {
					// the first comma (,) is missing
					b := []byte(`{
						"nodes": {
							"node/name/foo": {
								"image": "fancy/image:latest"
								"ports": [ 8080 ]
							}
						}
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV2GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV2GiantSwarm))
				})
			})

			Describe("unparsed variables", func() {
				BeforeEach(func() {
					b := []byte(`{
						"nodes": {
							"node/name/foo": {
								"image": "fancy/image:latest",
								"ports": [ $port ]
							}
						}
					}`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse def type DefTypeV2GiantSwarm", func() {
					Expect(t).To(Equal(DefTypeV2GiantSwarm))
				})
			})
		})
	})

	Describe("ParseName()", func() {
		Describe("DefTypeV1GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"app_name": "an",
						"services": [
							{
								"service_name": "sn",
								"components": [
									{ "component_name": "cn", "image": "foo" }
								]
							}
						]
					}`)

					n, err = ParseName(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse name", func() {
					Expect(n).To(Equal("an"))
				})
			})
		})

		Describe("DefTypeV2GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"nodes": {
							"node/name/foo": {
								"image": "fancy/image:latest",
								"ports": [ 8080 ]
							}
						}
					}`)

					n, err = ParseName(b)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse name", func() {
					Expect(n).To(Equal("63c100fb"))
				})
			})
		})

		Describe("invalid type", func() {
			BeforeEach(func() {
				b := []byte(`{
					"foo": "bar"
				}`)

				n, err = ParseName(b)
			})

			It("should throw error", func() {
				Expect(err).NotTo(BeNil())
				Expect(IsInvalidDefType(err)).To(BeTrue())
			})

			It("should NOT parse name", func() {
				Expect(n).To(BeEmpty())
			})
		})
	})
})
