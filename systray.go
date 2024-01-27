package webkitgtk

import (
	_ "embed"
	"fmt"
	"github.com/ebitengine/purego"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"image"
	"os"
	"sync"
)

const (
	dbusTrayItemPath = "/StatusNotifierItem"
	dbusTrayMenuPath = "/StatusNotifierMenu"
)

type dbusSystray struct {
	log         logFunc
	label       string
	icon        []byte
	onOpen      func()
	onClose     func()
	menu        *TrayMenu
	conn        *dbus.Conn
	props       *prop.Properties
	menuProps   *prop.Properties
	menuVersion uint32 // need to bump this anytime we change anything
	itemMap_    sync.Mutex
	itemMap     map[int32]*MenuItem
}

// dbusMenu is a named struct to map into generated bindings.
// It represents the layout of a menu item
type dbusItem = struct {
	V0 int32                   // items' unique id
	V1 map[string]dbus.Variant // layout properties
	V2 []dbus.Variant          // child menu(s)
}

func (s *dbusSystray) processMenu(menu *TrayMenu, parentItem *MenuItem) {

	for _, item := range menu.items {
		s.setMenuItem(item)

		item.dbusItem = &dbusItem{
			V0: item.id,
			V1: map[string]dbus.Variant{},
			V2: []dbus.Variant{},
		}

		item.dbusItem.V1["enabled"] = dbus.MakeVariant(!item.disabled)
		item.dbusItem.V1["visible"] = dbus.MakeVariant(!item.hidden)
		if item.label != "" {
			item.dbusItem.V1["label"] = dbus.MakeVariant(item.label)
		}
		if item.icon != nil {
			item.dbusItem.V1["icon-data"] = dbus.MakeVariant(item.icon)
		}

		switch item.itemType {
		case checkbox:
			item.dbusItem.V1["toggle-type"] = dbus.MakeVariant("checkmark")
			v := dbus.MakeVariant(0)
			if item.checked {
				v = dbus.MakeVariant(1)
			}
			item.dbusItem.V1["toggle-state"] = v
		case submenu:
			item.dbusItem.V1["children-display"] = dbus.MakeVariant("submenu")
			s.processMenu(item.submenu, item)
		case text:
		case radio:
			item.dbusItem.V1["toggle-type"] = dbus.MakeVariant("radio")
			v := dbus.MakeVariant(0)
			if item.checked {
				v = dbus.MakeVariant(1)
			}
			item.dbusItem.V1["toggle-state"] = v
		case separator:
			item.dbusItem.V1["type"] = dbus.MakeVariant("separator")
		}

		parentItem.dbusItem.V2 = append(parentItem.dbusItem.V2, dbus.MakeVariant(item.dbusItem))
	}
}

func (s *dbusSystray) refresh() {
	s.menuVersion++
	if err := s.menuProps.Set("com.canonical.dbusmenu", "Version",
		dbus.MakeVariant(s.menuVersion)); err != nil {
		fmt.Errorf("systray error: failed to update menu version: %v", err)
		return
	}
	if err := dbusEmit(s.conn, &dbusMenuLayoutUpdatedSignal{
		Path: dbusTrayMenuPath,
		Body: &dbusMenuLayoutUpdatedSignalBody{
			Revision: s.menuVersion,
		},
	}); err != nil {
		fmt.Errorf("systray error: failed to emit layout updated signal: %v", err)
	}
}

func (s *dbusSystray) Start(conn *dbus.Conn) error {
	err := dbusExportStatusNotifierItem(conn, dbusTrayItemPath, s)
	if err != nil {
		return fmt.Errorf("systray error: failed to export status notifier item: %v", err)
	}
	err = dbusExportMenu(conn, dbusTrayMenuPath, s)
	if err != nil {
		return fmt.Errorf("systray error: failed to export dbusmenu: %v", err)
	}

	name := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", os.Getpid()) // register id 1 for this process
	_, err = conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Errorf("systray error: failed to request name: %s\n", err)
	}
	props, err := prop.Export(conn, dbusTrayItemPath, s.createPropSpec())
	if err != nil {
		return fmt.Errorf("systray error: failed to export notifier properties to bus: %s\n", err)
	}
	menuProps, err := prop.Export(conn, dbusTrayMenuPath, s.createMenuPropSpec())
	if err != nil {
		return fmt.Errorf("systray error: failed to export menu properties to bus: %s\n", err)
	}

	s.conn = conn
	s.props = props
	s.menuProps = menuProps

	node := introspect.Node{
		Name: dbusTrayItemPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			dbusStatusNotifierItemIntrospectData,
		},
	}

	err = conn.Export(introspect.NewIntrospectable(&node), dbusTrayItemPath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		return fmt.Errorf("systray error: failed to export node introspection: %s\n", err)
	}
	menuNode := introspect.Node{
		Name: dbusTrayMenuPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			dbusMenuIntrospectData,
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&menuNode), dbusTrayMenuPath,
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		return fmt.Errorf("systray error: failed to export menu node introspection: %s\n", err)
	}

	s.register()

	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/DBus"),
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchSender("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
		dbus.WithMatchArg(0, "org.kde.StatusNotifierWatcher"),
	); err != nil {
		return fmt.Errorf("systray error: failed to register signal matching: %v\n", err)
	}

	// init menu
	rootItem := &MenuItem{
		tray: s,
		dbusItem: &dbusItem{
			V0: int32(0),
			V1: map[string]dbus.Variant{},
			V2: []dbus.Variant{},
		},
	}
	s.itemMap = map[int32]*MenuItem{0: rootItem}
	s.menu.processRadioGroups()
	s.processMenu(s.menu, rootItem)
	s.refresh()

	s.log("started")
	return nil
}

func (s *dbusSystray) Signal(sig *dbus.Signal) {
	if sig == nil {
		return // We get a nil signal when closing the window.
	}
	// sig.Body has the args, which are [name old_owner new_owner]
	if sig.Body[2] != "" {
		s.register()
	}
}

func (s *dbusSystray) Stop() {
	s.log("stopped")
}

func (s *dbusSystray) createMenuPropSpec() map[string]map[string]*prop.Prop {
	return map[string]map[string]*prop.Prop{
		"com.canonical.dbusmenu": {
			// update version each time we change something
			"Version": {
				Value:    s.menuVersion,
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"TextDirection": {
				Value:    "ltr",
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"Status": {
				Value:    "normal",
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"IconThemePath": {
				Value:    []string{},
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
		},
	}
}

func (s *dbusSystray) createPropSpec() map[string]map[string]*prop.Prop {
	props := map[string]*prop.Prop{
		"Status": {
			Value:    "Active", // Passive, Active or NeedsAttention
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"Title": {
			Value:    s.label,
			Writable: true,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"Id": {
			Value:    s.label,
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"Category": {
			Value:    "ApplicationStatus",
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"IconData": {
			Value:    "",
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},

		"IconName": {
			Value:    "",
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"IconThemePath": {
			Value:    "",
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"ItemIsMenu": {
			Value:    true,
			Writable: false,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"Menu": {
			Value:    dbus.ObjectPath(dbusTrayMenuPath),
			Writable: true,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
	}

	if iconPx, err := iconToPX(s.icon); err == nil {
		props["IconPixmap"] = &prop.Prop{
			Value:    []iconPX{iconPx},
			Writable: true,
			Emit:     prop.EmitTrue,
			Callback: nil,
		}
	}

	return map[string]map[string]*prop.Prop{
		"org.kde.StatusNotifierItem": props,
	}
}

func (s *dbusSystray) register() bool {
	obj := s.conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
	call := obj.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, dbusTrayItemPath)
	if call.Err != nil {
		s.log("systray error: failed to register", "error", call.Err)
		return false
	}

	return true
}

// AboutToShow is an implementation of the com.canonical.dbusmenu.AboutToShow method.
func (s *dbusSystray) AboutToShow(id int32) (needUpdate bool, err *dbus.Error) {
	return
}

// AboutToShowGroup is an implementation of the com.canonical.dbusmenu.AboutToShowGroup method.
func (s *dbusSystray) AboutToShowGroup(ids []int32) (updatesNeeded []int32, idErrors []int32, err *dbus.Error) {
	return
}

// GetProperty is an implementation of the com.canonical.dbusmenu.GetProperty method.
func (s *dbusSystray) GetProperty(id int32, name string) (value dbus.Variant, err *dbus.Error) {
	if item, ok := s.getMenuItem(id); ok {
		if p, ok := item.dbusItem.V1[name]; ok {
			return p, nil
		}
	}
	return
}

// Event is com.canonical.dbusmenu.Event method.
func (s *dbusSystray) Event(id int32, eventID string, data dbus.Variant, timestamp uint32) (err *dbus.Error) {
	s.log("event", "id", id, "eventID", eventID, "data", data, "timestamp", timestamp)
	if eventID == "clicked" {
		if item, ok := s.getMenuItem(id); ok {
			go item.handleClick()
		}
	}
	return
}

// EventGroup is an implementation of the com.canonical.dbusmenu.EventGroup method.
func (s *dbusSystray) EventGroup(events []struct {
	V0 int32
	V1 string
	V2 dbus.Variant
	V3 uint32
}) (idErrors []int32, err *dbus.Error) {
	for _, event := range events {
		if event.V1 == "clicked" {
			item, ok := s.getMenuItem(event.V0)
			if ok {
				item.handleClick()
			}
		}
	}
	return
}

// GetGroupProperties is an implementation of the com.canonical.dbusmenu.GetGroupProperties method.
func (s *dbusSystray) GetGroupProperties(ids []int32, propertyNames []string) (properties []struct {
	V0 int32
	V1 map[string]dbus.Variant
}, err *dbus.Error) {
	for _, id := range ids {
		if m, ok := s.getMenuItem(id); ok {
			p := struct {
				V0 int32
				V1 map[string]dbus.Variant
			}{
				V0: m.dbusItem.V0,
				V1: make(map[string]dbus.Variant, len(m.dbusItem.V1)),
			}
			for k, v := range m.dbusItem.V1 {
				p.V1[k] = v
			}
			properties = append(properties, p)
		}
	}
	return properties, nil
}

// GetLayout is an implementation of the com.canonical.dbusmenu.GetLayout method.
func (s *dbusSystray) GetLayout(parentID int32, recursionDepth int32, propertyNames []string) (revision uint32, layout dbusItem, err *dbus.Error) {
	if m, ok := s.getMenuItem(parentID); ok {
		return s.menuVersion, *m.dbusItem, nil
	}
	return
}

// Activate implements org.kde.StatusNotifierItem.Activate method.
func (s *dbusSystray) Activate(x int32, y int32) (err *dbus.Error) {
	s.log("Activate", x, y)
	return
}

// ContextMenu is org.kde.StatusNotifierItem.ContextMenu method
func (s *dbusSystray) ContextMenu(x int32, y int32) (err *dbus.Error) {
	s.log("ContextMenu", x, y)
	return
}

func (s *dbusSystray) Scroll(delta int32, orientation string) (err *dbus.Error) {
	s.log("Scroll", delta, orientation)
	return
}

// SecondaryActivate implements org.kde.StatusNotifierItem.SecondaryActivate method.
func (s *dbusSystray) SecondaryActivate(x int32, y int32) (err *dbus.Error) {
	s.log("SecondaryActivate", x, y)
	return
}

type TrayMenu struct {
	item   *MenuItem
	items  []*MenuItem
	label  string
	native ptr
}

func (m *TrayMenu) toTray(label string, icon []byte) *dbusSystray {
	if icon == nil {
		icon = defaultIcon
	}
	return &dbusSystray{
		log:         newLogFunc("dbus-systray"),
		menu:        m,
		label:       label,
		icon:        icon,
		menuVersion: 1,
	}
}

func (m *TrayMenu) Add(label string) *MenuItem {
	result := &MenuItem{
		label:    label,
		itemType: text,
	}
	m.items = append(m.items, result)
	return result
}

func (m *TrayMenu) AddSeparator() {
	result := &MenuItem{
		itemType: separator,
	}
	m.items = append(m.items, result)
}

func (m *TrayMenu) AddCheckbox(label string, checked bool) *MenuItem {
	result := &MenuItem{
		label:    label,
		checked:  checked,
		itemType: checkbox,
	}
	m.items = append(m.items, result)
	return result
}

func (m *TrayMenu) AddRadio(label string, checked bool) *MenuItem {
	result := &MenuItem{
		label:    label,
		checked:  checked,
		itemType: radio,
	}
	m.items = append(m.items, result)
	return result
}

func (m *TrayMenu) Update() {
	m.processRadioGroups()

	if m.native == 0 {
		m.native = lib.gtk.MenuNew()
	}
	m.update()
}

func (m *TrayMenu) AddSubmenu(label string) *TrayMenu {
	result := &MenuItem{
		label:    label,
		itemType: submenu,
	}
	result.submenu = &TrayMenu{
		item:  result,
		label: label,
	}
	m.items = append(m.items, result)
	return result.submenu
}

func (m *TrayMenu) Item() *MenuItem {
	return m.item
}

func (m *TrayMenu) processRadioGroups() {
	var radioGroup []*MenuItem
	for _, item := range m.items {
		if item.itemType == submenu {
			item.submenu.processRadioGroups()
			continue
		}
		if item.itemType == radio {
			radioGroup = append(radioGroup, item)
		} else {
			if len(radioGroup) > 0 {
				for _, item := range radioGroup {
					item.radioGroupMembers = radioGroup
				}
				radioGroup = nil
			}
		}
	}
	if len(radioGroup) > 0 {
		for _, item := range radioGroup {
			item.radioGroupMembers = radioGroup
		}
	}
}

func (m *TrayMenu) SetLabel(label string) {
	m.label = label
}

func (m *TrayMenu) update() {
	processMenu(m)
}

func processMenu(m *TrayMenu) {
	if m.native == 0 {
		m.native = lib.gtk.MenuNew()
	}
	var currentRadioGroup gsListPtr

	for _, item := range m.items {
		// drop the group if we have run out of radio items
		if item.itemType != radio {
			currentRadioGroup = nilRadioGroup
		}

		switch item.itemType {
		case submenu:
			m.native = lib.gtk.MenuItemNewWithLabel(item.label)
			processMenu(item.submenu)
			item.submenu.native = lib.gtk.MenuNew()
			lib.gtk.MenuItemSetSubmenu(item.native, item.submenu.native)
			lib.gtk.MenuShellAppend(m.native, item.native)
		case checkbox:
			item.native = lib.gtk.CheckMenuItemNewWithLabel(item.label)
			item.setChecked(item.checked)
			item.setDisabled(item.disabled)
			lib.gtk.MenuShellAppend(m.native, item.native)
		case text:
			item.native = lib.gtk.MenuItemNewWithLabel(item.label)
			item.setDisabled(item.disabled)
			lib.gtk.MenuShellAppend(m.native, item.native)
		case radio:
			item.native = lib.gtk.RadioMenuItemNewWithLabel(currentRadioGroup, item.label)
			item.setChecked(item.checked)
			item.setDisabled(item.disabled)
			lib.gtk.MenuShellAppend(m.native, item.native)
			currentRadioGroup = lib.gtk.RadioMenuItemGetGroup(item.native)
		case separator:
			lib.gtk.MenuShellAppend(m.native, lib.gtk.SeparatorMenuItemNew())
		}

	}
	for _, item := range m.items {
		if item.callback != nil {
			handler := func() {
				item := item
				switch item.itemType {
				case text, checkbox:
					//menuItemClicked <- item.id
					//println("text clicked")
				case radio:
					if lib.gtk.CheckMenuItemGetActive(item.native) == 1 {
						//menuItemClicked <- item.id
						//println("radio clicked")
					}
				}
			}
			item.handlerId = lib.g.SignalConnectObject(
				item.native,
				"activate",
				ptr(purego.NewCallback(handler)),
				item.native,
				0)
		}
	}
}

type menuItemType int

const (
	text menuItemType = iota
	separator
	checkbox
	radio
	submenu
)

type MenuItem struct {
	tray *dbusSystray

	id       int32
	label    string
	tooltip  string
	disabled bool
	checked  bool
	hidden   bool
	icon     []byte
	submenu  *TrayMenu
	callback func(bool)
	itemType menuItemType

	dbusItem          *dbusItem
	native            ptr
	handlerId         uint
	radioGroupMembers []*MenuItem
}

func (s *dbusSystray) setMenuItem(item *MenuItem) {
	s.itemMap_.Lock()
	defer s.itemMap_.Unlock()
	item.tray = s
	if item.id == 0 {
		item.id = int32(len(s.itemMap))
	}
	s.itemMap[item.id] = item
}

func (s *dbusSystray) getMenuItem(id int32) (*MenuItem, bool) {
	s.itemMap_.Lock()
	defer s.itemMap_.Unlock()
	item, ok := s.itemMap[id]
	return item, ok
}

func (item *MenuItem) handleClick() {
	if item.itemType == checkbox {
		item.checked = !item.checked
		item.setChecked(item.checked)
	}
	if item.itemType == radio {
		for _, member := range item.radioGroupMembers {
			member.checked = false
			if member != nil {
				member.setChecked(false)
			}
		}
		item.checked = true
		if item != nil {
			item.setChecked(true)
		}
	}
	if item.callback != nil {
		go item.callback(item.checked)
	}
}
func (item *MenuItem) SetLabel(label string) *MenuItem {
	item.label = label
	if item.dbusItem != nil {
		item.dbusItem.V1["label"] = dbus.MakeVariant(item.label)
		item.tray.refresh()
	}
	return item
}

func (item *MenuItem) SetDisabled(disabled bool) *MenuItem {
	item.disabled = disabled
	if item.dbusItem != nil {
		item.setDisabled(item.disabled)
	}
	return item
}

func (item *MenuItem) SetIcon(icon []byte) *MenuItem {
	item.icon = icon
	if item.dbusItem != nil {
		item.dbusItem.V1["icon-data"] = dbus.MakeVariant(item.icon)
		item.tray.refresh()
	}
	return item
}

func (item *MenuItem) SetChecked(checked bool) *MenuItem {
	item.checked = checked
	if item.dbusItem != nil {
		item.setChecked(item.checked)
	}
	return item
}

func (item *MenuItem) SetHidden(hidden bool) *MenuItem {
	item.hidden = hidden
	if item.dbusItem != nil {
		item.dbusItem.V1["visible"] = dbus.MakeVariant(!item.hidden)
		item.tray.refresh()
	}
	return item
}

func (item *MenuItem) Checked() bool {
	return item.checked
}

func (item *MenuItem) IsSeparator() bool {
	return item.itemType == separator
}

func (item *MenuItem) IsSubmenu() bool {
	return item.itemType == submenu
}

func (item *MenuItem) IsCheckbox() bool {
	return item.itemType == checkbox
}

func (item *MenuItem) IsRadio() bool {
	return item.itemType == radio
}

func (item *MenuItem) Hidden() bool {
	return item.hidden
}

func (item *MenuItem) OnClick(f func(bool)) *MenuItem {
	item.callback = f
	return item
}

func (item *MenuItem) Label() string {
	return item.label
}

func (item *MenuItem) Enabled() bool {
	return !item.disabled
}

func (item *MenuItem) setDisabled(disabled bool) {
	v := dbus.MakeVariant(!disabled)
	if item.dbusItem.V1["toggle-state"] != v {
		item.dbusItem.V1["enabled"] = v
		item.tray.refresh()
	}
}

func (item *MenuItem) setChecked(checked bool) {
	v := dbus.MakeVariant(0)
	if checked {
		v = dbus.MakeVariant(1)
	}
	if item.dbusItem.V1["toggle-state"] != v {
		item.dbusItem.V1["toggle-state"] = v
		item.tray.refresh()
	}
}

func ToARGB(img *image.RGBA) (int, int, []byte) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]byte, w*h*4)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			data[i] = byte(a)
			data[i+1] = byte(r)
			data[i+2] = byte(g)
			data[i+3] = byte(b)
			i += 4
		}
	}
	return w, h, data
}
