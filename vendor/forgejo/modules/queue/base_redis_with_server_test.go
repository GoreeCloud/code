package queue

import (
	"os"
	"testing"

	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/suite"
)

type baseRedisWithServerTestSuite struct {
	suite.Suite
}

func TestBaseRedisWithServer(t *testing.T) {
	suite.Run(t, &baseRedisWithServerTestSuite{})
}

func (suite *baseRedisWithServerTestSuite) TestNormal() {
	redisAddress := "redis://" + suite.testRedisHost() + "/0"
	queueSettings := setting.QueueSettings{
		Length:  10,
		ConnStr: redisAddress,
	}

	testQueueBasic(suite.T(), newBaseRedisSimple, toBaseConfig("baseRedis", queueSettings), false)
	testQueueBasic(suite.T(), newBaseRedisUnique, toBaseConfig("baseRedisUnique", queueSettings), true)
}

func (suite *baseRedisWithServerTestSuite) TestWithPrefix() {
	redisAddress := "redis://" + suite.testRedisHost() + "/0?prefix=forgejo:queue:"
	queueSettings := setting.QueueSettings{
		Length:  10,
		ConnStr: redisAddress,
	}

	testQueueBasic(suite.T(), newBaseRedisSimple, toBaseConfig("baseRedis", queueSettings), false)
	testQueueBasic(suite.T(), newBaseRedisUnique, toBaseConfig("baseRedisUnique", queueSettings), true)
}

func (suite *baseRedisWithServerTestSuite) testRedisHost() string {
	host := os.Getenv("TEST_REDIS_SERVER")
	if host == "" {
		suite.T().Skip("redis-server not found in Forgejo test yet")
	}
	return host
}
