package stream

import (
	"context"
	"github.com/jeven2016/mylibs/cache"
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"
	"go.uber.org/zap"
)

// this file forked and enhanced based on https://github.com/reugn/go-streams since I want to use go-redis/v9 and redis stream feature

//redis stream XADDArg 的values只能是如下格式
// XAddArgs accepts values in the following formats:
//   - XAddArgs.Values = []interface{}{"key1", "value1", "key2", "value2"}
//   - XAddArgs.Values = []string("key1", "value1", "key2", "value2")
//   - XAddArgs.Values = map[string]interface{}{"key1": "value1", "key2": "value2"}

// RedisStreamSource is a Redis Pub/Sub Source
type RedisStreamSource struct {
	ctx           context.Context
	redisClient   *cache.Redis
	out           chan interface{}
	streamName    string
	consumerGroup string
}

// NewRedisStreamSource returns a new RedisStreamSource instance
func NewRedisStreamSource(ctx context.Context, client *cache.Redis, streamName string,
	consumerGroup string) (*RedisStreamSource, error) {
	var err error
	if err = client.EnsureConsumeGroupCreated(ctx, streamName, consumerGroup); err != nil {
		return nil, err
	}

	source := &RedisStreamSource{
		ctx:           ctx,
		redisClient:   client,
		out:           make(chan interface{}),
		streamName:    streamName,
		consumerGroup: consumerGroup,
	}
	go func() {
		err = client.Consume(ctx, streamName, consumerGroup, source.out)
		if err != nil {
			zap.L().Warn("source stream stopped", zap.String("stream", streamName),
				zap.String("consumeGroup", consumerGroup), zap.Error(err))
		}
	}()
	return source, nil
}

// Via streams data through the given flow
func (rs *RedisStreamSource) Via(_flow streams.Flow) streams.Flow {
	flow.DoStream(rs, _flow)
	return _flow
}

// Out returns an output channel for sending data
func (rs *RedisStreamSource) Out() <-chan interface{} {
	return rs.out
}

// RedisStreamSink is a Redis Pub/Sub Sink
type RedisStreamSink struct {
	redisClient *cache.Redis
	in          chan interface{}
	streamName  string
}

// NewRedisStreamSink returns a new RedisStreamSink instance
func NewRedisStreamSink(ctx context.Context, client *cache.Redis, streamName string) *RedisStreamSink {
	sink := &RedisStreamSink{
		client,
		make(chan interface{}),
		streamName,
	}

	go sink.init(ctx)
	return sink
}

// init starts the main loop
func (rs *RedisStreamSink) init(ctx context.Context) {
	defer func() {
		//rs.redisClient.Close()
		if err := recover(); err != nil {
			zap.S().Errorf("an unexpected error occurs, %v", err)
		}
	}()

	for msg := range rs.in {
		if msg == nil {
			continue
		}
		if err := rs.redisClient.PublishMessage(ctx, msg, rs.streamName); err != nil {
			zap.S().Errorf("failed to send a message into stream %v: %v", rs.streamName, err)
		}

	}
}

// In returns an input channel for receiving data
func (rs *RedisStreamSink) In() chan<- interface{} {
	return rs.in
}
