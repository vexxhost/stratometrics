package keystoneauth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	log "github.com/sirupsen/logrus"
)

func Middleware() gin.HandlerFunc {
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.WithError(err).Fatal("could not load auth options from env")
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		log.WithError(err).Fatal("could not create authenticated client")
	}

	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.WithError(err).Fatal("could not create identity client")
	}

	return func(c *gin.Context) {
		token := c.GetHeader("X-Auth-Token")
		if token == "" {
			c.AbortWithError(401, fmt.Errorf("missing X-Auth-Token header"))
			return
		}

		projectData, err := tokens.Get(identityClient, token).ExtractProject()
		if err != nil {
			c.AbortWithError(401, err)
		}

		c.Set("project_id", projectData.ID)
		c.Next()
	}
}
