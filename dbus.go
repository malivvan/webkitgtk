package webkitgtk

import (
	"context"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"sync"
)

type dbusPlugin interface {
	Start(*dbus.Conn) error
	Signal(*dbus.Signal)
	Stop()
}

type dbusSession struct {
	wg      sync.WaitGroup
	log     logFunc
	quit    chan struct{}
	plugins []dbusPlugin
}

func newDBusSession(plugins []dbusPlugin) (*dbusSession, error) {
	s := &dbusSession{plugins: plugins}

	s.log = newLogFunc("dbus")

	s.log("starting dbus session routine")
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("dbusSystray error: failed to connect to DBus: %v\n", err)
	}

	for _, plugin := range s.plugins {
		if err := plugin.Start(conn); err != nil {
			return nil, fmt.Errorf("dbusSystray error: failed to start plugin: %v\n", err)
		}
	}

	s.wg.Add(1)
	s.log("dbus session routine started")
	go func() {
		defer func() {
			s.log("dbus session routine stopped")
			s.wg.Done()
		}()

		sc := make(chan *dbus.Signal, 10)
		conn.Signal(sc)
		s.quit = make(chan struct{}, 1)
		for {
			select {
			case sig := <-sc:
				s.log("dbus signal received", "signal", sig)
				if sig == nil {
					return // We get a nil signal when closing the window.
				}
				for _, plugin := range s.plugins {
					plugin.Signal(sig)
				}
			case <-s.quit:
				s.log("stopping dbus session routine")
				for _, plugin := range s.plugins {
					plugin.Stop()
				}
				_ = conn.Close()
				close(s.quit)
				s.quit = nil
				return
			}
		}
	}()

	return s, nil
}

func (s *dbusSession) close() {
	if s.quit != nil {
		s.quit <- struct{}{}
	}
	s.wg.Wait()
}

// dbusSignal is a common interface for all signals.
type dbusSignal interface {
	Name() string
	Interface() string
	Sender() string

	path() dbus.ObjectPath
	values() []interface{}
}

// dbusEmit sends the given signal to the bus.
func dbusEmit(conn *dbus.Conn, s dbusSignal) error {
	return conn.Emit(s.path(), s.Interface()+"."+s.Name(), s.values()...)
}

// dbusAddMatchSignal registers a match rule for the given signal,
// opts are appended to the automatically generated signal's rules.
func dbusAddMatchSignal(conn *dbus.Conn, s dbusSignal, opts ...dbus.MatchOption) error {
	return conn.AddMatchSignal(append([]dbus.MatchOption{
		dbus.WithMatchInterface(s.Interface()),
		dbus.WithMatchMember(s.Name()),
	}, opts...)...)
}

// dbusRemoveMatchSignal unregisters the previously registered subscription.
func dbusRemoveMatchSignal(conn *dbus.Conn, s dbusSignal, opts ...dbus.MatchOption) error {
	return conn.RemoveMatchSignal(append([]dbus.MatchOption{
		dbus.WithMatchInterface(s.Interface()),
		dbus.WithMatchMember(s.Name()),
	}, opts...)...)
}

// dbusErrUnknownSignal is returned by dbusLookupStatusNotifierItemSignal when a signal cannot be resolved.
var dbusErrUnknownSignal = errors.New("unknown signal")

// ###############################################
var dbusMenuIntrospectData = introspect.Interface{
	Name: "com.canonical.dbusmenu",
	Methods: []introspect.Method{{Name: "GetLayout", Args: []introspect.Arg{
		{Name: "parentId", Type: "i", Direction: "in"},
		{Name: "recursionDepth", Type: "i", Direction: "in"},
		{Name: "propertyNames", Type: "as", Direction: "in"},
		{Name: "revision", Type: "u", Direction: "out"},
		{Name: "layout", Type: "(ia{sv}av)", Direction: "out"},
	}},
		{Name: "GetGroupProperties", Args: []introspect.Arg{
			{Name: "ids", Type: "ai", Direction: "in"},
			{Name: "propertyNames", Type: "as", Direction: "in"},
			{Name: "properties", Type: "a(ia{sv})", Direction: "out"},
		}},
		{Name: "GetProperty", Args: []introspect.Arg{
			{Name: "id", Type: "i", Direction: "in"},
			{Name: "name", Type: "s", Direction: "in"},
			{Name: "value", Type: "v", Direction: "out"},
		}},
		{Name: "Event", Args: []introspect.Arg{
			{Name: "id", Type: "i", Direction: "in"},
			{Name: "eventId", Type: "s", Direction: "in"},
			{Name: "data", Type: "v", Direction: "in"},
			{Name: "timestamp", Type: "u", Direction: "in"},
		}},
		{Name: "EventGroup", Args: []introspect.Arg{
			{Name: "events", Type: "a(isvu)", Direction: "in"},
			{Name: "idErrors", Type: "ai", Direction: "out"},
		}},
		{Name: "AboutToShow", Args: []introspect.Arg{
			{Name: "id", Type: "i", Direction: "in"},
			{Name: "needUpdate", Type: "b", Direction: "out"},
		}},
		{Name: "AboutToShowGroup", Args: []introspect.Arg{
			{Name: "ids", Type: "ai", Direction: "in"},
			{Name: "updatesNeeded", Type: "ai", Direction: "out"},
			{Name: "idErrors", Type: "ai", Direction: "out"},
		}},
	},
	Signals: []introspect.Signal{{Name: "ItemsPropertiesUpdated", Args: []introspect.Arg{
		{Name: "updatedProps", Type: "a(ia{sv})", Direction: "out"},
		{Name: "removedProps", Type: "a(ias)", Direction: "out"},
	}},
		{Name: "LayoutUpdated", Args: []introspect.Arg{
			{Name: "revision", Type: "u", Direction: "out"},
			{Name: "parent", Type: "i", Direction: "out"},
		}},
		{Name: "ItemActivationRequested", Args: []introspect.Arg{
			{Name: "id", Type: "i", Direction: "out"},
			{Name: "timestamp", Type: "u", Direction: "out"},
		}},
	},
	Properties: []introspect.Property{{Name: "Version", Type: "u", Access: "read"},
		{Name: "TextDirection", Type: "s", Access: "read"},
		{Name: "Status", Type: "s", Access: "read"},
		{Name: "IconThemePath", Type: "as", Access: "read"},
	},
	Annotations: []introspect.Annotation{},
}

func dbusLookupMenuSignal(signal *dbus.Signal) (dbusSignal, error) {
	switch signal.Name {
	case dbusMenuInterface + "." + "ItemsPropertiesUpdated":
		v0, ok := signal.Body[0].([]struct {
			V0 int32
			V1 map[string]dbus.Variant
		})
		if !ok {
			return nil, fmt.Errorf("prop .UpdatedProps is %T, not []struct {V0 int32;V1 map[string]dbus.Variant}", signal.Body[0])
		}
		v1, ok := signal.Body[1].([]struct {
			V0 int32
			V1 []string
		})
		if !ok {
			return nil, fmt.Errorf("prop .RemovedProps is %T, not []struct {V0 int32;V1 []string}", signal.Body[1])
		}
		return &dbusMenuItemsPropertiesUpdatedSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body: &dbusMenuItemsPropertiesUpdatedSignalBody{
				UpdatedProps: v0,
				RemovedProps: v1,
			},
		}, nil
	case dbusMenuInterface + "." + "LayoutUpdated":
		v0, ok := signal.Body[0].(uint32)
		if !ok {
			return nil, fmt.Errorf("prop .Revision is %T, not uint32", signal.Body[0])
		}
		v1, ok := signal.Body[1].(int32)
		if !ok {
			return nil, fmt.Errorf("prop .Parent is %T, not int32", signal.Body[1])
		}
		return &dbusMenuLayoutUpdatedSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body: &dbusMenuLayoutUpdatedSignalBody{
				Revision: v0,
				Parent:   v1,
			},
		}, nil
	case dbusMenuInterface + "." + "ItemActivationRequested":
		v0, ok := signal.Body[0].(int32)
		if !ok {
			return nil, fmt.Errorf("prop .Id is %T, not int32", signal.Body[0])
		}
		v1, ok := signal.Body[1].(uint32)
		if !ok {
			return nil, fmt.Errorf("prop .Timestamp is %T, not uint32", signal.Body[1])
		}
		return &dbusMenuItemActivationRequestedSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body: &dbusMenuItemActivationRequestedSignalBody{
				Id:        v0,
				Timestamp: v1,
			},
		}, nil
	default:
		return nil, dbusErrUnknownSignal
	}
}

const dbusMenuInterface = "com.canonical.dbusmenu"

type dbusMenuer interface {
	// GetLayout is com.canonical.dbusmenu.GetLayout method.
	GetLayout(parentId int32, recursionDepth int32, propertyNames []string) (revision uint32, layout struct {
		V0 int32
		V1 map[string]dbus.Variant
		V2 []dbus.Variant
	}, err *dbus.Error)
	// GetGroupProperties is com.canonical.dbusmenu.GetGroupProperties method.
	GetGroupProperties(ids []int32, propertyNames []string) (properties []struct {
		V0 int32
		V1 map[string]dbus.Variant
	}, err *dbus.Error)
	// GetProperty is com.canonical.dbusmenu.GetProperty method.
	GetProperty(id int32, name string) (value dbus.Variant, err *dbus.Error)
	// Event is com.canonical.dbusmenu.Event method.
	Event(id int32, eventId string, data dbus.Variant, timestamp uint32) (err *dbus.Error)
	// EventGroup is com.canonical.dbusmenu.EventGroup method.
	EventGroup(events []struct {
		V0 int32
		V1 string
		V2 dbus.Variant
		V3 uint32
	}) (idErrors []int32, err *dbus.Error)
	// AboutToShow is com.canonical.dbusmenu.AboutToShow method.
	AboutToShow(id int32) (needUpdate bool, err *dbus.Error)
	// AboutToShowGroup is com.canonical.dbusmenu.AboutToShowGroup method.
	AboutToShowGroup(ids []int32) (updatesNeeded []int32, idErrors []int32, err *dbus.Error)
}

func dbusExportMenu(conn *dbus.Conn, path dbus.ObjectPath, v dbusMenuer) error {
	return conn.ExportSubtreeMethodTable(map[string]interface{}{
		"GetLayout":          v.GetLayout,
		"GetGroupProperties": v.GetGroupProperties,
		"GetProperty":        v.GetProperty,
		"Event":              v.Event,
		"EventGroup":         v.EventGroup,
		"AboutToShow":        v.AboutToShow,
		"AboutToShowGroup":   v.AboutToShowGroup,
	}, path, dbusMenuInterface)
}

func dbusUnexportMenu(conn *dbus.Conn, path dbus.ObjectPath) error {
	return conn.Export(nil, path, dbusMenuInterface)
}

type dbusUnimplementedMenu struct{}

func (*dbusUnimplementedMenu) iface() string {
	return dbusMenuInterface
}

func (*dbusUnimplementedMenu) GetLayout(parentId int32, recursionDepth int32, propertyNames []string) (revision uint32, layout struct {
	V0 int32
	V1 map[string]dbus.Variant
	V2 []dbus.Variant
}, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) GetGroupProperties(ids []int32, propertyNames []string) (properties []struct {
	V0 int32
	V1 map[string]dbus.Variant
}, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) GetProperty(id int32, name string) (value dbus.Variant, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) Event(id int32, eventId string, data dbus.Variant, timestamp uint32) (err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) EventGroup(events []struct {
	V0 int32
	V1 string
	V2 dbus.Variant
	V3 uint32
}) (idErrors []int32, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) AboutToShow(id int32) (needUpdate bool, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedMenu) AboutToShowGroup(ids []int32) (updatesNeeded []int32, idErrors []int32, err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func dbusNewMenu(object dbus.BusObject) *dbusMenu {
	return &dbusMenu{object}
}

type dbusMenu struct {
	object dbus.BusObject
}

func (o *dbusMenu) GetLayout(ctx context.Context, parentId int32, recursionDepth int32, propertyNames []string) (revision uint32, layout struct {
	V0 int32
	V1 map[string]dbus.Variant
	V2 []dbus.Variant
}, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".GetLayout", 0, parentId, recursionDepth, propertyNames).Store(&revision, &layout)
	return
}

func (o *dbusMenu) GetGroupProperties(ctx context.Context, ids []int32, propertyNames []string) (properties []struct {
	V0 int32
	V1 map[string]dbus.Variant
}, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".GetGroupProperties", 0, ids, propertyNames).Store(&properties)
	return
}

func (o *dbusMenu) GetProperty(ctx context.Context, id int32, name string) (value dbus.Variant, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".GetProperty", 0, id, name).Store(&value)
	return
}

func (o *dbusMenu) Event(ctx context.Context, id int32, eventId string, data dbus.Variant, timestamp uint32) (err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".Event", 0, id, eventId, data, timestamp).Store()
	return
}

func (o *dbusMenu) EventGroup(ctx context.Context, events []struct {
	V0 int32
	V1 string
	V2 dbus.Variant
	V3 uint32
}) (idErrors []int32, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".EventGroup", 0, events).Store(&idErrors)
	return
}

func (o *dbusMenu) AboutToShow(ctx context.Context, id int32) (needUpdate bool, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".AboutToShow", 0, id).Store(&needUpdate)
	return
}

func (o *dbusMenu) AboutToShowGroup(ctx context.Context, ids []int32) (updatesNeeded []int32, idErrors []int32, err error) {
	err = o.object.CallWithContext(ctx, dbusMenuInterface+".AboutToShowGroup", 0, ids).Store(&updatesNeeded, &idErrors)
	return
}

func (o *dbusMenu) GetVersion(ctx context.Context) (version uint32, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusMenuInterface, "Version").Store(&version)
	return
}

func (o *dbusMenu) GetTextDirection(ctx context.Context) (textDirection string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusMenuInterface, "TextDirection").Store(&textDirection)
	return
}

func (o *dbusMenu) GetStatus(ctx context.Context) (status string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusMenuInterface, "Status").Store(&status)
	return
}

func (o *dbusMenu) GetIconThemePath(ctx context.Context) (iconThemePath []string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusMenuInterface, "IconThemePath").Store(&iconThemePath)
	return
}

type dbusMenuItemsPropertiesUpdatedSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusMenuItemsPropertiesUpdatedSignalBody
}

func (s *dbusMenuItemsPropertiesUpdatedSignal) Name() string {
	return "ItemsPropertiesUpdated"
}

func (s *dbusMenuItemsPropertiesUpdatedSignal) Interface() string {
	return dbusMenuInterface
}

func (s *dbusMenuItemsPropertiesUpdatedSignal) Sender() string {
	return s.sender
}

func (s *dbusMenuItemsPropertiesUpdatedSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusMenuItemsPropertiesUpdatedSignal) values() []interface{} {
	return []interface{}{s.Body.UpdatedProps, s.Body.RemovedProps}
}

type dbusMenuItemsPropertiesUpdatedSignalBody struct {
	UpdatedProps []struct {
		V0 int32
		V1 map[string]dbus.Variant
	}
	RemovedProps []struct {
		V0 int32
		V1 []string
	}
}

type dbusMenuLayoutUpdatedSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusMenuLayoutUpdatedSignalBody
}

func (s *dbusMenuLayoutUpdatedSignal) Name() string {
	return "LayoutUpdated"
}

func (s *dbusMenuLayoutUpdatedSignal) Interface() string {
	return dbusMenuInterface
}

func (s *dbusMenuLayoutUpdatedSignal) Sender() string {
	return s.sender
}

func (s *dbusMenuLayoutUpdatedSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusMenuLayoutUpdatedSignal) values() []interface{} {
	return []interface{}{s.Body.Revision, s.Body.Parent}
}

type dbusMenuLayoutUpdatedSignalBody struct {
	Revision uint32
	Parent   int32
}

type dbusMenuItemActivationRequestedSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusMenuItemActivationRequestedSignalBody
}

func (s *dbusMenuItemActivationRequestedSignal) Name() string {
	return "ItemActivationRequested"
}

func (s *dbusMenuItemActivationRequestedSignal) Interface() string {
	return dbusMenuInterface
}

func (s *dbusMenuItemActivationRequestedSignal) Sender() string {
	return s.sender
}

func (s *dbusMenuItemActivationRequestedSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusMenuItemActivationRequestedSignal) values() []interface{} {
	return []interface{}{s.Body.Id, s.Body.Timestamp}
}

type dbusMenuItemActivationRequestedSignalBody struct {
	Id        int32
	Timestamp uint32
}

// #############################################################
var dbusStatusNotifierItemIntrospectData = introspect.Interface{
	Name: "org.kde.StatusNotifierItem",
	Methods: []introspect.Method{{Name: "ContextMenu", Args: []introspect.Arg{
		{Name: "x", Type: "i", Direction: "in"},
		{Name: "y", Type: "i", Direction: "in"},
	}},
		{Name: "Activate", Args: []introspect.Arg{
			{Name: "x", Type: "i", Direction: "in"},
			{Name: "y", Type: "i", Direction: "in"},
		}},
		{Name: "SecondaryActivate", Args: []introspect.Arg{
			{Name: "x", Type: "i", Direction: "in"},
			{Name: "y", Type: "i", Direction: "in"},
		}},
		{Name: "Scroll", Args: []introspect.Arg{
			{Name: "delta", Type: "i", Direction: "in"},
			{Name: "orientation", Type: "s", Direction: "in"},
		}},
	},
	Signals: []introspect.Signal{{Name: "NewTitle"},
		{Name: "NewIcon"},
		{Name: "NewAttentionIcon"},
		{Name: "NewOverlayIcon"},
		{Name: "NewStatus", Args: []introspect.Arg{
			{Name: "status", Type: "s", Direction: ""},
		}},
		{Name: "NewIconThemePath", Args: []introspect.Arg{
			{Name: "icon_theme_path", Type: "s", Direction: "out"},
		}},
		{Name: "NewMenu"},
	},
	Properties: []introspect.Property{{Name: "Category", Type: "s", Access: "read"},
		{Name: "Id", Type: "s", Access: "read"},
		{Name: "Title", Type: "s", Access: "read"},
		{Name: "Status", Type: "s", Access: "read"},
		{Name: "WindowId", Type: "i", Access: "read"},
		{Name: "IconThemePath", Type: "s", Access: "read"},
		{Name: "Menu", Type: "o", Access: "read"},
		{Name: "ItemIsMenu", Type: "b", Access: "read"},
		{Name: "IconName", Type: "s", Access: "read"},
		{Name: "IconPixmap", Type: "a(iiay)", Access: "read", Annotations: []introspect.Annotation{
			{Name: "org.qtproject.QtDBus.QtTypeName", Value: "KDbusImageVector"},
		}},
		{Name: "OverlayIconName", Type: "s", Access: "read"},
		{Name: "OverlayIconPixmap", Type: "a(iiay)", Access: "read", Annotations: []introspect.Annotation{
			{Name: "org.qtproject.QtDBus.QtTypeName", Value: "KDbusImageVector"},
		}},
		{Name: "AttentionIconName", Type: "s", Access: "read"},
		{Name: "AttentionIconPixmap", Type: "a(iiay)", Access: "read", Annotations: []introspect.Annotation{
			{Name: "org.qtproject.QtDBus.QtTypeName", Value: "KDbusImageVector"},
		}},
		{Name: "AttentionMovieName", Type: "s", Access: "read"},
		{Name: "ToolTip", Type: "(sa(iiay)ss)", Access: "read", Annotations: []introspect.Annotation{
			{Name: "org.qtproject.QtDBus.QtTypeName", Value: "KDbusToolTipStruct"},
		}},
	},
	Annotations: []introspect.Annotation{},
}

func dbusLookupStatusNotifierItemSignal(signal *dbus.Signal) (dbusSignal, error) {
	switch signal.Name {
	case dbusStatusNotifierItemInterface + "." + "NewTitle":
		return &dbusStatusNotifierItemNewTitleSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body:   &dbusStatusNotifierItemNewTitleSignalBody{},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewIcon":
		return &dbusStatusNotifierItemNewIconSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body:   &dbusStatusNotifierItemNewIconSignalBody{},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewAttentionIcon":
		return &dbusStatusNotifierItemNewAttentionIconSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body:   &dbusStatusNotifierItemNewAttentionIconSignalBody{},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewOverlayIcon":
		return &dbusStatusNotifierItemNewOverlayIconSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body:   &dbusStatusNotifierItemNewOverlayIconSignalBody{},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewStatus":
		v0, ok := signal.Body[0].(string)
		if !ok {
			return nil, fmt.Errorf("prop .Status is %T, not string", signal.Body[0])
		}
		return &dbusStatusNotifierItemNewStatusSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body: &dbusStatusNotifierItemNewStatusSignalBody{
				Status: v0,
			},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewIconThemePath":
		v0, ok := signal.Body[0].(string)
		if !ok {
			return nil, fmt.Errorf("prop .IconThemePath is %T, not string", signal.Body[0])
		}
		return &dbusStatusNotifierItemNewIconThemePathSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body: &dbusStatusNotifierItemNewIconThemePathSignalBody{
				IconThemePath: v0,
			},
		}, nil
	case dbusStatusNotifierItemInterface + "." + "NewMenu":
		return &dbusStatusNotifierItemNewMenuSignal{
			sender: signal.Sender,
			Path:   signal.Path,
			Body:   &dbusStatusNotifierItemNewMenuSignalBody{},
		}, nil
	default:
		return nil, dbusErrUnknownSignal
	}
}

const dbusStatusNotifierItemInterface = "org.kde.StatusNotifierItem"

type dbusStatusNotifierItemer interface {
	// ContextMenu is org.kde.dbusStatusNotifierItem.ContextMenu method.
	ContextMenu(x int32, y int32) (err *dbus.Error)
	// Activate is org.kde.dbusStatusNotifierItem.Activate method.
	Activate(x int32, y int32) (err *dbus.Error)
	// SecondaryActivate is org.kde.dbusStatusNotifierItem.SecondaryActivate method.
	SecondaryActivate(x int32, y int32) (err *dbus.Error)
	// Scroll is org.kde.dbusStatusNotifierItem.Scroll method.
	Scroll(delta int32, orientation string) (err *dbus.Error)
}

func dbusExportStatusNotifierItem(conn *dbus.Conn, path dbus.ObjectPath, v dbusStatusNotifierItemer) error {
	return conn.ExportSubtreeMethodTable(map[string]interface{}{
		"ContextMenu":       v.ContextMenu,
		"Activate":          v.Activate,
		"SecondaryActivate": v.SecondaryActivate,
		"Scroll":            v.Scroll,
	}, path, dbusStatusNotifierItemInterface)
}

func dbusUnexportStatusNotifierItem(conn *dbus.Conn, path dbus.ObjectPath) error {
	return conn.Export(nil, path, dbusStatusNotifierItemInterface)
}

type dbusUnimplementedStatusNotifierItem struct{}

func (*dbusUnimplementedStatusNotifierItem) iface() string {
	return dbusStatusNotifierItemInterface
}

func (*dbusUnimplementedStatusNotifierItem) ContextMenu(x int32, y int32) (err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedStatusNotifierItem) Activate(x int32, y int32) (err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedStatusNotifierItem) SecondaryActivate(x int32, y int32) (err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func (*dbusUnimplementedStatusNotifierItem) Scroll(delta int32, orientation string) (err *dbus.Error) {
	err = &dbus.ErrMsgUnknownMethod
	return
}

func dbusNewStatusNotifierItem(object dbus.BusObject) *dbusStatusNotifierItem {
	return &dbusStatusNotifierItem{object}
}

type dbusStatusNotifierItem struct {
	object dbus.BusObject
}

func (o *dbusStatusNotifierItem) ContextMenu(ctx context.Context, x int32, y int32) (err error) {
	err = o.object.CallWithContext(ctx, dbusStatusNotifierItemInterface+".ContextMenu", 0, x, y).Store()
	return
}

func (o *dbusStatusNotifierItem) Activate(ctx context.Context, x int32, y int32) (err error) {
	err = o.object.CallWithContext(ctx, dbusStatusNotifierItemInterface+".Activate", 0, x, y).Store()
	return
}

func (o *dbusStatusNotifierItem) SecondaryActivate(ctx context.Context, x int32, y int32) (err error) {
	err = o.object.CallWithContext(ctx, dbusStatusNotifierItemInterface+".SecondaryActivate", 0, x, y).Store()
	return
}

func (o *dbusStatusNotifierItem) Scroll(ctx context.Context, delta int32, orientation string) (err error) {
	err = o.object.CallWithContext(ctx, dbusStatusNotifierItemInterface+".Scroll", 0, delta, orientation).Store()
	return
}

func (o *dbusStatusNotifierItem) GetCategory(ctx context.Context) (category string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "Category").Store(&category)
	return
}

func (o *dbusStatusNotifierItem) GetId(ctx context.Context) (id string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "Id").Store(&id)
	return
}

func (o *dbusStatusNotifierItem) GetTitle(ctx context.Context) (title string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "Title").Store(&title)
	return
}

func (o *dbusStatusNotifierItem) GetStatus(ctx context.Context) (status string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "Status").Store(&status)
	return
}

func (o *dbusStatusNotifierItem) GetWindowId(ctx context.Context) (windowId int32, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "WindowId").Store(&windowId)
	return
}

func (o *dbusStatusNotifierItem) GetIconThemePath(ctx context.Context) (iconThemePath string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "IconThemePath").Store(&iconThemePath)
	return
}

func (o *dbusStatusNotifierItem) GetMenu(ctx context.Context) (menu dbus.ObjectPath, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "Menu").Store(&menu)
	return
}

func (o *dbusStatusNotifierItem) GetItemIsMenu(ctx context.Context) (itemIsMenu bool, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "ItemIsMenu").Store(&itemIsMenu)
	return
}

func (o *dbusStatusNotifierItem) GetIconName(ctx context.Context) (iconName string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "IconName").Store(&iconName)
	return
}

func (o *dbusStatusNotifierItem) GetIconPixmap(ctx context.Context) (iconPixmap []struct {
	V0 int32
	V1 int32
	V2 []byte
}, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "IconPixmap").Store(&iconPixmap)
	return
}

func (o *dbusStatusNotifierItem) GetOverlayIconName(ctx context.Context) (overlayIconName string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "OverlayIconName").Store(&overlayIconName)
	return
}

func (o *dbusStatusNotifierItem) GetOverlayIconPixmap(ctx context.Context) (overlayIconPixmap []struct {
	V0 int32
	V1 int32
	V2 []byte
}, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "OverlayIconPixmap").Store(&overlayIconPixmap)
	return
}

func (o *dbusStatusNotifierItem) GetAttentionIconName(ctx context.Context) (attentionIconName string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "AttentionIconName").Store(&attentionIconName)
	return
}

func (o *dbusStatusNotifierItem) GetAttentionIconPixmap(ctx context.Context) (attentionIconPixmap []struct {
	V0 int32
	V1 int32
	V2 []byte
}, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "AttentionIconPixmap").Store(&attentionIconPixmap)
	return
}

func (o *dbusStatusNotifierItem) GetAttentionMovieName(ctx context.Context) (attentionMovieName string, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "AttentionMovieName").Store(&attentionMovieName)
	return
}

func (o *dbusStatusNotifierItem) GetToolTip(ctx context.Context) (toolTip struct {
	V0 string
	V1 []struct {
		V0 int32
		V1 int32
		V2 []byte
	}
	V2 string
	V3 string
}, err error) {
	err = o.object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, dbusStatusNotifierItemInterface, "ToolTip").Store(&toolTip)
	return
}

type dbusStatusNotifierItemNewTitleSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewTitleSignalBody
}

func (s *dbusStatusNotifierItemNewTitleSignal) Name() string {
	return "NewTitle"
}

func (s *dbusStatusNotifierItemNewTitleSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewTitleSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewTitleSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewTitleSignal) values() []interface{} {
	return []interface{}{}
}

type dbusStatusNotifierItemNewTitleSignalBody struct {
}

type dbusStatusNotifierItemNewIconSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewIconSignalBody
}

func (s *dbusStatusNotifierItemNewIconSignal) Name() string {
	return "NewIcon"
}

func (s *dbusStatusNotifierItemNewIconSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewIconSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewIconSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewIconSignal) values() []interface{} {
	return []interface{}{}
}

type dbusStatusNotifierItemNewIconSignalBody struct {
}

type dbusStatusNotifierItemNewAttentionIconSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewAttentionIconSignalBody
}

func (s *dbusStatusNotifierItemNewAttentionIconSignal) Name() string {
	return "NewAttentionIcon"
}

func (s *dbusStatusNotifierItemNewAttentionIconSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewAttentionIconSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewAttentionIconSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewAttentionIconSignal) values() []interface{} {
	return []interface{}{}
}

type dbusStatusNotifierItemNewAttentionIconSignalBody struct {
}

type dbusStatusNotifierItemNewOverlayIconSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewOverlayIconSignalBody
}

func (s *dbusStatusNotifierItemNewOverlayIconSignal) Name() string {
	return "NewOverlayIcon"
}

func (s *dbusStatusNotifierItemNewOverlayIconSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewOverlayIconSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewOverlayIconSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewOverlayIconSignal) values() []interface{} {
	return []interface{}{}
}

type dbusStatusNotifierItemNewOverlayIconSignalBody struct {
}

type dbusStatusNotifierItemNewStatusSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewStatusSignalBody
}

func (s *dbusStatusNotifierItemNewStatusSignal) Name() string {
	return "NewStatus"
}

func (s *dbusStatusNotifierItemNewStatusSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewStatusSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewStatusSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewStatusSignal) values() []interface{} {
	return []interface{}{s.Body.Status}
}

type dbusStatusNotifierItemNewStatusSignalBody struct {
	Status string
}

type dbusStatusNotifierItemNewIconThemePathSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewIconThemePathSignalBody
}

func (s *dbusStatusNotifierItemNewIconThemePathSignal) Name() string {
	return "NewIconThemePath"
}

func (s *dbusStatusNotifierItemNewIconThemePathSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewIconThemePathSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewIconThemePathSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewIconThemePathSignal) values() []interface{} {
	return []interface{}{s.Body.IconThemePath}
}

type dbusStatusNotifierItemNewIconThemePathSignalBody struct {
	IconThemePath string
}

type dbusStatusNotifierItemNewMenuSignal struct {
	sender string
	Path   dbus.ObjectPath
	Body   *dbusStatusNotifierItemNewMenuSignalBody
}

func (s *dbusStatusNotifierItemNewMenuSignal) Name() string {
	return "NewMenu"
}

func (s *dbusStatusNotifierItemNewMenuSignal) Interface() string {
	return dbusStatusNotifierItemInterface
}

func (s *dbusStatusNotifierItemNewMenuSignal) Sender() string {
	return s.sender
}

func (s *dbusStatusNotifierItemNewMenuSignal) path() dbus.ObjectPath {
	return s.Path
}

func (s *dbusStatusNotifierItemNewMenuSignal) values() []interface{} {
	return []interface{}{}
}

type dbusStatusNotifierItemNewMenuSignalBody struct {
}
