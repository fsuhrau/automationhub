package hub

import (
	"encoding/base64"
	"github.com/fsuhrau/automationhub/hub/action"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) printBody(c *gin.Context) {
	debugmap := map[string]interface{}{}
	c.Bind(&debugmap)
	s.logger.Debugf("body: %v\n", debugmap)
}

func (s *Server) renderError(c *gin.Context, err error) {
	s.logger.Errorf("%v", err)
	c.JSON(http.StatusNotFound, &ServerResponse{
		SessionID: "",
		State:     "Failed",
		Message:   err.Error(),
		HCode:     time.Now().UTC().Unix(),
	})
}

func (s *Server) GetGraph(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	a := &action.GetSceenGraph{}
	if err := s.deviceManager.SendAction(log, session, a); err != nil {
		s.renderError(c, err)
		return
	}

	c.String(http.StatusOK, a.Content())
}

func (s *Server) GetScreen(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	var data string
	var payload []byte

	if false {
		a := &action.GetScreenshot{}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		data = base64.StdEncoding.EncodeToString(a.ScreenshotData())
		payload = a.SceengraphXML()
	} else {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			a := &action.GetSceenGraph{}
			if err := s.deviceManager.SendAction(log, session, a); err != nil {
				s.renderError(c, err)
			}
			payload = []byte(a.Content())
			wg.Done()
		}()

		go func() {
			rawData, err := session.Lock.Device.GetScreenshot()
			if err != nil {
				s.logger.Errorf("Screenshot could not be created: %v", err)
				s.renderError(c, err)
			}
			data = base64.StdEncoding.EncodeToString(rawData)
			wg.Done()
		}()

		wg.Wait()
	}

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     data,
		Payload:   payload,
	})
}

func (s *Server) TakeScreenshot(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	var data string
	var payload []byte

	if false {
		a := &action.GetScreenshot{}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
		xmlPath, err := session.Storage.StoreSceneGraph(a.SceengraphXML())
		if err != nil {
			s.logger.Errorf("Store SceneGraph failed: %v", err)
		}

		pngPath, err := session.Storage.StoreImage(a.ScreenshotData())
		if err != nil {
			s.logger.Errorf("Store SceneImage failed: %v", err)
		}

		s.logger.Infof("Files:\n%s\n%s", xmlPath, pngPath)

		data = base64.StdEncoding.EncodeToString(a.ScreenshotData())
		payload = a.SceengraphXML()

	} else {
		a := &action.GetSceenGraph{}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		rawData, err := session.Lock.Device.GetScreenshot()
		if err != nil {
			s.logger.Errorf("Screenshot could not be created: %v", err)
		}
		_, err = session.Storage.StoreSceneGraph([]byte(a.Content()))
		if err != nil {
			s.logger.Errorf("Store SceneGraph failed: %v", err)
		}
		_, err = session.Storage.StoreImage(rawData)
		if err != nil {
			s.logger.Errorf("Store SceneImage failed: %v", err)
		}

		data = base64.StdEncoding.EncodeToString(rawData)
		payload = []byte(a.Content())
	}

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     data,
		Payload:   payload,
	})
}

func (s *Server) SetTimeouts(session *Session, c *gin.Context) {

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Server) RestartApp(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	type Request struct {
		Url string `json:"url"`
	}

	var req Request
	c.Bind(&req)

	if len(req.Url) > 0 {
		log.Debugf("Custom Command: %s", req.Url)
		a := &action.Custom{
			RequestData: req.Url,
		}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		c.JSON(http.StatusOK, &ServerResponse{
			SessionID: session.SessionID,
			State:     "success",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
		})
	} else {
		session.Lock.Device.StopApp(session.AppParameter)
		session.Lock.Device.StartApp(session.AppParameter, "", s.hostIP)
		session.WaitForConnection()

		c.JSON(http.StatusOK, &ServerResponse{
			SessionID: session.SessionID,
			State:     "success",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
		})
	}
}

func (s *Server) NavigateBack(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	if session.Lock.Device.HasFeature("back") {
		session.Lock.Device.Execute("back")
	} else {
		a := &action.Custom{
			RequestData: "back",
		}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}
