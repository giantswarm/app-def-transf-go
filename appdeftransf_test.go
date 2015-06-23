package appdeftransf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

			It("should not parse def type DefTypeV1GiantSwarm", func() {
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
					Expect(n).To(Equal("ad7ca034"))
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
