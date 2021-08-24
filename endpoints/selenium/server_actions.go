package selenium

import (
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/spf13/viper"

	"github.com/antchfx/xmlquery"
	"github.com/gin-gonic/gin"
)

var CacheXMLFile = false
var LookupRetry = 50 * time.Millisecond

const (
	SUCCESS_STRING                      = "success"
	SUCCESS                             = 0
	NO_SUCH_SESSION                     = 6
	NO_SUCH_ELEMENT                     = 7
	NO_SUCH_FRAME                       = 8
	UNKNOWN_COMMAND                     = 9
	STALE_ELEMENT_REFERENCE             = 10
	ELEMENT_NOT_VISIBLE                 = 11
	INVALID_ELEMENT_STATE               = 12
	UNHANDLED_ERROR                     = 13
	ELEMENT_NOT_SELECTABLE              = 15
	JAVASCRIPT_ERROR                    = 17
	XPATH_LOOKUP_ERROR                  = 19
	TIMEOUT                             = 21
	NO_SUCH_WINDOW                      = 23
	INVALID_COOKIE_DOMAIN               = 24
	UNABLE_TO_SET_COOKIE                = 25
	UNEXPECTED_ALERT_PRESENT            = 26
	NO_ALERT_PRESENT                    = 27
	ASYNC_SCRIPT_TIMEOUT                = 28
	INVALID_ELEMENT_COORDINATES         = 29
	IME_NOT_AVAILABLE                   = 30
	IME_ENGINE_ACTIVATION_FAILED        = 31
	INVALID_SELECTOR_ERROR              = 32
	SESSION_NOT_CREATED                 = 33
	MOVE_TARGET_OUT_OF_BOUNDS           = 34
	INVALID_XPATH_SELECTOR              = 51
	INVALID_XPATH_SELECTOR_RETURN_TYPER = 52
	ELEMENT_NOT_INTERACTABLE            = 60
	INVALID_ARGUMENT                    = 61
	NO_SUCH_COOKIE                      = 62
	UNABLE_TO_CAPTURE_SCREEN            = 63
)

func (s *SeleniumService) getElements(session *Session) error {

	if viper.GetBool("use_element_cache") && session.XmlDocument != nil {
		return nil
	}

	log := session.GetLogger().WithField("prefix", "action")

	a := &action.GetSceenGraph{}
	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		return err
	}
	var err error
	session.XmlDocument, err = a.XML()
	if err != nil {
		log.Infof("Could not Convert Content to XML: %v", err)
	}
	return nil
}

func (s *SeleniumService) GetElements(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	action := Request{}
	c.Bind(&action)

	until := session.GetActionTimeout()

	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")

			c.JSON(http.StatusOK, &Response{
				SessionID: session.GetSessionID(),
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		log.Debugf("locate element Using: %s Value: %s", action.Using, action.Value)

		if err := s.getElements(session); err != nil {
			// elements commented
			s.renderError(c, err)
			return
		}
		if session.XmlDocument == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if action.Using == "xpath" {

			elements := xmlquery.Find(session.XmlDocument, action.Value)
			if len(elements) > 0 {
				var ids []map[string]string
				for _, e := range elements {
					ids = append(ids, map[string]string{
						"ELEMENT": e.SelectAttr("ID"),
					})
				}
				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value:     ids,
				})
				return
			}

		} else {

			attribute := action.Using
			if action.Using == "name" {
				attribute = "LabelText"
			} else if action.Using == "tag name" {
				attribute = "Name"
			}
			log.Debugf("%v", action)
			elements := xmlquery.Find(session.XmlDocument, fmt.Sprintf("//*[@%s='%s']", attribute, action.Value))
			if len(elements) > 0 {
				var ids []map[string]string
				for _, e := range elements {
					ids = append(ids, map[string]string{
						"ELEMENT": e.SelectAttr("ID"),
					})
				}

				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value:     ids,
				})
				return
			}

		}
		log.Debugf("element not found retry with fresh one")
		session.XmlDocument = nil
		time.Sleep(50 * time.Millisecond)
	}
}

func (s *SeleniumService) ElementIsDisplayed(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	elementID := c.Param("elementID")

	a := &action.IsDisplayed{
		ElementID: elementID,
	}
	err := s.devicesManager.SendAction(log, session, a)
	if err != nil {
		s.renderError(c, err)
		return
	}

	log.Debugf("Is Displayed response: %v", a.IsDisplayed)

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    SUCCESS,
		Value:     a.IsDisplayed,
	})
}

func (s *SeleniumService) ElementClick(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	elementID := c.Param("elementID")

	until := session.GetActionTimeout()

	var result string
	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")
			c.JSON(http.StatusOK, &Response{
				SessionID: session.GetSessionID(),
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		a := &action.TouchElement{ElementID: elementID}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		if a.Success {
			result = "success"
			break
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}

func (s *SeleniumService) ElementSetValue(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type ActionRequest struct {
		ID    string   `json:"id"`
		Value []string `json:"value"`
	}

	req := &ActionRequest{}

	c.Bind(req)

	elementID := c.Param("elementID")
	a := &action.SetValue{
		ElementID: elementID,
		Attr:      "text",
		Value:     strings.Join(req.Value, ""),
	}

	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		s.renderError(c, err)
		return
	}

	result := ""
	if a.Success {
		result = "sentValue"
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}

func (s *SeleniumService) ElementGetValue(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")
	attr := c.Param("attribute")
	elementID := c.Param("elementID")

	a := &action.GetValue{
		ElementID: elementID,
		Attr:      attr,
	}

	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		s.renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     a.Value,
	})
}

func (s *SeleniumService) ElementGetText(session *Session, c *gin.Context) {
	c.Params = append(c.Params, gin.Param{"attribute", "text"})
	if viper.GetBool("innium_support") {
		s.ElementGetValue(session, c)
	} else {
		s.ElementGetAttribute(session, c)
	}
}

func (s *SeleniumService) ElementGetCSS(session *Session, c *gin.Context) {
	c.Params = append(c.Params, gin.Param{"attribute", "css"})
	if viper.GetBool("innium_support") {
		s.ElementGetValue(session, c)
	} else {
		s.ElementGetAttribute(session, c)
	}
}

func (s *SeleniumService) ElementGetName(session *Session, c *gin.Context) {
	c.Params = append(c.Params, gin.Param{"attribute", "Name"})
	s.ElementGetAttribute(session, c)
}

func mapSeleniumAttr(attr string) string {
	switch attr {
	case "id":
	case "href":
		attr = "ID"
	case "css":
		attr = "CSS"
	case "text":
		attr = "Text"
	case "name":
	case "tag":
		attr = "Name"
	case "class":
	case "type":
		attr = "Class"
	case "value":
		attr = "LabelText"
	}
	return attr
}

func (s *SeleniumService) ElementGetAttribute(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")
	attr := c.Param("attribute")
	elementID := c.Param("elementID")

	if attr == "displayed" {
		s.ElementIsDisplayed(session, c)
		return
	}

	attr = mapSeleniumAttr(attr)

	action := Request{}
	c.Bind(&action)

	var element *xmlquery.Node

	until := session.GetActionTimeout()

	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")

			c.JSON(http.StatusOK, &Response{
				SessionID: session.GetSessionID(),
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}
		log.Debugf("locate element Using: %s Value: %s", action.Using, action.Value)

		if err := s.getElements(session); err != nil {
			// elements commented
			s.renderError(c, err)
			return
		}
		if session.XmlDocument == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if action.Using == "xpath" {
			element = xmlquery.FindOne(session.XmlDocument, action.Value)
			if element != nil {
				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value:     element.SelectAttr(attr),
				})
				return
			}

		} else {
			element := xmlquery.FindOne(session.XmlDocument, fmt.Sprintf("//*[@ID='%s']", elementID))
			if element != nil {
				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value:     element.SelectAttr(attr),
				})
				return
			}
		}
		log.Debugf("element not found retry with fresh one")
		session.XmlDocument = nil
		time.Sleep(50 * time.Millisecond)
	}
}

type MoveElementToElement struct {
	From string
}

var moveElementToElement *MoveElementToElement
var elementID string

func (s *SeleniumService) MoveTo(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type Move struct {
		XOffset int64  `json:"xoffset"`
		YOffset int64  `json:"yoffset"`
		Element string `json:"element"`
	}

	move := Move{}
	c.Bind(&move)

	if move.XOffset != 0 && move.YOffset != 0 {
		a := &action.Scroll{
			OffsetX: move.XOffset,
			OffsetY: move.YOffset,
		}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
	} else {
		if moveElementToElement == nil {
			moveElementToElement = &MoveElementToElement{}
		}
		log.Debugf("MoveTo: %s\n", move.Element)
		elementID = move.Element
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) ButtonDown(session *Session, c *gin.Context) {

	if moveElementToElement != nil {
		moveElementToElement.From = elementID
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) ButtonUp(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	a := &action.DragAndDrop{
		From: moveElementToElement.From,
		To:   elementID,
	}
	err := s.devicesManager.SendAction(log, session, a)
	if err != nil {
		s.renderError(c, err)
		return
	}
	moveElementToElement = nil
	elementID = ""

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     true,
	})
}

func (s *SeleniumService) TouchPosition(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	if device, ok := session.GetDevice().(*androiddevice.Device); ok {
		device.Tap(int64(req.X), int64(req.Y))
	} else {
		downAction := &action.TouchDownPosition{int64(req.X), int64(req.Y), false}
		if err := s.devicesManager.SendAction(log, session, downAction); err != nil {
			s.renderError(c, err)
			return
		}

		upAction := &action.TouchUpPosition{int64(req.X), int64(req.Y), false}
		if err := s.devicesManager.SendAction(log, session, upAction); err != nil {
			s.renderError(c, err)
			return
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) TouchDown(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	downAction := &action.TouchDownPosition{int64(req.X), int64(req.Y), false}
	if err := s.devicesManager.SendAction(log, session, downAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) TouchMove(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	downAction := &action.TouchMovePosition{int64(req.X), int64(req.Y), false}
	if err := s.devicesManager.SendAction(log, session, downAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) TouchUp(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	upAction := &action.TouchUpPosition{int64(req.X), int64(req.Y), false}
	if err := s.devicesManager.SendAction(log, session, upAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) LongClickElement(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")
	type Request struct {
		ElementID string `json:"element"`
	}
	req := &Request{}
	c.Bind(req)
	until := session.GetActionTimeout()

	var result string
	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")
			c.JSON(http.StatusOK, &Response{
				SessionID: session.GetSessionID(),
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		a := &action.LongTouchElement{
			ElementID: req.ElementID,
		}

		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		if a.Success {
			result = "success"
			break
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}


func (s *SeleniumService) printBody(c *gin.Context) {
	debugmap := map[string]interface{}{}
	c.Bind(&debugmap)
	s.logger.Debugf("body: %v\n", debugmap)
}

func (s *SeleniumService) renderError(c *gin.Context, err error) {
	s.logger.Errorf("%v", err)
	c.JSON(http.StatusNotFound, &Response{
		SessionID: "",
		State:     "Failed",
		Message:   err.Error(),
		HCode:     time.Now().UTC().Unix(),
	})
}

func (s *SeleniumService) GetGraph(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	a := &action.GetSceenGraph{}
	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		s.renderError(c, err)
		return
	}

	c.String(http.StatusOK, a.Content())
}

type screen struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Data   string `json:"data"`
}

func (s *SeleniumService) GetScreen(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	var data *screen
	var payload []byte
	success := false
	if viper.GetBool("use_os_screenshot") {
		st := time.Now()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			start := time.Now()
			a := &action.GetSceenGraph{}
			if err := s.devicesManager.SendAction(log, session, a); err != nil {
				s.renderError(c, err)
				return
			}
			payload = []byte(a.Content())
			log.Infof("GetGraph took: %d ms", time.Now().Sub(start).Milliseconds())
			wg.Done()
		}()

		go func() {
			start := time.Now()
			rawData, width, height, err := session.GetDevice().GetScreenshot()
			if err != nil {
				s.logger.Errorf("Screenshot could not be created: %v", err)
				s.renderError(c, err)
				return
			}
			data = &screen{
				Width:  width,
				Height: height,
				Data:   base64.StdEncoding.EncodeToString(rawData),
			}
			success = true
			log.Infof("GetScreenshot took: %d ms", time.Now().Sub(start).Milliseconds())
			wg.Done()
		}()
		wg.Wait()
		log.Infof("Complete Action took: %d ms", time.Now().Sub(st).Milliseconds())

	} else {
		st := time.Now()
		a := &action.GetScreenshot{}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
		if a.Success {
			a.ScreenshotData()
			data = &screen{
				Width:  a.Width(),
				Height: a.Height(),
				Data:   base64.StdEncoding.EncodeToString(a.ScreenshotData()),
			}
			payload = a.SceengraphXML()
			success = true
		}
		log.Infof("Complete Action took: %d ms", time.Now().Sub(st).Milliseconds())
	}

	if success {
		c.JSON(http.StatusOK, &Response{
			SessionID: session.GetSessionID(),
			State:     "success",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
			Value:     data,
			Payload:   payload,
		})
	} else {
		c.JSON(http.StatusBadRequest, &Response{
			SessionID: session.GetSessionID(),
			State:     "failed",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
		})
	}
}

func (s *SeleniumService) TakeScreenshot(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	var data string
	var payload []byte

	if viper.GetBool("use_os_screenshot") {
		a := &action.GetSceenGraph{}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		rawData, _, _, err := session.GetDevice().GetScreenshot()
		if err != nil {
			s.logger.Errorf("Screenshot could not be created: %v", err)
		}
		_, err = session.GetStorage().StoreSceneGraph([]byte(a.Content()))
		if err != nil {
			s.logger.Errorf("Store SceneGraph failed: %v", err)
		}
		_, err = session.GetStorage().StoreImage(rawData)
		if err != nil {
			s.logger.Errorf("Store SceneImage failed: %v", err)
		}

		data = base64.StdEncoding.EncodeToString(rawData)
		payload = []byte(a.Content())
	} else {
		st := time.Now()
		a := &action.GetScreenshot{}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
		xmlPath, err := session.GetStorage().StoreSceneGraph(a.SceengraphXML())
		if err != nil {
			s.logger.Errorf("Store SceneGraph failed: %v", err)
		}

		pngPath, err := session.GetStorage().StoreImage(a.ScreenshotData())
		if err != nil {
			s.logger.Errorf("Store SceneImage failed: %v", err)
		}

		s.logger.Infof("Files:\n%s\n%s", xmlPath, pngPath)

		data = base64.StdEncoding.EncodeToString(a.ScreenshotData())
		payload = a.SceengraphXML()
		log.Infof("Complete Action took: %d ms", time.Now().Sub(st).Milliseconds())
	}

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     data,
		Payload:   payload,
	})
}

func (s *SeleniumService) SetTimeouts(session *Session, c *gin.Context) {
	type timeout struct {
		Type string `json:"type"`
		MS   int    `json:"ms"`
	}
	var to timeout
	if err := c.Bind(&to); err != nil {
		s.renderError(c, err)
		return
	}
	s.logger.Infof("set timeout: %v", to)
	session.SetActionTimeout(time.Millisecond * time.Duration(to.MS))
	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *SeleniumService) RestartApp(session *Session, c *gin.Context) {
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
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		c.JSON(http.StatusOK, &Response{
			SessionID: session.GetSessionID(),
			State:     "success",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
		})
	} else {
		if err := session.GetDevice().StopApp(session.AppParameter); err != nil {
			logrus.Errorf("Stop App failed: %v", err)
		}
		if err := session.GetDevice().StartApp(session.AppParameter, "", s.hostIP); err != nil {
			logrus.Errorf("Start App failed: %v", err)
		}
		session.WaitForConnection()

		c.JSON(http.StatusOK, &Response{
			SessionID: session.GetSessionID(),
			State:     "success",
			HCode:     time.Now().UTC().Unix(),
			Status:    0,
		})
	}
}

func (s *SeleniumService) NavigateBack(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	if session.GetDevice().HasFeature("back") {
		session.GetDevice().Execute("back")
	} else {
		a := &action.Custom{
			RequestData: "back",
		}
		if err := s.devicesManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}
