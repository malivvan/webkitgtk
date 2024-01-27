package webkitgtk

import (
	"context"
	"fmt"
	"github.com/godbus/dbus/v5"
	"sync"
	"time"
)

const (
	notifierDbusObjectPath    = "/org/freedesktop/Notifications"
	notifierDbusInterfacePath = "org.freedesktop.Notifications"
)

type notificationAction struct {
	label    string
	callback func()
}

type Notification struct {
	notifier *dbusNotify

	id        uint32                 // The notification ID.
	replaceID uint32                 // The optional notification to replace.
	icon      []byte                 // The optional icon data.
	title     string                 // The summary text briefly describing the notification.
	message   string                 // The optional detailed body text.
	actions   []notificationAction   // The actions send a request message back to the notification client when invoked.
	onClose   []func()               // The optional callback function to be called when the notification is closed.
	hints     map[string]interface{} // hints are a way to provide extra data to a notification server.
	timeout   time.Duration          // The timeout since the display of the notification at which the notification should
}

func (a *App) Notify(title, message string) *Notification {
	return &Notification{
		notifier: a.notifier,
		message:  message,
		title:    title,
	}
}

func (n *Notification) Icon(icon []byte) *Notification {
	n.icon = icon
	return n
}

func (n *Notification) Timeout(timeout time.Duration) *Notification {
	n.timeout = timeout
	return n
}

func (n *Notification) Action(label string, callback func()) *Notification {
	n.actions = append(n.actions, notificationAction{
		label:    label,
		callback: callback,
	})
	return n
}

func (n *Notification) Closed(callback func()) *Notification {
	n.onClose = append(n.onClose, callback)
	return n
}

type dbusNotify struct {
	log  logFunc
	conn *dbus.Conn

	appName string
	appIcon string

	notifications_ sync.Mutex
	notifications  map[uint32]*Notification

	server        string // Notification Server Name
	vendor        string // Notification Server Vendor
	version       string // Notification Server Version
	specification string // Spec Version

	actionIcons    bool // Supports using icons instead of text for displaying actions.
	actions        bool // The server will provide any specified actions to the user.
	body           bool // Supports body text. Some implementations may only show the summary.
	bodyHyperlinks bool // The server supports hyperlinks in the notifications.
	bodyImages     bool // The server supports images in the notifications.
	bodyMarkup     bool // Supports markup in the body text.
	iconMulti      bool // The server will render an animation of all the frames in a given image array.
	iconStatic     bool // Supports display of exactly 1 frame of any given image array.
	persistence    bool // The server supports persistence of notifications.
	sound          bool // The server supports sounds on notifications.
}

func (n *dbusNotify) Start(conn *dbus.Conn) error {
	n.log = newLogFunc("dbus-notify")
	n.conn = conn
	n.notifications = make(map[uint32]*Notification)
	/////////////

	var d = make(chan *dbus.Call, 1)
	var o = conn.Object(notifierDbusInterfacePath, notifierDbusObjectPath)
	o.GoWithContext(context.Background(),
		"org.freedesktop.Notifications.GetServerInformation",
		0,
		d)
	err := (<-d).Store(&n.server,
		&n.vendor,
		&n.version,
		&n.specification)
	if err != nil {
		return fmt.Errorf("error getting notification server information: %w", err)
	}
	n.log("notification server information", "server", n.server, "vendor", n.vendor, "version", n.version, "specification", n.specification)

	//var d = make(chan *dbus.Call, 1)
	//var o = n.dbus.conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	var s = make([]string, 0)
	o.GoWithContext(context.Background(),
		"org.freedesktop.Notifications.GetCapabilities",
		0,
		d)
	err = (<-d).Store(&s)
	if err != nil {
		return fmt.Errorf("error getting notification server capabilities: %w", err)
	}
	for _, v := range s {
		switch v {
		case "action-icons":
			n.actionIcons = true
			break
		case "actions":
			n.actions = true
			break
		case "body":
			n.body = true
			break
		case "body-hyperlinks":
			n.bodyHyperlinks = true
			break
		case "body-images":
			n.bodyImages = true
			break
		case "body-markup":
			n.bodyMarkup = true
			break
		case "icon-multi":
			n.iconMulti = true
			break
		case "icon-static":
			n.iconStatic = true
			break
		case "persistence":
			n.persistence = true
			break
		case "sound":
			n.sound = true
			break
		}
	}

	n.log("notification server capabilities", "action-icons", n.actionIcons, "actions", n.actions, "body", n.body, "body-hyperlinks", n.bodyHyperlinks, "body-images", n.bodyImages, "body-markup", n.bodyMarkup, "icon-multi", n.iconMulti, "icon-static", n.iconStatic, "persistence", n.persistence, "sound", n.sound)
	n.log("started")
	return nil
}

// Show sends the information in the notification object to the server to be
// displayed.
func (n *Notification) Show() (uint32, error) {

	timeout := int32(-1)
	if n.timeout > 0 {
		timeout = int32(n.timeout.Milliseconds())
	}

	// We need to convert the interface type of the map to dbus.Variant as
	// people dont want to have to import the dbus package just to make use
	// of the notification hints.
	hints := make(map[string]interface{})
	for k, v := range n.hints {
		hints[k] = dbus.MakeVariant(v)
	}

	var actions []string
	if n.notifier.actions {
		for _, v := range n.actions {
			actions = append(actions, v.label)
			actions = append(actions, v.label)
		}
	}

	var appIcon string

	var d = make(chan *dbus.Call, 1)
	var o = n.notifier.conn.Object(notifierDbusInterfacePath, notifierDbusObjectPath)
	o.GoWithContext(context.Background(),
		"org.freedesktop.Notifications.Notify",
		0,
		d,
		n.notifier.appName,
		n.replaceID,
		appIcon,
		n.title,
		n.message,
		actions,
		hints,
		timeout)
	err := (<-d).Store(&n.id)
	if err != nil {
		return 0, fmt.Errorf("error showing notification: %w", err)
	}

	n.notifier.notifications_.Lock()
	n.notifier.notifications[n.id] = n
	n.notifier.notifications_.Unlock()

	return n.id, nil
}
func (n *dbusNotify) Signal(sig *dbus.Signal) {
	if sig.Path != notifierDbusObjectPath {
		return
	}
	switch sig.Name {
	case "org.freedesktop.Notifications.NotificationClosed":
		id := sig.Body[0].(uint32)

		n.notifications_.Lock()
		notification, ok := n.notifications[id]
		delete(n.notifications, id)
		n.notifications_.Unlock()

		n.log("notification closed", "id", id, "reason", sig.Body[1].(uint32))
		if !ok {
			n.log("notification not found", "id", id)
			return
		}

		for _, onClose := range notification.onClose {
			onClose()
		}

	case "org.freedesktop.Notifications.ActionInvoked":
		id := sig.Body[0].(uint32)

		n.notifications_.Lock()
		notification, ok := n.notifications[id]
		n.notifications_.Unlock()

		n.log("notification action invoked", "id", sig.Body[0].(uint32), "action", sig.Body[1].(string))
		if !ok {
			n.log("notification not found", "id", id)
			return
		}

		action := sig.Body[1].(string)
		for _, v := range notification.actions {
			if v.label == action {
				v.callback()
				return
			}
		}
	}
}

func (n *dbusNotify) Stop() {
	n.log("stopped")
}
