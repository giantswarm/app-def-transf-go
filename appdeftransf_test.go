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

			It("should not parse to any known type", func() {
				Expect(t).To(Equal(DefType("")))
			})
		})

		Describe("DefTypeV2GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"components": {
							"component/name/foo": {
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
					// wrong keys are "Components" and "portS"
					b := []byte(`{
						"Components": {
							"component/name/foo": {
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
						"components": {
							"component/name/foo": {
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
						"components": {
							"component/name/foo": {
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

		Describe("DefTypeV2GiantSwarm", func() {
			Describe("valid", func() {
				BeforeEach(func() {
					b := []byte(`{
						"components": {
							"component/name/foo": {
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
					Expect(n).To(Equal("8b75c8d7"))
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
