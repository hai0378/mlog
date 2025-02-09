package cache

import (
	"time"

	"github.com/goburrow/cache"
	"github.com/mlogclub/simple"
	"github.com/sirupsen/logrus"

	"github.com/mlogclub/mlog/model"
	"github.com/mlogclub/mlog/repositories"
)

var TopicCache = newTopicCache()

type topicCache struct {
	recommendCache cache.LoadingCache
}

func newTopicCache() *topicCache {
	return &topicCache{
		recommendCache: cache.NewLoadingCache(
			func(key cache.Key) (value cache.Value, e error) {
				topics, err := repositories.TopicRepository.QueryCnd(simple.GetDB(),
					simple.NewQueryCnd("status = ?", model.TopicStatusOk).Order("id desc").Size(50))
				if err != nil {
					logrus.Error(err)
				} else {
					value = topics
				}
				return
			},
			cache.WithMaximumSize(10),
			cache.WithExpireAfterAccess(10*time.Minute),
		),
	}
}

func (this *topicCache) GetRecommendTopics() []model.Topic {
	val, err := this.recommendCache.Get(recommendCacheKey)
	if err != nil {
		return nil
	}
	if val != nil {
		return val.([]model.Topic)
	}
	return nil
}

func (this *topicCache) InvalidateRecommend() {
	this.recommendCache.Invalidate(recommendCacheKey)
}
