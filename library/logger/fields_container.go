package logger

import (
	"container/list"
	"context"
	"sync"
)

func InitFieldsContainer(ctx context.Context) context.Context {
	if f := findLogFields(ctx); f != nil {
		return ctx
	}
	return context.WithValue(ctx, contextLogFields, newFieldsContainer())
}

func ForkContext(ctx context.Context) context.Context {
	if fields := mustFindLogFields(ctx); fields != nil {
		return context.WithValue(ctx, contextLogFields, fields.clone(false))
	}
	return context.WithValue(ctx, contextLogFields, newFieldsContainer())
}

func ForkContextOnlyMeta(ctx context.Context) context.Context {
	if fields := mustFindLogFields(ctx); fields != nil {
		return context.WithValue(ctx, contextLogFields, fields.clone(true))
	}
	return context.WithValue(ctx, contextLogFields, newFieldsContainer())
}

func ExtractFields(ctx context.Context) []Field {
	fields := []Field{}
	RangeFields(ctx, func(f Field) {
		fields = append(fields, f)
	})
	return fields
}

func AddField(ctx context.Context, fields ...Field) {
	mustFindLogFields(ctx).addFields(fields...)
}

func DeleteField(ctx context.Context, keys ...string) {
	mustFindLogFields(ctx).delFields(keys...)
}

func FindField(ctx context.Context, key string) Field {
	return mustFindLogFields(ctx).findField(key)
}

func RangeFields(ctx context.Context, f func(f Field)) {
	if fields := mustFindLogFields(ctx); fields != nil {
		fields.rangeFields(f)
	}
}

func mustFindLogFields(ctx context.Context) (s *fieldsContainer) {
	f := findLogFields(ctx)
	if f == nil {
		panic("fieldsContainer not init")
	}

	return f
}

func findLogFields(ctx context.Context) (s *fieldsContainer) {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(contextLogFields); v != nil {
		if fm, ok := v.(*fieldsContainer); ok {
			return fm
		}
	}
	return nil
}

type fieldsContainer struct {
	entry *list.List // Using list to ensure order.
	keys  map[string]*list.Element
	mtx   sync.RWMutex
}

func newFieldsContainer() *fieldsContainer {
	return &fieldsContainer{
		entry: list.New(),
		keys:  make(map[string]*list.Element),
	}
}

func (fc *fieldsContainer) addFields(fs ...Field) {
	fc.mtx.Lock()
	defer fc.mtx.Unlock()

	for _, f := range fs {
		if f == nil {
			continue
		}
		if _, ok := fc.keys[f.Key()]; ok {
			fc.keys[f.Key()].Value = f
		} else {
			fc.keys[f.Key()] = fc.entry.PushBack(f)
		}
	}
}

func (fc *fieldsContainer) delFields(keys ...string) {
	fc.mtx.Lock()
	defer fc.mtx.Unlock()

	for _, key := range keys {
		if f, ok := fc.keys[key]; ok {
			fc.entry.Remove(f)
			delete(fc.keys, key)
		}
	}
}

func (fc *fieldsContainer) findField(key string) Field {
	fc.mtx.RLock()
	defer fc.mtx.RUnlock()

	if f, ok := fc.keys[key]; ok {
		return f.Value.(Field)
	}
	return &field{}
}

func (fc *fieldsContainer) clone(onlyMeta bool) *fieldsContainer {
	fc.mtx.RLock()
	defer fc.mtx.RUnlock()

	copied := newFieldsContainer()
	for e := fc.entry.Front(); e != nil; e = e.Next() {
		f := e.Value.(Field)
		_, ok := metaFields[f.Key()]
		if onlyMeta && !ok {
			continue
		}
		copied.addFields(f)
	}
	return copied
}

func (fc *fieldsContainer) rangeFields(rangeFunc func(f Field)) {
	fc.mtx.RLock()
	defer fc.mtx.RUnlock()

	for f := fc.entry.Front(); f != nil; f = f.Next() {
		rangeFunc(f.Value.(Field))
	}
}
