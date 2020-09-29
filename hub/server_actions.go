package hub

import (
	"fmt"
	"net/http"
	"strings"
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

type ActionRequest struct {
	Using string `json:"using"`
	Value string `json:"value"`
}

func (s *Service) getElements(session *Session) error {

	if viper.GetBool("use_element_cache") && session.XmlDocument != nil {
		return nil
	}

	log := session.logger.WithField("prefix", "action")

	a := &action.GetSceenGraph{}
	if err := s.deviceManager.SendAction(log, session, a); err != nil {
		return err
	}
	var err error
	session.XmlDocument, err = a.XML()
	if err != nil {
		log.Infof("Could not Convert Content to XML: %v", err)
	}
	return nil
}

func (s *Service) GetElement(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	action := ActionRequest{}
	c.Bind(&action)

	var element *xmlquery.Node

	until := session.GetActionTimeout()

	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")

			c.JSON(http.StatusOK, &ServerResponse{
				SessionID: session.SessionID,
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		if err := s.getElements(session); err != nil {
			// elements commented
			s.renderError(c, err)
			return
		}

		if session.XmlDocument == nil {
			time.Sleep(LookupRetry)
			continue
		}

		log.Debugf("locate element Using: %s Value: %s", action.Using, action.Value)

		if action.Using == "xpath" {
			element = xmlquery.FindOne(session.XmlDocument, action.Value)
			if element != nil {
				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value: map[string]interface{}{
						"ELEMENT": element.SelectAttr("ID"),
					},
				})
				return
			}
		} else {

			attribute := action.Using
			switch action.Using {
			case "name":
				attribute = "LabelText"
			case "tag name":
				attribute = "Name"
			}

			log.Debugf("%v", action)

			element := xmlquery.FindOne(session.XmlDocument, fmt.Sprintf("//*[@%s='%s']", attribute, action.Value))
			if element != nil {
				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
					State:     "success",
					HCode:     time.Now().UTC().Unix(),
					Status:    SUCCESS,
					Value: map[string]interface{}{
						"ELEMENT": element.SelectAttr("ID"),
					},
				})
				return
			}
		}
		log.Debugf("element not found retry with fresh one")
		session.XmlDocument = nil
		time.Sleep(LookupRetry)
	}
}

func (s *Service) GetElements(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	action := ActionRequest{}
	c.Bind(&action)

	until := session.GetActionTimeout()

	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")

			c.JSON(http.StatusOK, &ServerResponse{
				SessionID: session.SessionID,
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
				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
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

				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
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

func (s *Service) ElementIsDisplayed(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	elementID := c.Param("elementID")

	a := &action.IsDisplayed{
		ElementID: elementID,
	}
	err := s.deviceManager.SendAction(log, session, a)
	if err != nil {
		s.renderError(c, err)
		return
	}

	log.Debugf("Is Displayed response: %v", a.IsDisplayed)

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    SUCCESS,
		Value:     a.IsDisplayed,
	})
}

func (s *Service) ElementClick(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	elementID := c.Param("elementID")

	until := session.GetActionTimeout()

	var result string
	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")
			c.JSON(http.StatusOK, &ServerResponse{
				SessionID: session.SessionID,
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		a := &action.TouchElement{ElementID: elementID}
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		if a.Success {
			result = "success"
			break
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}

func (s *Service) ElementSetValue(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

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

	if err := s.deviceManager.SendAction(log, session, a); err != nil {
		s.renderError(c, err)
		return
	}

	result := ""
	if a.Success {
		result = "sentValue"
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}

func (s *Service) ElementGetValue(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	attr := c.Param("attribute")
	elementID := c.Param("elementID")

	a := &action.GetValue{
		ElementID: elementID,
		Attr:      attr,
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
		Value:     a.Value,
	})
}

func (s *Service) ElementGetText(session *Session, c *gin.Context) {
	c.Params = append(c.Params, gin.Param{"attribute", "text"})
	if viper.GetBool("innium_support") {
		s.ElementGetValue(session, c)
	} else {
		s.ElementGetAttribute(session, c)
	}
}

func (s *Service) ElementGetCSS(session *Session, c *gin.Context) {
	c.Params = append(c.Params, gin.Param{"attribute", "css"})
	if viper.GetBool("innium_support") {
		s.ElementGetValue(session, c)
	} else {
		s.ElementGetAttribute(session, c)
	}
}

func (s *Service) ElementGetName(session *Session, c *gin.Context) {
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

func (s *Service) ElementGetAttribute(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
	attr := c.Param("attribute")
	elementID := c.Param("elementID")

	if attr == "displayed" {
		s.ElementIsDisplayed(session, c)
		return
	}

	attr = mapSeleniumAttr(attr)

	action := ActionRequest{}
	c.Bind(&action)

	var element *xmlquery.Node

	until := session.GetActionTimeout()

	for {
		if time.Now().After(until) {
			log.Warningf("action timed out")

			c.JSON(http.StatusOK, &ServerResponse{
				SessionID: session.SessionID,
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
				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
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
				c.JSON(http.StatusOK, &ServerResponse{
					SessionID: session.SessionID,
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

func (s *Service) MoveTo(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

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
		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
		}
	} else {
		if moveElementToElement == nil {
			moveElementToElement = &MoveElementToElement{}
		}
		log.Debugf("MoveTo: %s\n", move.Element)
		elementID = move.Element
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) ButtonDown(session *Session, c *gin.Context) {

	if moveElementToElement != nil {
		moveElementToElement.From = elementID
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) ButtonUp(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	a := &action.DragAndDrop{
		From: moveElementToElement.From,
		To:   elementID,
	}
	err := s.deviceManager.SendAction(log, session, a)
	if err != nil {
		s.renderError(c, err)
	}
	moveElementToElement = nil
	elementID = ""

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     true,
	})
}

func (s *Service) TouchPosition(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	if device, ok := session.Lock.Device.(*androiddevice.Device); ok {
		device.Tap(int64(req.X), int64(req.Y))
	} else {
		downAction := &action.TouchDownPosition{int64(req.X), int64(req.Y), false}
		if err := s.deviceManager.SendAction(log, session, downAction); err != nil {
			s.renderError(c, err)
			return
		}

		upAction := &action.TouchUpPosition{int64(req.X), int64(req.Y), false}
		if err := s.deviceManager.SendAction(log, session, upAction); err != nil {
			s.renderError(c, err)
			return
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) TouchDown(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	downAction := &action.TouchDownPosition{int64(req.X), int64(req.Y), false}
	if err := s.deviceManager.SendAction(log, session, downAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) TouchMove(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	downAction := &action.TouchMovePosition{int64(req.X), int64(req.Y), false}
	if err := s.deviceManager.SendAction(log, session, downAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) TouchUp(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")

	type Request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	req := &Request{}
	c.Bind(req)

	upAction := &action.TouchUpPosition{int64(req.X), int64(req.Y), false}
	if err := s.deviceManager.SendAction(log, session, upAction); err != nil {
		s.renderError(c, err)
		return
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
	})
}

func (s *Service) LongClickElement(session *Session, c *gin.Context) {
	log := session.logger.WithField("prefix", "action")
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
			c.JSON(http.StatusOK, &ServerResponse{
				SessionID: session.SessionID,
				State:     "failed",
				HCode:     time.Now().UTC().Unix(),
				Status:    NO_SUCH_ELEMENT,
			})
			return
		}

		a := &action.LongTouchElement{
			ElementID: req.ElementID,
		}

		if err := s.deviceManager.SendAction(log, session, a); err != nil {
			s.renderError(c, err)
			return
		}

		if a.Success {
			result = "success"
			break
		}
	}

	session.XmlDocument = nil

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "success",
		HCode:     time.Now().UTC().Unix(),
		Status:    0,
		Value:     result,
	})
}
