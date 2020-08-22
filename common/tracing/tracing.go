package tracing

import (
	"context"
	"time"
	"fmt"
	//"github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/ext"
	"github.com/openzipkin/zipkin-go"
	//zipkinmodel "github.com/openzipkin/zipkin-go/model"
	httpmiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	//zipkinlogreporter "github.com/openzipkin/zipkin-go/reporter/log"
	//"log"
	//"os"
	//zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Tracer instance
var tracer *zipkin.Tracer
var ServerMiddleware func(http.Handler) http.Handler
var TracingClient *httpmiddleware.Client

// SetTracer can be used by unit tests to provide a NoopTracer instance. Real users should always
// use the InitTracing func.
//func SetTracer(initializedTracer opentracing.Tracer) {
//	tracer = initializedTracer
//}

// InitTracing connects the calling service to Zipkin and initializes the tracer.
func InitTracing(zipkinURL string, serviceName string) {
	logrus.Infof("Connecting to zipkin server at %v", zipkinURL)
	spanUrl := fmt.Sprintf("%s/api/v2/spans", zipkinURL)
	logrus.Info("Zipkin Span Url:", spanUrl)
	reporter := httpreporter.NewReporter(
		spanUrl,
		httpreporter.RequestCallback(func (r *http.Request) {
			logrus.Debugf("Reporter Method: %s, Url: %s, Length: %v", r.Method, r.URL, r.ContentLength)
		}),)
	//reporter := zipkinlogreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint(serviceName, "")
	if err != nil {
		logrus.Fatalf("unable to create local endpoint: %+v", err)
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logrus.Errorln("Error starting new zipkin tracer. Error: " + err.Error())
		panic("Error starting new zipkin tracer. Error: " + err.Error())
	}

	// create zipkin traced http client
	client, err := httpmiddleware.NewClient(nativeTracer, httpmiddleware.ClientTrace(false))
	if err != nil {
		logrus.Fatalf("unable to create client: %+v", err)
	}
	TracingClient = client

	ServerMiddleware = httpmiddleware.NewServerMiddleware(
		nativeTracer,
		httpmiddleware.TagResponseSize(true),
		httpmiddleware.SpanName(serviceName),
	)

	//opentracing.SetGlobalTracer(zipkinot.Wrap(nativeTracer))
	//tracer = zipkinot.Wrap(nativeTracer)
	tracer = nativeTracer
	logrus.Infof("Successfully started zipkin tracer for service '%v'", serviceName)
}

//// StartHTTPTrace loads tracing information from an INCOMING HTTP request.
//func StartHTTPTrace(r *http.Request, opName string) zipkin.Span {
//	carrier := opentracing.HTTPHeadersCarrier(r.Header)
//	clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
//	if err == nil {
//		return tracer.StartSpan(
//			opName, ext.RPCServerOption(clientContext))
//	} else {
//		return tracer.StartSpan(opName)
//	}
//}
//
//// MapToCarrier converts a generic map to opentracing http headers carrier
//func MapToCarrier(headers map[string]interface{}) opentracing.HTTPHeadersCarrier {
//	carrier := make(opentracing.HTTPHeadersCarrier)
//	for k, v := range headers {
//		// delivery.Headers
//		carrier.Set(k, v.(string))
//	}
//	return carrier
//}
//
//// CarrierToMap converts a TextMapCarrier to the amqp headers format
//func CarrierToMap(values map[string]string) map[string]interface{} {
//	headers := make(map[string]interface{})
//	for k, v := range values {
//		headers[k] = v
//	}
//	return headers
//}
//
//// StartTraceFromCarrier extracts tracing info from a generic map and starts a new span.
//func StartTraceFromCarrier(carrier map[string]interface{}, spanName string) opentracing.Span {
//
//	clientContext, err := tracer.Extract(opentracing.HTTPHeaders, MapToCarrier(carrier))
//	var span opentracing.Span
//	if err == nil {
//		span = tracer.StartSpan(
//			spanName, ext.RPCServerOption(clientContext))
//	} else {
//		span = tracer.StartSpan(spanName)
//	}
//	return span
//}
//
//// AddTracingToReq adds tracing information to an OUTGOING HTTP request
//func AddTracingToReq(req *http.Request, span opentracing.Span) {
//	carrier := opentracing.HTTPHeadersCarrier(req.Header)
//	err := tracer.Inject(
//		span.Context(),
//		opentracing.HTTPHeaders,
//		carrier)
//	if err != nil {
//		panic("Unable to inject tracing context: " + err.Error())
//	}
//}
//
//// AddTracingToReqFromContext adds tracing information to an OUTGOING HTTP request
//func AddTracingToReqFromContext(ctx context.Context, req *http.Request) {
//	if ctx.Value("opentracing-span") == nil {
//		return
//	}
//	carrier := opentracing.HTTPHeadersCarrier(req.Header)
//	err := tracer.Inject(
//		ctx.Value("opentracing-span").(opentracing.Span).Context(),
//		opentracing.HTTPHeaders,
//		carrier)
//	if err != nil {
//		panic("Unable to inject tracing context: " + err.Error())
//	}
//}
//
//func AddTracingToTextMapCarrier(span opentracing.Span, val opentracing.TextMapCarrier) error {
//	return tracer.Inject(span.Context(), opentracing.TextMap, val)
//}
//
//// StartSpanFromContext starts a span.
//func StartSpanFromContext(ctx context.Context, opName string) opentracing.Span {
//	span := ctx.Value("opentracing-span").(opentracing.Span)
//	child := tracer.StartSpan(opName, ext.RPCServerOption(span.Context()))
//	return child
//}
//
// StartChildSpanFromContext starts a child span from span within the supplied context, if available.
func StartChildSpanFromContext(ctx context.Context, opName string) zipkin.Span {
	logrus.Debugf("StartChildSpanFromContext Span: %s", opName)

	//if ctx.Value("opentracing-span") == nil {
	//	return tracer.StartSpan(opName, ext.RPCServerOption(nil))
	//}
	//parent := ctx.Value("opentracing-span").(opentracing.Span)
	//child := tracer.StartSpan(opName, opentracing.ChildOf(parent.Context()))

	child, _ := tracer.StartSpanFromContext(ctx, opName)
	return child
}

// StartSpanFromContextWithLogEvent starts span from context with logevent
func StartSpanFromContextWithLogEvent(ctx context.Context, opName string, logStatement string) zipkin.Span {
	logrus.Debugf("StartChildSpanFromContextWithLogEvent Span: %s, Log: %s", opName, logStatement)

	//span := ctx.Value("opentracing-span").(opentracing.Span)
	//child := tracer.StartSpan(opName, ext.RPCServerOption(span.Context()))
	//child.LogEvent(logStatement)
//	child, _ := tracer.StartSpanFromContext(ctx, opName,)

	child := StartChildSpanFromContext(ctx, opName)
	child.Annotate(time.Now(), logStatement)
	return child
}

// CloseSpan logs event finishes span.
func CloseSpan(span zipkin.Span, event string) {
	logrus.Debugf("CloseSpan with event: %s", event)

	//sc := span.Context()
	//logrus.Debugf("Closing TraceID: %s, ID: %s, Parent ID: %s, event: %s",
		//sc.TraceID, sc.ID, sc.ParentID, event)
	//span.Annotate(time.Now(), event)

	span.Annotate(time.Now(), event)
	span.Finish()
}

// LogEventToOngoingSpan extracts span from context and adds LogEvent
func LogEventToOngoingSpan(ctx context.Context, logMessage string) {
	logrus.Debugf("LogEventToOngoingSpan Log: %s", logMessage)

	//if ctx.Value("opentracing-span") != nil {
	//	ctx.Value("opentracing-span").(opentracing.Span).LogEvent(logMessage)
	//}
	span := zipkin.SpanFromContext(ctx)
	if span != nil {
		//sc := span.Context()
		//logrus.Debugf("LogEventToOngoingSpan TraceID: %s, ID: %s, Parent ID: %s, log: %s",
		//	sc.TraceID, sc.ID, sc.ParentID, logMessage)
		span.Annotate(time.Now(), logMessage)
	} else {
		logrus.Error("Cannot log events to ongoing span, no span found in context")
	}
}

//// UpdateContext updates the supplied context with the supplied span.
//func UpdateContext(ctx context.Context, span opentracing.Span) context.Context {
//	return context.WithValue(ctx, "opentracing-span", span)
//}
