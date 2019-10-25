package dashboard

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// List all articles
func Index(c *gin.Context) {
	// db := c.MustGet("db").(*mgo.Database)
	//err := db.C(models.CollectionArticle).Find(nil).Sort("-updated_on").All(&articles)
	//if err != nil {
	//	c.Error(err)
	//}
	c.HTML(http.StatusOK, "dashboard/dashboard", gin.H{
		// "markets": markets,
	})
}

