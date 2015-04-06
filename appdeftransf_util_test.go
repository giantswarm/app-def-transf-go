package appdeftransf_test

import (
	"github.com/giantswarm/app-def-transf-go"
	"github.com/giantswarm/app-def-transf-go/logger"

	"code.google.com/p/go-uuid/uuid"
	logPkg "github.com/op/go-logging"

	instancePkg "github.com/giantswarm/app-service/service/instance-service"
	ambassadorPkg "github.com/giantswarm/app-service/service/instance-service/ambassador"
	lbregisterPkg "github.com/giantswarm/app-service/service/instance-service/lb-register"
	presencePkg "github.com/giantswarm/app-service/service/instance-service/presence"
	storagePkg "github.com/giantswarm/app-service/service/instance-service/storage"
	userPkg "github.com/giantswarm/app-service/service/instance-service/user"
	schedulerPkg "github.com/giantswarm/app-service/service/scheduler-service"
	unitfileServicePkg "github.com/giantswarm/app-service/service/unitfile-service"
)

func GivenUnitfileService() *unitfileServicePkg.UnitfileService {
	if srv, err := unitfileServicePkg.NewUnitfileService(); err != nil {
		// Go Template error propably
		panic(err.Error())
	} else {
		return srv
	}
}

func GivenLocalInstanceServiceRepository() *instancePkg.LocalInstanceGroupRepository {
	return instancePkg.NewLocalInstanceGroupRepository()
}

func GivenLocalScheduler() *schedulerPkg.LocalScheduler {
	return schedulerPkg.NewLocalScheduler().(*schedulerPkg.LocalScheduler)
}

func GivenInstanceService(repo instancePkg.InstanceGroupRepositoryI, scheduler schedulerPkg.SchedulerService) instancePkg.InstanceServiceI {
	simpleIDFactory := func() string {
		return uuid.New()
	}
	unitfileSrv := GivenUnitfileService()

	logger := logPkg.MustGetLogger("instance-service-test")

	userSrv := userPkg.NewUserInstanceService(
		userPkg.UserInstanceServiceConfig{
			CoreosHostIp:          "${COREOS_IP}",
			PrivateDockerRegistry: "private.registry.acme.com",
			PublicDockerRegistry:  "public.registry.acme.com",
		},
		userPkg.UserInstanceServiceDependencies{
			UnitfileService: unitfileSrv,
		},
	)

	presenceSrv := presencePkg.NewPresenceInstanceService(
		presencePkg.PresenceInstanceServiceConfig{
			Registry:   "public.registry.acme.com",
			Image:      "giantswarm/presence",
			Version:    "1.0.0",
			DockerPort: "10000",
		},
		presencePkg.PresenceInstanceServiceDependencies{
			UnitfileService: unitfileSrv,
		},
	)

	ambassadorSrv := ambassadorPkg.NewAmbassadorInstanceService(
		ambassadorPkg.AmbassadorInstanceServiceConfig{
			UtilsPath: "/opt/coreos/",
			Registry:  "public.registry.acme.com",
			Image:     "giantswarm/ambassador",
			Version:   "10.100.0",
		},
		ambassadorPkg.AmbassadorInstanceServiceDependencies{
			UnitfileService: unitfileSrv,
		},
	)

	storageSrv := storagePkg.NewStorageInstanceService(
		storagePkg.StorageInstanceServiceConfig{
			UtilsPath:          "/opt/coreos/",
			InstanceVolumePath: "/media/ebs-volumes/",
			Registry:           "public.registry.acme.com",
			Image:              "giantswarm/storage-sidekick",
			Version:            "1.999.2",
			AWSAccessKey:       "aws-access-key",
			AWSSecretKey:       "aws-secret-key",
			AWSRegion:          "aws-region",
		},
		storagePkg.StorageInstanceServiceDependencies{
			UnitfileService: unitfileSrv,
		},
	)

	lbRegisterSrv := lbregisterPkg.NewLBRegisterService(
		lbregisterPkg.LBRegisterServiceConfig{
			Image:      "public.registry.acme.com/giantswarm/lb-register-sidekick",
			Version:    "1.0.2000",
			EtcdPath:   "/lb-register/prefix",
			DockerPort: "10000",
		},
		lbregisterPkg.LBRegisterServiceDependencies{
			UnitfileService: unitfileSrv,
		},
	)

	return instancePkg.NewInstanceService(
		instancePkg.InstanceServiceConfig{
			UnitfilePrefix: "tests",
		},
		instancePkg.InstanceServiceDependencies{
			IDFactory: simpleIDFactory,

			UserInstanceService:       userSrv,
			PresenceInstanceService:   presenceSrv,
			AmbassadorInstanceService: ambassadorSrv,
			StorageInstanceService:    storageSrv,
			LBRegisterInstanceService: lbRegisterSrv,

			Logger:                  logger,
			InstanceGroupRepository: repo,
			SchedulerService:        scheduler,
		},
	)
}

func GivenAppDefTransf(is instancePkg.InstanceServiceI) *appdeftransf.AppDefTransf {
	conf := appdeftransf.Conf{}

	deps := appdeftransf.Deps{
		InstanceService: is,
		Logger:          logger.NewLogger(logger.Conf{Name: "test"}),
	}

	return appdeftransf.NewAppDefTransf(conf, deps)
}
