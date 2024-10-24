package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
	"museum/util"
	"strconv"
	"syscall"
	"time"
)

type DockerApplicationProvisionerService struct {
	ExhibitService              service.ExhibitService
	LivecheckFactoryService     service.LivecheckFactoryService
	EnvironmentTemplateResolver service.EnvironmentTemplateResolverService
	Client                      *docker.Client
	LockService                 service.LockService
	RuntimeInfoService          service.RuntimeInfoService
	LastAccessedService         service.LastAccessedService
	Eventing                    persistence.Eventing
	Log                         *zap.SugaredLogger
	Provider                    trace.TracerProvider
	Config                      config.Config
	VolumeProvisionerFactory    service.VolumeProvisionerFactoryService
}

func (d DockerApplicationProvisionerService) startApplicationInsideLock(ctx context.Context, exhibit *domain.Exhibit) error {
	sortedObjects := exhibit.Objects
	if exhibit.Order != nil {
		sortedObjects = make([]domain.Object, 0)
		for _, s := range exhibit.Order {
			for _, o := range exhibit.Objects {
				if s == o.Name {
					sortedObjects = append(sortedObjects, o)
				}
			}
		}
	}

	containerNameMapping := make(map[string]string)

	networkInspect, err := d.Client.NetworkInspect(ctx, exhibit.Name, network.InspectOptions{})
	if err != nil {
		d.Log.Warnw("network not found, creating", "exhibit", exhibit.Name)

		if docker.IsErrNotFound(err) {
			_, err := d.Client.NetworkCreate(ctx, exhibit.Name, network.CreateOptions{
				Driver: "bridge",
			})
			if err != nil {
				d.Log.Errorw("error creating network", "exhibit", exhibit.Name, "error", err)
				return err
			}

			networkInspect, err = d.Client.NetworkInspect(ctx, exhibit.Name, network.InspectOptions{})
			if err != nil {
				d.Log.Errorw("error inspecting network", "exhibit", exhibit.Name, "error", err)
				return err
			}
		} else {
			d.Log.Errorw("error inspecting network", "exhibit", exhibit.Name, "error", err)
			return err
		}
	}

	stepCount := 1

	// create a container on the swarm for each object
	idx := 0
	for _, o := range sortedObjects {
		err := d.startExhibitObject(ctx, exhibit, o, networkInspect, idx, &stepCount, &containerNameMapping)
		if err != nil {
			//TODO: rollback
			d.Log.Errorw("error starting exhibit object", "exhibit", exhibit.Name, "object", o.Name, "error", err)
			return err
		}
		idx++
	}

	exhibit.RuntimeInfo.Status = domain.Running

	return nil
}

func (d DockerApplicationProvisionerService) startExhibitObject(ctx context.Context, exhibit *domain.Exhibit, object domain.Object, network network.Inspect, idx int, stepCount *int, templateContainer *map[string]string) error {
	containerImage := object.Image + ":" + object.Label
	containerConfig := &container.Config{
		Image: containerImage,
		Env:   make([]string, 0),
	}

	err, env := d.EnvironmentTemplateResolver.FillEnvironmentTemplate(exhibit, object, templateContainer)
	if err != nil {
		return err
	}

	for k, v := range env {
		containerConfig.Env = append(containerConfig.Env, k+"="+v)
	}

	name := exhibit.Name + "_" + object.Name

	d.Log.Debugw("starting object", "object", object.Name, "exhibit", exhibit.Name)

	ctx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "startApplicationInsideLock", trace.WithAttributes(attribute.String("container", name), attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	span.AddEvent("inspecting container")
	d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
		Object: idx,
		Step:   domain.ObjectStartingStepClean,
	})

	//check if container already exists
	inspect, err := d.Client.ContainerInspect(ctx, name)
	if err == nil && inspect.ID != "" {
		d.Log.Warnw("container already exists", "container", name, "exhibitId", exhibit.Id)

		err = d.doCleanup(inspect, exhibit, ctx)
		if err != nil {
			d.Log.Errorw("error cleaning up container", "container", name, "exhibitId", exhibit.Id, "error", err)
			d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
				Object: idx,
				Step:   domain.ObjectStartingStepClean,
				Error:  err,
			})
			return err
		}
	}

	span.AddEvent("creating container")
	d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
		Object: idx,
		Step:   domain.ObjectStartingStepCreate,
	})

	// setup container to be in docker network
	containerConfig.Hostname = name
	containerConfig.Domainname = object.Name + "." + exhibit.Name

	// setup container mounts
	if len(object.Mounts) != 0 {
		for containerVolume, containerMount := range object.Mounts {
			// find corresponding volume
			volume := domain.Volume{}
			for _, v := range exhibit.Volumes {
				if v.Name == containerVolume {
					volume = v
					break
				}
			}

			util.Nop(volume)
			util.Nop(containerMount)

			provisioner, err := d.VolumeProvisionerFactory.GetForDriverType(volume.Driver.Type)
			if err != nil {
				d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
					Object: idx,
					Step:   domain.ObjectStartingStepCreate,
					Error:  err,
				})
				return err
			}

			//TODO

			util.Nop(provisioner)
			containerConfig.Volumes = nil

			//containerConfig.Volumes

			//containerConfig.Volumes[containerMount], _ = provisioner.ProvisionStorage(ctx, volume.Driver.Config)
		}
	}

	create, err := d.Client.ContainerCreate(ctx, containerConfig, nil, nil, nil, name)
	if err != nil {
		d.Log.Errorw("error creating container", "container", name, "exhibitId", exhibit.Id, "error", err)
		d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
			Object: idx,
			Step:   domain.ObjectStartingStepCreate,
			Error:  err,
		})
		return err
	}

	err = d.Client.NetworkConnect(ctx, network.ID, create.ID, nil)
	if err != nil {
		d.Log.Errorw("error connecting container to network", "container", name, "exhibitId", exhibit.Id, "error", err)
		d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
			Object: idx,
			Step:   domain.ObjectStartingStepCreate,
			Error:  err,
		})
		return err
	}

	span.AddEvent("starting container")
	d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
		Object: idx,
		Step:   domain.ObjectStartingStepStart,
	})

	d.Log.Debugw("starting container", "container", name, "exhibitId", exhibit.Id)
	err = d.Client.ContainerStart(ctx, create.ID, container.StartOptions{})
	if err != nil {
		d.Log.Errorw("error starting container", "container", name, "exhibitId", exhibit.Id, "error", err)
		d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
			Object: idx,
			Step:   domain.ObjectStartingStepStart,
			Error:  err,
		})
		return err
	}

	if object.Name == exhibit.Expose {
		exhibit.RuntimeInfo.Hostname = name
	}

	if object.Livecheck != nil {
		span.AddEvent("doing livecheck")
		d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
			Object: idx,
			Step:   domain.ObjectStartingStepLivecheck,
		})

		err := d.doLivecheck(ctx, *exhibit, object)
		if err != nil {
			d.Log.Errorw("error doing livecheck", "exhibitId", exhibit.Id, "error", err)
			d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
				Object: idx,
				Step:   domain.ObjectStartingStepLivecheck,
				Error:  err,
			})
			return err
		}
	}

	exhibit.RuntimeInfo.RelatedContainers = append(exhibit.RuntimeInfo.RelatedContainers, create.ID)

	// get container ip
	inspect, err = d.Client.ContainerInspect(ctx, create.ID)
	if err != nil {
		d.Log.Errorw("error inspecting container", "container", name, "exhibitId", exhibit.Id, "error", err)
		d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
			Object: idx,
			Step:   domain.ObjectStartingStepReady,
			Error:  err,
		})
		return err
	}
	(*templateContainer)[object.Name] = inspect.NetworkSettings.Networks["bridge"].IPAddress

	d.Eventing.DispatchExhibitStartingEvent(ctx, *exhibit, stepCount, domain.ExhibitStartingStep{
		Object: idx,
		Step:   domain.ObjectStartingStepReady,
	})

	return nil
}

func (d DockerApplicationProvisionerService) doCleanup(inspect types.ContainerJSON, exhibit *domain.Exhibit, ctx context.Context) error {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "doCleanup("+inspect.Name+")", trace.WithAttributes(attribute.String("container", inspect.Name), attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	// container already exists, check if it's running
	if inspect.State.Running {
		span.AddEvent("stopping container")

		d.Log.Warnw("container is running, stopping", "container", inspect.Name, "exhibitId", exhibit.Id)

		// stop container
		err := d.Client.ContainerStop(subCtx, inspect.ID, container.StopOptions{})
		if err != nil {
			return err
		}
	}

	span.AddEvent("removing container")

	d.Log.Debugw("removing container", "container", inspect.Name, "exhibitId", exhibit.Id)

	// remove container
	err := d.Client.ContainerRemove(subCtx, inspect.ID, container.RemoveOptions{})
	if err != nil {
		return err
	}

	for _, volume := range exhibit.Volumes {
		provisioner, e := d.VolumeProvisionerFactory.GetForDriverType(volume.Driver.Type)
		if e != nil {
			return err
		}

		e = provisioner.DeprovisionStorage(subCtx, volume.Driver.Config)
		if e != nil {
			return err
		}
	}

	return nil
}

func (d DockerApplicationProvisionerService) doLivecheck(ctx context.Context, exhibit domain.Exhibit, object domain.Object) error {
	var err error = nil
	runtimeInfoCopy := *exhibit.RuntimeInfo
	exhibit.RuntimeInfo = &runtimeInfoCopy

	livecheck := d.LivecheckFactoryService.GetLivecheckService(object.Livecheck.Type)
	if livecheck == nil {
		return errors.New("livecheck type not found")
	}

	maxRetries := 10
	if r, ok := object.Livecheck.Config["maxRetries"]; ok {
		maxRetries, err = strconv.Atoi(r)
		if err != nil {
			return err
		}
	}

	interval := 1 * time.Second
	if i, ok := object.Livecheck.Config["interval"]; ok {
		interval, err = time.ParseDuration(i)
		if err != nil {
			return err
		}
	}

	retry := true
	for counter := 0; (retry && err == nil) && counter < maxRetries; counter++ {
		retry, err = livecheck.Check(ctx, exhibit, object)

		// if the error is that the connection was refused, we can retry
		if err != nil && errors.Is(err, syscall.ECONNREFUSED) {
			err = nil
		}

		if counter != 0 {
			time.Sleep(interval)
		}
	}

	if retry || err != nil {
		return errors.New("livecheck failed")
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationStartingStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationStartingStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	span.AddEvent("acquiring runtime_info lock")

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	span.AddEvent("runtime_info lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	span.AddEvent("checking exhibit status")

	// check that exhibit is not already started after lock is acquired
	if exhibit.RuntimeInfo.Status == domain.Running {
		return nil
	}

	if exhibit.RuntimeInfo.Status != domain.Stopped && exhibit.RuntimeInfo.Status != domain.NotCreated {
		return errors.New(string("cannot start application in state " + exhibit.RuntimeInfo.Status))
	}

	span.AddEvent("setting exhibit status to starting")

	exhibit.RuntimeInfo.Status = domain.Starting
	exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)

	err = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	span.AddEvent("exhibit status set to starting")

	err = d.LastAccessedService.SetLastAccessed(subCtx, exhibitId, time.Now().Unix())
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationRunningStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationRunningStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	span.AddEvent("acquiring runtime_info lock")

	exhibitRlock := d.LockService.GetRwLock(subCtx, exhibitId, "exhibit")
	err = exhibitRlock.RLock()
	if err != nil {
		d.Log.Errorw("error locking exhibit", "exhibitId", exhibitId, "error", err)
		return err
	}

	span.AddEvent("exhibit lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.RUnlock()
	}(exhibitRlock)

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	span.AddEvent("acquiring runtime_info lock")

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	span.AddEvent("runtime_info lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	err = d.startApplicationInsideLock(subCtx, &exhibit)
	if err != nil {
		span.AddEvent("error starting application, reverting status to stopped")

		exhibit.RuntimeInfo.Status = domain.Stopped
		exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)
		_ = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
		return err
	}

	return d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
}

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "StartApplication", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	err := d.applicationStartingStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error starting application", "exhibitId", exhibitId, "error", err)
		return err
	}

	err = d.applicationRunningStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error starting application", "exhibitId", exhibitId, "error", err)
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationStoppingStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationStoppingStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	d.Eventing.DispatchExhibitStoppingEvent(subCtx, exhibit)

	span.AddEvent("acquiring runtime_info lock")

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	span.AddEvent("runtime_info lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	span.AddEvent("checking exhibit status")

	// check that exhibit is not already stopped after lock is acquired
	if exhibit.RuntimeInfo.Status == domain.Stopped {
		return nil
	}

	if exhibit.RuntimeInfo.Status != domain.Running {
		return errors.New(string("cannot stop application in state " + exhibit.RuntimeInfo.Status))
	}

	span.AddEvent("setting exhibit status to stopping")

	exhibit.RuntimeInfo.Status = domain.Stopping
	err = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationStoppedStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationStoppedStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	span.AddEvent("acquiring runtime_info lock")

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	span.AddEvent("runtime_info lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	for _, c := range exhibit.RuntimeInfo.RelatedContainers {
		span.AddEvent("stopping container " + c)

		err = d.Client.ContainerStop(subCtx, c, container.StopOptions{})
		if docker.IsErrNotFound(err) {
			span.AddEvent("container not found, skipping")
			continue
		}

		if err != nil {
			return err
		}
	}

	span.AddEvent("setting exhibit status to stopped")

	exhibit.RuntimeInfo.Status = domain.Stopped
	err = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) StopApplication(ctx context.Context, exhibitId string) error {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "CleanupApplication", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	err := d.applicationStoppingStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error stopping application", "exhibitId", exhibitId, "error", err)
		return err
	}

	err = d.applicationStoppedStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error stopping application", "exhibitId", exhibitId, "error", err)
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) CleanupApplication(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "CleanupApplication", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	span.AddEvent("acquiring runtime_info lock")

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	span.AddEvent("runtime_info lock acquired")

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	// check that exhibit is stopped+
	if exhibit.RuntimeInfo.Status != domain.Stopped {
		return errors.New(string("cannot cleanup application in state " + exhibit.RuntimeInfo.Status))
	}

	for _, containerId := range exhibit.RuntimeInfo.RelatedContainers {
		inspect, err := d.Client.ContainerInspect(subCtx, containerId)
		if docker.IsErrNotFound(err) {
			span.AddEvent("container not found, skipping")
			continue
		}

		if err != nil {
			return err
		}

		err = d.doCleanup(inspect, &exhibit, subCtx)
		if err != nil {
			return err
		}
	}

	networks, err := d.Client.NetworkList(subCtx, network.ListOptions{})
	if err != nil {
		d.Log.Errorw("error listing networks", "error", err)
		networks = make([]network.Inspect, 0)
	}

	for _, summary := range networks {
		if summary.Name == exhibit.Name {
			err = d.Client.NetworkRemove(subCtx, summary.ID)
			if err != nil {
				d.Log.Errorw("error removing network", "error", err)
			}
			break
		}
	}

	span.AddEvent("resetting runtime_info")
	exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)
	exhibit.RuntimeInfo.Hostname = ""

	err = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	return nil
}
