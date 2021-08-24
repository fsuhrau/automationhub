package selenium

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (s *SeleniumService) GetElement(session *Session, c *gin.Context) {
	log := session.GetLogger().WithField("prefix", "action")

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
				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
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
				c.JSON(http.StatusOK, &Response{
					SessionID: session.GetSessionID(),
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
