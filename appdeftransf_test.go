package appdeftransf_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/app-def-transf-go"
	"github.com/giantswarm/app-service/service/instance-service"
	"github.com/giantswarm/app-service/service/scheduler-service"
	"github.com/giantswarm/user-config"
)

func TestAppDefTransf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "appdeftransf")
}

var _ = Describe("appdeftransf", func() {
	var (
		err      error
		igs      []*instanceservice.InstanceGroup
		metadata appdeftransf.Metadata
		nodes    userconfig.V2AppDefinition

		instanceSrv    instanceservice.InstanceServiceI
		instanceRepo   *instanceservice.LocalInstanceGroupRepository
		localScheduler *schedulerservice.LocalScheduler
		appDefTransf   *appdeftransf.AppDefTransf
	)

	BeforeEach(func() {
		err = nil
		nodes = nil
		metadata = appdeftransf.Metadata{Organization: "test-org", Environment: "test-env"}

		instanceRepo = GivenLocalInstanceServiceRepository()
		localScheduler = GivenLocalScheduler()
		instanceSrv = GivenInstanceService(instanceRepo, localScheduler)
		appDefTransf = GivenAppDefTransf(instanceSrv)
	})

	Describe("current weather example", func() {
		BeforeEach(func() {
			rawDef := `[
        {
          "name": "v2currentweather/webserver",
          "image": "registry.giantswarm.io/giantswarm/currentweather:latest",
          "ports": [ 8080 ],
          "links": [
            { "node": "v2currentweather/redis", "port": 6379 }
          ],
          "domains": {
            "v2currentweather.definition-v2.giantswarm.io": 8080
          }
        },
        {
          "name": "v2currentweather/redis",
          "image": "redis:latest",
          "ports": [ 6379 ]
        }
      ]`

			umErr := json.Unmarshal([]byte(rawDef), &nodes)
			if umErr != nil {
				panic(umErr)
			}

			igs, err = appDefTransf.V2toInstanceGroups(metadata, nodes)
		})

		It("should not throw error", func() {
			Expect(err).To(BeNil())
		})

		It("should generate 2 instance groups", func() {
			Expect(igs).To(HaveLen(2))
		})

		It("should generate 1 redis instance groups", func() {
			count := 0

			for _, ig := range igs {
				if ig.Metadata.Component == "redis" {
					count++
				}
			}

			Expect(count).To(Equal(1))
		})

		It("should generate 1 webserver instance groups", func() {
			count := 0

			for _, ig := range igs {
				if ig.Metadata.Component == "webserver" {
					count++
				}
			}

			Expect(count).To(Equal(1))
		})
	})

	Describe("scaling nodes", func() {
		BeforeEach(func() {
			rawDef := `[
        {
          "name": "v2currentweather",
          "scale": { "min": 2 }
        },
        {
          "name": "v2currentweather/webserver",
          "scale": { "min": 3 }
        },
        {
          "name": "v2currentweather/webserver",
          "image": "registry.giantswarm.io/giantswarm/currentweather:latest",
          "ports": [ 8080 ],
          "links": [
            { "node": "v2currentweather/redis", "port": 6379 }
          ],
          "domains": {
            "v2currentweather.definition-v2.giantswarm.io": 8080
          }
        },
        {
          "name": "v2currentweather/redis",
          "image": "redis:latest",
          "ports": [ 6379 ]
        }
      ]`

			umErr := json.Unmarshal([]byte(rawDef), &nodes)
			if umErr != nil {
				panic(umErr)
			}

			igs, err = appDefTransf.V2toInstanceGroups(metadata, nodes)
		})

		It("should not throw error", func() {
			Expect(err).To(BeNil())
		})

		It("should generate 7 instance groups", func() {
			Expect(igs).To(HaveLen(7))
		})

		It("should generate 2 redis instance groups", func() {
			count := 0

			for _, ig := range igs {
				if ig.Metadata.Component == "redis" {
					count++
				}
			}

			Expect(count).To(Equal(2))
		})

		It("should generate 5 webserver instance groups", func() {
			count := 0

			for _, ig := range igs {
				if ig.Metadata.Component == "webserver" {
					count++
				}
			}

			Expect(count).To(Equal(5))
		})
	})
})
