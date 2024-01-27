package webkitgtk

import (
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	GSourceRemove int = 0

	// https://gitlab.gnome.org/GNOME/gtk/-/blob/gtk-3-24/gdk/gdkwindow.h#L121
	GdkHintMinSize = 1 << 1
	GdkHintMaxSize = 1 << 2
	// https://gitlab.gnome.org/GNOME/gtk/-/blob/gtk-3-24/gdk/gdkevents.h#L512
	GdkWindowStateIconified  = 1 << 1
	GdkWindowStateMaximized  = 1 << 2
	GdkWindowStateFullscreen = 1 << 4

	// https://gitlab.gnome.org/GNOME/gtk/-/blob/gtk-3-24/gtk/gtkmessagedialog.h#L87
	GtkButtonsNone     int = 0
	GtkButtonsOk           = 1
	GtkButtonsClose        = 2
	GtkButtonsCancel       = 3
	GtkButtonsYesNo        = 4
	GtkButtonsOkCancel     = 5

	// https://gitlab.gnome.org/GNOME/gtk/-/blob/gtk-3-24/gtk/gtkdialog.h#L36
	GtkDialogModal             = 1 << 0
	GtkDialogDestroyWithParent = 1 << 1
	GtkDialogUseHeaderBar      = 1 << 2 // actions in header bar instead of action area

	GtkOrientationVertical = 1
)

var dialogMapID = make(map[uint64]struct{})
var dialogIDLock sync.RWMutex

func getDialogID() uint64 {
	dialogIDLock.Lock()
	defer dialogIDLock.Unlock()
	dialogID := uint64(1)
	for {
		if _, ok := dialogMapID[dialogID]; !ok {
			dialogMapID[dialogID] = struct{}{}
			break
		}
		dialogID++
		if dialogID == 0 {
			panic("no more dialog IDs")
		}
	}
	return dialogID
}

func freeDialogID(id uint64) {
	dialogIDLock.Lock()
	defer dialogIDLock.Unlock()
	delete(dialogMapID, id)
}

type Dialog struct {
	app    *App
	window *Window
}

func (a *App) Dialog() *Dialog {
	return &Dialog{
		app:    a,
		window: a.CurrentWindow(),
	}
}

func (w *Window) Dialog() *Dialog {
	return &Dialog{
		app:    w.app,
		window: w,
	}
}
func (d *Dialog) Open(title string) *OpenFileDialog {
	return d.open(title)
}

func (d *Dialog) Save(title string) *SaveFileDialog {
	return d.save(title)
}

func (d *Dialog) Info(title string, message string, actions ...string) *MessageDialog {
	if len(actions) == 0 {
		actions = []string{"OK"}
	}
	return d.message(0, title, message, actions)
}

func (d *Dialog) Warn(title string, message string, actions ...string) *MessageDialog {
	if len(actions) == 0 {
		actions = []string{"OK"}
	}
	return d.message(1, title, message, actions)
}

func (d *Dialog) Ask(title string, message string, actions ...string) *MessageDialog {
	if len(actions) == 0 {
		actions = []string{"Yes", "No"}
	}
	return d.message(2, title, message, actions)
}

func (d *Dialog) Fail(title string, message string, actions ...string) *MessageDialog {
	if len(actions) == 0 {
		actions = []string{"OK"}
	}
	return d.message(3, title, message, actions)
}

func (d *Dialog) open(title string) *OpenFileDialog {
	return &OpenFileDialog{
		app:                  d.app,
		window:               d.window,
		log:                  newLogFunc("open-dialog"),
		title:                title,
		canChooseDirectories: false,
		canChooseFiles:       true,
		canCreateDirectories: true,
		resolvesAliases:      false,
	}
}

func (d *Dialog) save(title string) *SaveFileDialog {
	return &SaveFileDialog{
		app:                  d.app,
		window:               d.window,
		log:                  newLogFunc("save-dialog"),
		title:                title,
		canCreateDirectories: true,
	}
}

// enum GtkMessageType:  GtkMessageInfo = 0  GtkMessageWarning = 1  GtkMessageQuestion = 2  GtkMessageError = 3
func (d *Dialog) message(dtype int, title string, message string, actions []string) *MessageDialog {
	var buttons []*dialogMessageButton
	for _, action := range actions {
		buttons = append(buttons, &dialogMessageButton{
			Action: action,
			Label:  action,
		})
	}
	var firstButton *dialogMessageButton
	if len(buttons) > 0 {
		firstButton = buttons[0]
	}
	if firstButton != nil {
		firstButton.IsDefault = true
	}
	var lastButton *dialogMessageButton
	if len(buttons) > 0 {
		lastButton = buttons[len(buttons)-1]
	}
	if lastButton != nil {
		lastButton.IsCancel = true
	}
	return &MessageDialog{
		app:     d.app,
		window:  d.window,
		log:     newLogFunc("msg-dialog"),
		dtype:   dtype,
		title:   title,
		message: message,
		buttons: buttons,
	}
}

type dialogMessageButton struct {
	Action    string
	Label     string
	IsCancel  bool
	IsDefault bool
}

type MessageDialog struct {
	app    *App
	log    logFunc
	id     atomic.Uint64
	result chan int

	dtype   int
	title   string
	message string
	buttons []*dialogMessageButton
	icon    []byte
	window  *Window
}

func (d *MessageDialog) run() {
	if d.id.Load() == 0 {
		id := getDialogID()
		d.id.Store(id)
		defer func() {
			freeDialogID(id)
			d.id.Store(0)
			d.log("free", "id", id)
		}()
		d.log("open", "id", id, "title", d.title, "message", d.message)
		action := d.app.thread.InvokeSyncWithResult(func() any {
			return runMessageDialog(d)
		})
		if d.result != nil {
			result, ok := action.(int)
			if !ok {
				d.result <- -1
			} else {
				d.result <- result
			}
			close(d.result)
		}
	}
}

func (d *MessageDialog) SetIcon(icon []byte) *MessageDialog {
	d.icon = icon
	return d
}

func (d *MessageDialog) AddDefault(action string, label string) *MessageDialog {
	d.buttons = append(d.buttons, &dialogMessageButton{
		Action:    action,
		Label:     label,
		IsDefault: true,
	})
	return d
}
func (d *MessageDialog) AddCancel(action string, label string) *MessageDialog {
	d.buttons = append(d.buttons, &dialogMessageButton{
		Action:   action,
		Label:    label,
		IsCancel: true,
	})
	return d
}

func (d *MessageDialog) AddButton(action string, label string) *MessageDialog {
	d.buttons = append(d.buttons, &dialogMessageButton{
		Action: action,
		Label:  label,
	})
	return d
}

func (d *MessageDialog) Show(callbacks ...func(int)) chan int {
	d.result = make(chan int, 1)
	d.app.started.run(d)
	if len(callbacks) > 0 {
		go func() {
			result := <-d.result
			for _, callback := range callbacks {
				callback(result)
			}
		}()
		return nil
	}
	return d.result
}

type dialogFileFilter struct {
	DisplayName string // Filter information EG: "Image Files (*.jpg, *.png)"
	Pattern     string // semicolon separated list of extensions, EG: "*.jpg;*.png"
}

type OpenFileDialog struct {
	app    *App
	window *Window
	result chan []string

	id  atomic.Uint64
	log logFunc

	title                           string
	message                         string
	buttonText                      string
	directory                       string
	filters                         []dialogFileFilter
	canChooseDirectories            bool
	canChooseFiles                  bool
	canCreateDirectories            bool
	showHiddenFiles                 bool
	resolvesAliases                 bool
	allowsMultipleSelection         bool
	hideExtension                   bool
	canSelectHiddenExtension        bool
	treatsFilePackagesAsDirectories bool
	allowsOtherFileTypes            bool
}

func (d *OpenFileDialog) run() {
	if d.id.Load() == 0 {
		id := getDialogID()
		d.id.Store(id)
		defer func() {
			freeDialogID(id)
			d.id.Store(0)
			d.log("free", "id", id)
		}()
		d.log("open", "id", id, "title", d.title, "message", d.message, "directory", d.directory, "buttonText", d.buttonText, "filters", d.filters)
		selections, err := d.app.thread.InvokeSyncWithResultAndError(func() (any, error) {
			return runOpenFileDialog(d)
		})
		if d.result != nil {
			result, ok := selections.([]string)
			if err != nil || !ok {
				d.result <- []string{}
			} else {
				d.result <- result
			}
			close(d.result)
		}
	}
}

func (d *OpenFileDialog) CanChooseFiles(canChooseFiles bool) *OpenFileDialog {
	d.canChooseFiles = canChooseFiles
	return d
}

func (d *OpenFileDialog) CanChooseDirectories(canChooseDirectories bool) *OpenFileDialog {
	d.canChooseDirectories = canChooseDirectories
	return d
}

func (d *OpenFileDialog) CanCreateDirectories(canCreateDirectories bool) *OpenFileDialog {
	d.canCreateDirectories = canCreateDirectories
	return d
}

func (d *OpenFileDialog) AllowsOtherFileTypes(allowsOtherFileTypes bool) *OpenFileDialog {
	d.allowsOtherFileTypes = allowsOtherFileTypes
	return d
}
func (d *OpenFileDialog) AllowsMultipleSelection(allowsMultipleSelection bool) *OpenFileDialog {
	d.allowsMultipleSelection = allowsMultipleSelection
	return d
}

func (d *OpenFileDialog) ShowHiddenFiles(showHiddenFiles bool) *OpenFileDialog {
	d.showHiddenFiles = showHiddenFiles
	return d
}

func (d *OpenFileDialog) HideExtension(hideExtension bool) *OpenFileDialog {
	d.hideExtension = hideExtension
	return d
}

func (d *OpenFileDialog) TreatsFilePackagesAsDirectories(treatsFilePackagesAsDirectories bool) *OpenFileDialog {
	d.treatsFilePackagesAsDirectories = treatsFilePackagesAsDirectories
	return d
}

func (d *OpenFileDialog) ResolvesAliases(resolvesAliases bool) *OpenFileDialog {
	d.resolvesAliases = resolvesAliases
	return d
}

// AddFilter adds a filter to the dialog. The filter is a display name and a semicolon separated list of extensions.
// EG: AddFilter("Image Files", "*.jpg;*.png")
func (d *OpenFileDialog) AddFilter(displayName, pattern string) *OpenFileDialog {
	d.filters = append(d.filters, dialogFileFilter{
		DisplayName: strings.TrimSpace(displayName),
		Pattern:     strings.TrimSpace(pattern),
	})
	return d
}

func (d *OpenFileDialog) SetButtonText(text string) *OpenFileDialog {
	d.buttonText = text
	return d
}

func (d *OpenFileDialog) SetDirectory(directory string) *OpenFileDialog {
	d.directory = directory
	return d
}

func (d *OpenFileDialog) CanSelectHiddenExtension(canSelectHiddenExtension bool) *OpenFileDialog {
	d.canSelectHiddenExtension = canSelectHiddenExtension
	return d
}

func (d *OpenFileDialog) Show(callbacks ...func([]string)) chan []string {
	d.result = make(chan []string, 1)
	d.app.started.run(d)
	if len(callbacks) > 0 {
		go func() {
			result := <-d.result
			for _, callback := range callbacks {
				callback(result)
			}
		}()
		return nil
	}
	return d.result
}

type SaveFileDialog struct {
	id  atomic.Uint64
	log logFunc

	app    *App
	window *Window
	result chan string

	canCreateDirectories            bool
	showHiddenFiles                 bool
	canSelectHiddenExtension        bool
	allowOtherFileTypes             bool
	hideExtension                   bool
	treatsFilePackagesAsDirectories bool
	title                           string
	message                         string
	directory                       string
	filename                        string
	buttonText                      string
	filters                         []dialogFileFilter
}

func (d *SaveFileDialog) run() {
	if d.id.Load() == 0 {
		id := getDialogID()
		d.id.Store(id)
		defer func() {
			freeDialogID(id)
			d.id.Store(0)
			d.log("free", "id", id)
		}()
		d.log("open", "id", id, "title", d.title, "message", d.message, "directory", d.directory, "filename", d.filename, "buttonText", d.buttonText, "filters", d.filters)
		selections, err := d.app.thread.InvokeSyncWithResultAndError(func() (any, error) {
			return runSaveFileDialog(d)
		})
		if d.result != nil {
			result, ok := selections.(string)
			if err != nil || !ok {
				d.result <- ""
			} else {
				d.result <- result
			}
			close(d.result)
		}
	}
}

// AddFilter adds a filter to the dialog. The filter is a display name and a semicolon separated list of extensions.
// EG: AddFilter("Image Files", "*.jpg;*.png")
func (d *SaveFileDialog) AddFilter(displayName, pattern string) *SaveFileDialog {
	d.filters = append(d.filters, dialogFileFilter{
		DisplayName: strings.TrimSpace(displayName),
		Pattern:     strings.TrimSpace(pattern),
	})
	return d
}

func (d *SaveFileDialog) CanCreateDirectories(canCreateDirectories bool) *SaveFileDialog {
	d.canCreateDirectories = canCreateDirectories
	return d
}

func (d *SaveFileDialog) CanSelectHiddenExtension(canSelectHiddenExtension bool) *SaveFileDialog {
	d.canSelectHiddenExtension = canSelectHiddenExtension
	return d
}

func (d *SaveFileDialog) ShowHiddenFiles(showHiddenFiles bool) *SaveFileDialog {
	d.showHiddenFiles = showHiddenFiles
	return d
}

func (d *SaveFileDialog) SetDirectory(directory string) *SaveFileDialog {
	d.directory = directory
	return d
}

func (d *SaveFileDialog) SetButtonText(text string) *SaveFileDialog {
	d.buttonText = text
	return d
}

func (d *SaveFileDialog) SetFilename(filename string) *SaveFileDialog {
	d.filename = filename
	return d
}

func (d *SaveFileDialog) AllowsOtherFileTypes(allowOtherFileTypes bool) *SaveFileDialog {
	d.allowOtherFileTypes = allowOtherFileTypes
	return d
}

func (d *SaveFileDialog) HideExtension(hideExtension bool) *SaveFileDialog {
	d.hideExtension = hideExtension
	return d
}

func (d *SaveFileDialog) TreatsFilePackagesAsDirectories(treatsFilePackagesAsDirectories bool) *SaveFileDialog {
	d.treatsFilePackagesAsDirectories = treatsFilePackagesAsDirectories
	return d
}

func (d *SaveFileDialog) Show(callbacks ...func(string)) chan string {
	d.result = make(chan string, 1)
	d.app.started.run(d)
	if len(callbacks) > 0 {
		go func() {
			result := <-d.result
			for _, callback := range callbacks {
				callback(result)
			}
		}()
		return nil
	}
	return d.result
}

func runChooserDialog(window windowPtr, allowMultiple, createFolders, showHidden bool, currentFolder, title string, action int, acceptLabel string, filters []dialogFileFilter) ([]string, error) {
	GtkResponseCancel := 0
	GtkResponseAccept := 1

	fc := lib.gtk.FileChooserDialogNew(
		title,
		window,
		action,
		"_Cancel",
		GtkResponseCancel,
		acceptLabel,
		GtkResponseAccept,
		0)

	lib.gtk.FileChooserSetAction(fc, action)

	var gtkFilters []ptr
	for _, filter := range filters {
		f := lib.gtk.FileFilterNew()
		lib.gtk.FileFilterSetName(f, filter.DisplayName)
		lib.gtk.FileFilterAddPattern(f, filter.Pattern)
		lib.gtk.FileChooserAddFilter(fc, f)
		gtkFilters = append(gtkFilters, f)
	}
	lib.gtk.FileChooserSetSelectMultiple(fc, allowMultiple)
	lib.gtk.FileChooserSetCreateFolders(fc, createFolders)
	lib.gtk.FileChooserSetShowHidden(fc, showHidden)

	if currentFolder != "" {
		lib.gtk.FileChooserSetCurrentFolder(fc, currentFolder)
	}

	buildStringAndFree := func(s ptr) string {
		bytes := []byte{}
		p := unsafe.Pointer(s)
		for {
			val := *(*byte)(p)
			if val == 0 { // this is the null terminator
				break
			}
			bytes = append(bytes, val)
			p = unsafe.Add(p, 1)
		}
		lib.g.Free(s) // so we don't have to iterate a second time
		return string(bytes)
	}

	response := lib.gtk.DialogRun(fc)
	var selections []string
	if response == GtkResponseAccept {
		filenames := lib.gtk.FileChooserGetFilenames(fc)
		iter := filenames
		count := 0
		for {
			selections = append(selections, buildStringAndFree(iter.data))
			iter = iter.next
			if iter == nil || count == 1024 {
				break
			}
			count++
		}
	}
	defer lib.gtk.WidgetDestroy(windowPtr(fc))
	return selections, nil
}

// dialog related
func runOpenFileDialog(d *OpenFileDialog) ([]string, error) {
	window := windowPtr(0)
	if d.window != nil {
		window = d.window.pointer
	}
	buttonText := d.buttonText
	if buttonText == "" {
		buttonText = "_Open"
	}
	return runChooserDialog(
		window,
		d.allowsMultipleSelection,
		d.canCreateDirectories,
		d.showHiddenFiles,
		d.directory,
		d.title,
		0, // GtkFileChooserActionOpen
		buttonText,
		d.filters)
}

func runMessageDialog(d *MessageDialog) int {
	window := windowPtr(0)
	if d.window != nil {
		window = d.window.pointer
	}

	buttonMask := GtkButtonsOk
	if len(d.buttons) > 0 {
		buttonMask = GtkButtonsNone
	}
	dialog := lib.gtk.MessageDialogNew(
		window,
		GtkDialogModal|GtkDialogDestroyWithParent,
		d.dtype,
		buttonMask,
		d.message)

	if d.title != "" {
		lib.gtk.WindowSetTitle(dialog, d.title)
	}

	GdkColorspaceRGB := 0
	if img, err := pngToImage(d.icon); err == nil {
		gbytes := lib.g.BytesNewStatic(uintptr(unsafe.Pointer(&img.Pix[0])), len(img.Pix))

		defer lib.g.BytesUnref(gbytes)
		pixBuf := lib.gdk.PixbufNewFromBytes(
			gbytes,
			GdkColorspaceRGB,
			1, // has_alpha
			8,
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			img.Stride,
		)
		image := lib.gtk.ImageNewFromPixbuf(pixBuf)
		widgetSetVisible(image, false)
		contentArea := lib.gtk.DialogGetContentArea(dialog)
		lib.gtk.ContainerAdd(contentArea, image)
	}
	for i, button := range d.buttons {
		lib.gtk.DialogAddButton(
			dialog,
			button.Label,
			i,
		)
		if button.IsDefault {
			lib.gtk.DialogSetDefaultResponse(dialog, i)
		}
	}
	defer lib.gtk.WidgetDestroy(dialog)
	return lib.gtk.DialogRun(dialog)
}

func runSaveFileDialog(d *SaveFileDialog) (string, error) {
	window := windowPtr(0)
	if d.window != nil {
		window = d.window.pointer
	}
	buttonText := d.buttonText
	if buttonText == "" {
		buttonText = "_Save"
	}
	results, err := runChooserDialog(
		window,
		false, // multiple selection
		d.canCreateDirectories,
		d.showHiddenFiles,
		d.directory,
		d.title,
		1, // GtkFileChooserActionSave
		buttonText,
		d.filters)

	if err != nil || len(results) == 0 {
		return "", err
	}
	return results[0], nil
}
