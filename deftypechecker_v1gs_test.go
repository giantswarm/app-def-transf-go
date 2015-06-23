package appdeftransf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("v1GiantSwarmDefTypeChecker", func() {
	var (
		t    DefType
		prob float64
		dtc  defTypeChecker
	)

	BeforeEach(func() {
		t = ""
		prob = 0.0
		dtc = newV1GiantSwarmDefTypeChecker()
	})

	Describe("Parse()", func() {
		Describe("valid definition", func() {
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV1GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV1GiantSwarm))
			})

			It("should parse probability of 100", func() {
				Expect(prob).To(Equal(float64(100)))
			})
		})

		Describe("invalid definition", func() {
			BeforeEach(func() {
				b := []byte(`{
					"foo": "bar"
				}`)

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV1GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV1GiantSwarm))
			})

			It("should parse probability of 0", func() {
				Expect(prob).To(Equal(float64(0)))
			})
		})

		Describe("broken definition", func() {
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV1GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV1GiantSwarm))
			})

			It("should parse probability of 80", func() {
				Expect(prob).To(Equal(float64(80)))
			})
		})

		Describe("broken DefTypeV1GiantSwarm JSON", func() {
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV1GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV1GiantSwarm))
			})

			It("should parse probability of 10", func() {
				Expect(prob).To(Equal(float64(10)))
			})
		})

		Describe("unparsed variables", func() {
			BeforeEach(func() {
				b := []byte(`{
					"app_name": "an",
					"services": [
						{
							"service_name": $serviceName,
							"components": [
								{ "component_name": "cn" }
							]
						}
					]
				}`)

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV1GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV1GiantSwarm))
			})

			It("should parse probability of 60", func() {
				Expect(prob).To(Equal(float64(60)))
			})
		})
	})
})
