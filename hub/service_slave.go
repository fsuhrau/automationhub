package hub

import (
	"context"
	"fmt"
	"net"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type grpcSlave struct {
}
/*
func (g grpcSlave) DeviceList(srv action.RemoteDevice_DeviceListServer) error {
	logrus.Infof("DeviceList")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logrus.Infof("waiting for message")
		_, err := srv.Recv()
		if err == io.EOF {
			logrus.Infof("EOF")
			return nil
		}
		if err != nil {
			logrus.Errorf("stream error: %v", err)
			return err
		}

		var deviceList []*action.Device

		resp := action.DeviceResponse{
			Devices: deviceList,
		}

		logrus.Infof("Send Device List")

		if err := srv.Send(&resp); err != nil {
			return err
		}
	}

	return nil
}

func (g grpcSlave) DeviceInteraction(srv action.RemoteDevice_DeviceInteractionServer) error {
	logrus.Infof("DeviceInteraction")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		resp := action.InteractionResponse{}

		if err := srv.Send(&resp); err != nil {
			return err
		}
	}

	return nil
}

func (g grpcSlave) ActionExecute(srv action.RemoteDevice_ActionExecuteServer) error {
	logrus.Infof("ActionExecute")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		resp := action.Response{}

		if err := srv.Send(&resp); err != nil {
			return err
		}
	}

	return nil
}
*/
func (s *Service) RunSlave() error {

	// showRemlog := viper.GetBool("display_remlog")

	var serviceConfig config.Service

	if err := viper.Unmarshal(&serviceConfig); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(serviceConfig.HostIP) > 0 {
		s.hostIP = net.ParseIP(serviceConfig.HostIP)
	}
	if s.hostIP == nil {
		s.hostIP = utils.GetOutboundIP()
	}

	if len(serviceConfig.ServerURL) == 0 {
		return fmt.Errorf("slave need a server url to connect to")
	}

	go ZeroConfServer(ctx, "", s.hostIP.String())

	// remoteLoggingService := remlog.NewService(s.logger)
	// if err := remoteLoggingService.Run(func(msg string) {
	// 	data := strings.Split(msg, "|")
	// 	if len(data) > 1 {
	// 		session, _ := s.sessionManager.GetSession(data[0])
	// 		if session != nil {
	// 			session.Storage.RemoteDeviceLog(data[1])
	// 			if showRemlog {
	// 				session.logger.Debug(data[1])
	// 			}
	// 		} else {
	// 			if showRemlog {
	// 				s.logger.Debug(msg)
	// 			}
	// 		}
	// 	}
	// }); err != nil {
	// 	return err
	// }

	// start device observer thread
	/*
	if d, ok := serviceConfig.DeviceManager[androiddevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager android_device")
		s.deviceManager.AddHandler(androiddevice.NewHandler())
	}
	if d, ok := serviceConfig.DeviceManager[iossim.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_sim")
		s.deviceManager.AddHandler(iossim.NewHandler(&d, s.hostIP))
	}
	if d, ok := serviceConfig.DeviceManager[iosdevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_device")
		s.deviceManager.AddHandler(iosdevice.NewHandler(&d))
	}
	if d, ok := serviceConfig.DeviceManager[macos.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager macos")
		s.deviceManager.AddHandler(macos.NewHandler(&d, s.hostIP))
	}
	if d, ok := serviceConfig.DeviceManager[unityeditor.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager unity_editor")
		s.deviceManager.AddHandler(unityeditor.NewHandler(&d, s.hostIP))
	}
	*/
	if err := s.deviceManager.Run(ctx); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	s.sessionManager.Run(ctx)

	rpxService := grpc.NewServer()
	// action.RegisterRemoteDeviceServer(rpxService, grpcSlave{})

	logrus.Infof("Connect to Master")
	masterConnection, err := net.Dial("tcp", serviceConfig.ServerURL)
	if err != nil {
		return err
	}
	defer masterConnection.Close()

	return rpxService.Serve(localListener{conn: masterConnection})
}

type localListener struct {
	conn net.Conn
}

func (l localListener) Accept() (net.Conn, error) {
	return l.conn, nil
}

func (l localListener) Close() error {
	return nil
}

func (l localListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}
