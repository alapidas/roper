package interfaces

import (

	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TheSuite struct {

}

var _ = Suite(&TheSuite{})

// Create a temporary directory + db + persister object to use
func (suite *TheSuite) SetUpTest(c *C) {

}

func (suite *TheSuite) TestWebServer(c *C) {}
