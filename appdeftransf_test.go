package appdeftransf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppDefTransf", func() {
	var (
		err error
		t   DefType
	)

	BeforeEach(func() {
		err = nil
		t = ""
	})

	Describe("ParseTypeFromBytes()", func() {
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

			Describe("invalid", func() {
				BeforeEach(func() {
					b := []byte(`{
            "foo": "bar"
          }`)

					t, err = ParseTypeFromBytes(b)
				})

				It("should throw error", func() {
					Expect(err).NotTo(BeNil())
					Expect(err).To(Equal(invalidDefTypeErr))
				})

				It("should not parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefType("")))
				})
			})

			Describe("broken", func() {
				BeforeEach(func() {
					// wrong keys are "Component_name" and "appname"
					b := []byte(`{
            "appname": "an",
            "services": [
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
					b := []byte(`{
            "appname": "an"
            "services": [
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

				It("should throw error", func() {
					Expect(err).NotTo(BeNil())
					Expect(err).To(Equal(invalidDefTypeErr))
				})

				It("should not parse def type DefTypeV1GiantSwarm", func() {
					Expect(t).To(Equal(DefType("")))
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
	})
})
