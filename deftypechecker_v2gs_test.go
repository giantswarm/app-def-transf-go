package appdeftransf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("v2GiantSwarmDefTypeChecker", func() {
	var (
		t    DefType
		prob float64
		dtc  defTypeChecker
	)

	BeforeEach(func() {
		t = ""
		prob = 0.0
		dtc = newV2GiantSwarmDefTypeChecker()
	})

	Describe("Parse()", func() {
		Describe("valid definition", func() {
			BeforeEach(func() {
				b := []byte(`{
					"components": {
						"component/name/foo": {
							"image": "fancy/image:latest",
							"ports": [ 8080 ]
						}
					}
				}`)

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV2GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV2GiantSwarm))
			})

			It("should parse probability of 57", func() {
				Expect(prob).To(Equal(float64(57)))
			})
		})

		Describe("invalid definition", func() {
			BeforeEach(func() {
				b := []byte(`{
					"foo": "bar"
				}`)

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV2GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV2GiantSwarm))
			})

			It("should parse probability of 0", func() {
				Expect(prob).To(Equal(float64(0)))
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV2GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV2GiantSwarm))
			})

			It("should parse probability of 57", func() {
				Expect(prob).To(Equal(float64(57)))
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV2GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV2GiantSwarm))
			})

			It("should parse probability of 10", func() {
				Expect(prob).To(Equal(float64(10)))
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

				t, prob = dtc.Parse(b)
			})

			It("should parse def type DefTypeV2GiantSwarm", func() {
				Expect(t).To(Equal(DefTypeV2GiantSwarm))
			})

			It("should parse probability of 60", func() {
				Expect(prob).To(Equal(float64(60)))
			})
		})
	})
})
