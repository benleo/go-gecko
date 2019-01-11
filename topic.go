package gecko

import "strings"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 需要Topic支持
type NeedTopicFilter interface {
	// 设置当前Driver可处理的Topic列表
	setTopics(topics []string)
	// 返回当前Driver可处理的Topic列表
	GetTopicExpr() []*TopicExpr
}

type AbcTopicFilter struct {
	topics []*TopicExpr
}

func (ad *AbcTopicFilter) setTopics(topics []string) {
	if nil == ad.topics {
		ad.topics = make([]*TopicExpr, 0, len(topics))
	}
	for _, t := range topics {
		ad.topics = append(ad.topics, newTopicExpr(t))
	}
}

func (ad *AbcTopicFilter) GetTopicExpr() []*TopicExpr {
	return ad.topics
}

////

// Topic表达式，类似MQTT的Topic匹配方式
type TopicExpr struct {
	exprs []string
}

// 判断当前Topic与外部Topic是否匹配
func (t *TopicExpr) matches(topic string) bool {
	ss := strings.Split(topic, "/")
	if len(t.exprs) > len(ss) {
		return false
	}
	for i, ex := range t.exprs {
		if "#" == ex {
			return true
		} else if ex != "+" && ex != ss[i] {
			return false
		}
	}
	return true
}

func newTopicExpr(expr string) *TopicExpr {
	return &TopicExpr{
		exprs: strings.Split(expr, "/"),
	}
}
