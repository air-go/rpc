package gorm

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	libraryOtel "github.com/air-go/rpc/library/otel"
)

const (
	callBackBeforeName = "opentelemetry:before"
	callBackAfterName  = "opentelemetry:after"
)

// before gorm before execute action do something
func before(db *gorm.DB) {
	if !libraryOtel.CheckHasTraceID(db.Statement.Context) {
		return
	}
	db.Statement.Context, _ = libraryOtel.Tracer().Start(db.Statement.Context, semconv.DBSystemMySQL.Value.AsString(), trace.WithSpanKind(trace.SpanKindClient))
}

// after gorm after execute action do something
func after(db *gorm.DB) {
	if !libraryOtel.CheckHasTraceID(db.Statement.Context) {
		return
	}

	span := trace.SpanFromContext(db.Statement.Context)
	defer span.End()

	if db.Error != nil {
		span.SetStatus(codes.Error, db.Error.Error())
		span.SetAttributes(libraryOtel.AttributeRedisError.String(db.Error.Error()))
	}

	span.AddEvent("SQL", trace.WithAttributes([]attribute.KeyValue{
		semconv.DBStatementKey.String(db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)),
	}...))
}

type opentelemetryPlugin struct{}

func NewOpentelemetryPlugin() gorm.Plugin {
	return &opentelemetryPlugin{}
}

func (op *opentelemetryPlugin) Name() string {
	return "opentelemetryPlugin"
}

func (op *opentelemetryPlugin) Initialize(db *gorm.DB) (err error) {
	// create
	if err = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after); err != nil {
		return err
	}

	// query
	if err = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after); err != nil {
		return err
	}

	// delete
	if err = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after); err != nil {
		return err
	}

	// update
	if err = db.Callback().Update().Before("gorm:before_update").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after); err != nil {
		return err
	}

	// row
	if err = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after); err != nil {
		return err
	}

	// raw
	if err = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after); err != nil {
		return err
	}

	// associations
	if err = db.Callback().Raw().Before("gorm:save_before_associations").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Update().After("gorm:save_after_associations").Register(callBackAfterName, after); err != nil {
		return err
	}
	return nil
}
