package spanlogger

import (
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/common/user"
)

func TestSpanLoggerLogInnerLevelError(t *testing.T) {
	span, ctx := New(context.Background(), "test")
	newSpan := FromContext(ctx)
	require.Equal(t, span.Span, newSpan.Span)

	fields, err, errorKeyIndex, logAsError := span.logInner("bar", "bar_value", "level", "error", "err", errors.New("err"), "metric2", "2hgd")
	require.Nil(t, err)
	require.True(t, logAsError)
	require.Equal(t, errorKeyIndex, 4)
	require.Equal(t, len(fields), 3)
	require.Equal(t, fields[0].Key(), "bar")
	require.Equal(t, fields[1].Key(), "level")
	require.Equal(t, fields[2].Key(), "metric2")
}

func TestSpanLoggerLogInnerLevelNonError(t *testing.T) {
	span, ctx := New(context.Background(), "test")
	newSpan := FromContext(ctx)
	require.Equal(t, span.Span, newSpan.Span)

	fields, err, errorKeyIndex, logAsError := span.logInner("bar", "bar_value", "level", "info", "err", errors.New("err"), "metric2", "2hgd")
	require.Nil(t, err)
	require.False(t, logAsError)
	require.Equal(t, errorKeyIndex, -1)
	require.Equal(t, len(fields), 4)
}

func TestSpanLoggerLogInnerLevelErrorWithNoErrorEntry(t *testing.T) {
	span, ctx := New(context.Background(), "test")
	newSpan := FromContext(ctx)
	require.Equal(t, span.Span, newSpan.Span)

	fields, err, errorKeyIndex, logAsError := span.logInner("bar", "bar_value", "level", "error", "err", "err_string", "metric2", "2hgd")
	require.Nil(t, err)
	require.False(t, logAsError)
	require.Equal(t, errorKeyIndex, -1)
	require.Equal(t, len(fields), 4)
}

func TestSpanLoggerLogInnerLogsFirstError(t *testing.T) {
	span, ctx := New(context.Background(), "test")
	newSpan := FromContext(ctx)
	require.Equal(t, span.Span, newSpan.Span)

	fields, err, errorKeyIndex, logAsError := span.logInner("bar", "bar_value", "first_err", errors.New("first"), "second_err", errors.New("second"), "level", "error")
	require.Nil(t, err)
	require.True(t, logAsError)
	require.Equal(t, errorKeyIndex, 2)
	require.Equal(t, len(fields), 3)
}

func TestSpanLogger_CustomLogger(t *testing.T) {
	var logged [][]interface{}
	var logger funcLogger = func(keyvals ...interface{}) error {
		logged = append(logged, keyvals)
		return nil
	}
	span, ctx := NewWithLogger(context.Background(), logger, "test")
	_ = span.Log("msg", "original spanlogger")

	span = FromContextWithFallback(ctx, log.NewNopLogger())
	_ = span.Log("msg", "restored spanlogger")

	span = FromContextWithFallback(context.Background(), logger)
	_ = span.Log("msg", "fallback spanlogger")

	expect := [][]interface{}{
		{"method", "test", "msg", "original spanlogger"},
		{"msg", "restored spanlogger"},
		{"msg", "fallback spanlogger"},
	}
	require.Equal(t, expect, logged)
}

func TestSpanCreatedWithTenantTag(t *testing.T) {
	mockSpan := createSpan(user.InjectOrgID(context.Background(), "team-a"))

	require.Equal(t, []string{"team-a"}, mockSpan.Tag(TenantIDTagName))
}

func TestSpanCreatedWithoutTenantTag(t *testing.T) {
	mockSpan := createSpan(context.Background())

	_, exist := mockSpan.Tags()[TenantIDTagName]
	require.False(t, exist)
}

func createSpan(ctx context.Context) *mocktracer.MockSpan {
	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)

	logger, _ := New(ctx, "name")
	return logger.Span.(*mocktracer.MockSpan)
}

type funcLogger func(keyvals ...interface{}) error

func (f funcLogger) Log(keyvals ...interface{}) error {
	return f(keyvals...)
}
