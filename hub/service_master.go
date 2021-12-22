package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/device/unityeditor"
	"net/http"

	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (s *Service) RunMaster() error {
	//showRemlog := viper.GetBool("display_remlog")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if s.cfg.AutoDiscovery {
		go ZeroConfServer(ctx, "", s.hostIP.String())
	}
/*
	remoteLoggingService := remlog.NewService(s.logger)
	if err := remoteLoggingService.Run(func(msg string) {
		data := strings.Split(msg, "|")
		if len(data) > 1 {
			session, _ := s.sessionManager.GetSession(data[0])
			if session != nil {
				session.Storage.RemoteDeviceLog(data[1])
				if showRemlog {
					session.logger.Debug(data[1])
				}
			} else {
				if showRemlog {
					s.logger.Debug(msg)
				}
			}
		}
	}); err != nil {
		return err
	}
*/
	// start device observer thread
	if d, ok := s.cfg.DeviceManager[androiddevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager android_device")
		s.deviceManager.AddHandler(androiddevice.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[iossim.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_sim")
		s.deviceManager.AddHandler(iossim.NewHandler(s.sd, s.hostIP))
	}
	if d, ok := s.cfg.DeviceManager[iosdevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_device")
		s.deviceManager.AddHandler(iosdevice.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[macos.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager macos")
		s.deviceManager.AddHandler(macos.NewHandler(s.sd, s.hostIP))
	}
	if d, ok := s.cfg.DeviceManager[unityeditor.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager unity_editor")
		s.deviceManager.AddHandler(unityeditor.NewHandler(s.sd, s.hostIP))
	}

	if err := s.deviceManager.Run(ctx); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	s.sessionManager.Run(ctx)

	s.router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

//	go grpcConnect()

	runOn := ":8002"
	if s.cfg.Port != 0 {
		runOn = fmt.Sprintf(":%d", s.cfg.Port)
	}
	err := s.router.Run(runOn)
	logrus.Infof("Stopping Server")
	if err != nil {
		return err
	}

	return nil
}
/*
func grpcConnect() {
	l, err := net.Listen("tcp", ":8003")
	if err != nil {
		logrus.Errorf("listen: %w", err)
	}
	defer l.Close()
	for {
		slaveConnection, err := l.Accept()
		if err != nil {
			logrus.Errorf("accept: %v", err)
			continue
		}
		go grpcHandler(slaveConnection)
	}
}

func grpcHandler(outerConnection net.Conn) {
	defer outerConnection.Close()

	conn, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithContextDialer(func(c context.Context, s string) (net.Conn, error) {
		return outerConnection, nil
	}))
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := action.NewRemoteDeviceClient(conn)
	stream, err := client.DeviceList(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	ctx := stream.Context()
	done := make(chan bool)

	req := action.DeviceRequest{}
	if err := stream.Send(&req); err != nil {
		log.Fatalf("can not send %v", err)
	}

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			devices := resp.Devices
			logrus.Infof("devices: %s", devices)
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
}


*/